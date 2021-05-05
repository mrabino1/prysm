package gateway

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
	"reflect"
	"strconv"
	"strings"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/gorilla/mux"
)

func registerApiMiddleware(gatewayAddress string) {
	r := mux.NewRouter()
	handleApiEndpoint(r, gatewayAddress, "/eth/v1/beacon/states/{state_id}/fork", &StateForkResponseJson{})
	// TODO: make configurable?
	if err := http.ListenAndServe(":4500", r); err != nil {
		panic(err)
	}
}

func handleApiEndpoint(r *mux.Router, gatewayAddress string, endpoint string, m interface{}) {
	r.HandleFunc(endpoint, func(writer http.ResponseWriter, request *http.Request) {
		// Structs for body deserialization.
		e := ErrorJson{}

		if request.Method == "POST" {
			// Deserialize the body into the 'm' struct, and post it to grpc-gateway.
			if err := json.NewDecoder(request.Body).Decode(&m); err != nil {
				panic(err)
			}
			// Encode all fields tagged 'bytes' into a base64 string.
			processHexField(m, func(v reflect.Value) {
				v.SetString(base64.StdEncoding.EncodeToString([]byte(v.String())))
			})
			// Serialize the struct, which now includes a base64-encoded value, into JSON.
			j, err := json.Marshal(m)
			if err != nil {
				panic(err)
			}
			// Set the body to the new JSON.
			request.Body = ioutil.NopCloser(bytes.NewReader(j))
			request.Header.Set("Content-Length", strconv.Itoa(len(j)))
			request.ContentLength = int64(len(j))
		}
		request.URL.Scheme = "http"
		request.URL.Host = gatewayAddress
		request.RequestURI = ""

		splitEndpoint := strings.Split(endpoint, "/")
		for i, s := range splitEndpoint {
			if s == "{state_id}" || s == "{block_id}" {
				routeVar := base64.StdEncoding.EncodeToString([]byte(mux.Vars(request)[s[1:len(s)-1]]))
				splitPath := strings.Split(request.URL.Path, "/")
				splitPath[i] = routeVar
				request.URL.Path = strings.Join(splitPath, "/")
				break
			}
		}

		grpcResp, err := http.DefaultClient.Do(request)
		if err != nil {
			panic(err)
		}
		if grpcResp == nil {
			panic("nil response from grpc-gateway")
		}
		// Deserialize the output of grpc-gateway's server into the 'e' struct.
		body, err := ioutil.ReadAll(grpcResp.Body)
		if err != nil {
			panic(err)
		}
		if err := json.Unmarshal(body, &e); err != nil {
			panic(err)
		}
		var j []byte
		// The output might have contained a 'Not Found' error,
		// in which case the 'Message' field will be populated.
		if e.Message != "" {
			// Serialize the error message into JSON.
			j, err = json.Marshal(e)
			if err != nil {
				panic(err)
			}
		} else {
			// Deserialize the output of grpc-gateway's server into the 'm' struct.
			if err := json.Unmarshal(body, &m); err != nil {
				panic(err)
			}
			// Decode all fields tagged 'bytes' from a base64 string.
			processHexField(m, func(v reflect.Value) {
				b, err := base64.StdEncoding.DecodeString(v.Interface().(string))
				if err != nil {
					panic(err)
				}
				v.SetString(hexutil.Encode(b))
			})
			// Serialize the return value into JSON.
			j, err = json.Marshal(m)
			if err != nil {
				panic(err)
			}
		}

		// Write the response (headers + content) and PROFIT!
		for h, vs := range grpcResp.Header {
			for _, v := range vs {
				writer.Header().Set(h, v)
			}
		}
		writer.Header().Set("Content-Length", strconv.Itoa(len(j)))
		writer.WriteHeader(grpcResp.StatusCode)
		if _, err := io.Copy(writer, ioutil.NopCloser(bytes.NewReader(j))); err != nil {
			panic(err)
		}

		if err := grpcResp.Body.Close(); err != nil {
			panic(err)
		}
	})
}

// processHexField calls 'processor' on any field that has the 'bytes' tag set.
// It is a recursive function.
func processHexField(s interface{}, processor func(value reflect.Value)) {
	t := reflect.TypeOf(s).Elem()
	v := reflect.Indirect(reflect.ValueOf(s))

	for i := 0; i < t.NumField(); i++ {
		switch v.Field(i).Kind() {
		case reflect.Slice:
			kind := reflect.TypeOf(v.Field(i)).Kind()
			if kind == reflect.Struct {
				for j := 0; j < v.Field(i).Len(); j++ {
					processHexField(v.Field(i).Index(j).Interface(), processor)
				}
			}
		case reflect.Ptr:
			if v.Field(i).Elem().Kind() == reflect.Struct {
				processHexField(v.Field(i).Interface(), processor)
			}
		default:
			f := t.Field(i)
			_, isBytes := f.Tag.Lookup("hex")
			if isBytes {
				processor(v.Field(i))
			}
		}
	}
}
