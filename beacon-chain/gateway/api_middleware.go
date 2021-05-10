package gateway

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"reflect"
	"strconv"
	"strings"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/gorilla/mux"
	"github.com/wealdtech/go-bytesutil"
)

type endpointData struct {
	postRequest interface{}
	getResponse interface{}
}

func registerApiMiddleware(gatewayAddress string) {
	r := mux.NewRouter()

	handleApiEndpoint(r, gatewayAddress, "/eth/v1/beacon/genesis", endpointData{getResponse: &GenesisResponseJson{}})
	handleApiEndpoint(r, gatewayAddress, "/eth/v1/beacon/states/{state_id}/root", endpointData{getResponse: &StateRootResponseJson{}})
	handleApiEndpoint(r, gatewayAddress, "/eth/v1/beacon/states/{state_id}/fork", endpointData{getResponse: &StateForkResponseJson{}})
	handleApiEndpoint(r, gatewayAddress, "/eth/v1/beacon/states/{state_id}/finality_checkpoints", endpointData{getResponse: &StateFinalityCheckpointResponseJson{}})
	handleApiEndpoint(r, gatewayAddress, "/eth/v1/beacon/headers/{block_id}", endpointData{getResponse: &BlockHeaderResponseJson{}})
	handleApiEndpoint(r, gatewayAddress, "/eth/v1/beacon/blocks", endpointData{postRequest: &BeaconBlockContainerJson{}})
	handleApiEndpoint(r, gatewayAddress, "/eth/v1/beacon/blocks/{block_id}", endpointData{getResponse: &BlockResponseJson{}})
	handleApiEndpoint(r, gatewayAddress, "/eth/v1/beacon/blocks/{block_id}/root", endpointData{getResponse: &BlockRootResponseJson{}})
	handleApiEndpoint(r, gatewayAddress, "/eth/v1/beacon/blocks/{block_id}/attestations", endpointData{getResponse: &BlockAttestationsResponseJson{}})
	handleApiEndpoint(r, gatewayAddress, "/eth/v1/beacon/pool/attestations", endpointData{postRequest: &SubmitAttestationRequestJson{}, getResponse: &BlockAttestationsResponseJson{}})
	handleApiEndpoint(r, gatewayAddress, "/eth/v1/beacon/pool/attester_slashings", endpointData{postRequest: &AttesterSlashingJson{}, getResponse: &AttesterSlashingsPoolResponseJson{}})
	handleApiEndpoint(r, gatewayAddress, "/eth/v1/beacon/pool/proposer_slashings", endpointData{postRequest: &ProposerSlashingJson{}, getResponse: &ProposerSlashingsPoolResponseJson{}})
	handleApiEndpoint(r, gatewayAddress, "/eth/v1/beacon/pool/voluntary_exits", endpointData{postRequest: &SignedVoluntaryExitJson{}, getResponse: &VoluntaryExitsPoolResponseJson{}})

	if err := http.ListenAndServe(":4500", r); err != nil {
		panic(err)
	}
}

func handleApiEndpoint(r *mux.Router, gatewayAddress string, endpoint string, data endpointData) {
	r.HandleFunc(endpoint, func(writer http.ResponseWriter, request *http.Request) {
		if request.Method == "POST" {
			// Deserialize the body into the 'm' struct, and post it to grpc-gateway.
			if err := json.NewDecoder(request.Body).Decode(&data.postRequest); err != nil {
				panic(err)
			}
			// Posted graffiti needs to have length of 32 bytes.
			if block, ok := data.postRequest.(*BeaconBlockContainerJson); ok {
				b := bytesutil.ToBytes32([]byte(block.Message.Body.Graffiti))
				block.Message.Body.Graffiti = hexutil.Encode(b[:])
			}
			// Encode all fields tagged 'hex' into a base64 string.
			if err := processHexField(data.postRequest, func(v reflect.Value) error {
				b, err := bytesutil.FromHexString(v.String())
				if err != nil {
					return err
				}
				v.SetString(base64.StdEncoding.EncodeToString(b))
				return nil
			}); err != nil {
				log.WithError(err).Error("Could not handle API call")
			}
			// Serialize the struct, which now includes a base64-encoded value, into JSON.
			j, err := json.Marshal(data.postRequest)
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

		// Handle hex in URL parameters.
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

		// Deserialize the output of grpc-gateway into the error struct.
		body, err := ioutil.ReadAll(grpcResp.Body)
		if err != nil {
			panic(err)
		}
		e := ErrorJson{}
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
			if request.Method == "GET" {
				// Deserialize the output of grpc-gateway.
				if err := json.Unmarshal(body, &data.getResponse); err != nil {
					panic(err)
				}
				// Decode all fields tagged 'hex' from a base64 string.
				if err := processHexField(data.getResponse, func(v reflect.Value) error {
					b, err := base64.StdEncoding.DecodeString(v.Interface().(string))
					if err != nil {
						return err
					}
					v.SetString(hexutil.Encode(b))
					return nil
				}); err != nil {
					log.WithError(err).Error("Could not handle API call")
				}
				// Serialize the return value into JSON.
				j, err = json.Marshal(data.getResponse)
				if err != nil {
					panic(err)
				}
			}
		}

		// Write the response (headers + content) and PROFIT!
		for h, vs := range grpcResp.Header {
			for _, v := range vs {
				writer.Header().Set(h, v)
			}
		}
		if e.Message != "" || request.Method == "GET" {
			writer.Header().Set("Content-Length", strconv.Itoa(len(j)))
			writer.WriteHeader(grpcResp.StatusCode)
			if _, err := io.Copy(writer, ioutil.NopCloser(bytes.NewReader(j))); err != nil {
				panic(err)
			}
		} else if request.Method == "POST" {
			writer.WriteHeader(grpcResp.StatusCode)
		}

		if err := grpcResp.Body.Close(); err != nil {
			panic(err)
		}
	})
}

// processHexField calls 'processor' on any field that has the 'hex' tag set.
// It is a recursive function.
func processHexField(s interface{}, processor func(value reflect.Value) error) error {
	t := reflect.TypeOf(s).Elem()
	v := reflect.Indirect(reflect.ValueOf(s))

	for i := 0; i < t.NumField(); i++ {
		switch v.Field(i).Kind() {
		case reflect.Slice:
			sliceElem := t.Field(i).Type.Elem()
			kind := sliceElem.Kind()
			if kind == reflect.Ptr && sliceElem.Elem().Kind() == reflect.Struct {
				for j := 0; j < v.Field(i).Len(); j++ {
					if err := processHexField(v.Field(i).Index(j).Interface(), processor); err != nil {
						return fmt.Errorf("could not process field: %w", err)
					}
				}
			}
			if kind == reflect.String {
				_, isBytes := t.Field(i).Tag.Lookup("hex")
				if isBytes {
					for j := 0; j < v.Field(i).Len(); j++ {
						if err := processor(v.Field(i).Index(j)); err != nil {
							return fmt.Errorf("could not process field: %w", err)
						}
					}
				}
			}
		case reflect.Ptr:
			if v.Field(i).Elem().Kind() == reflect.Struct {
				if err := processHexField(v.Field(i).Interface(), processor); err != nil {
					return fmt.Errorf("could not process field: %w", err)
				}
			}
		default:
			f := t.Field(i)
			_, isBytes := f.Tag.Lookup("hex")
			if isBytes {
				if err := processor(v.Field(i)); err != nil {
					return fmt.Errorf("could not process field: %w", err)
				}
			}
		}
	}
	return nil
}
