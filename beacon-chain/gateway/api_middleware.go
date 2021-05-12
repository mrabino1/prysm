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

// ApiProxyMiddleware is a proxy between an Eth2 API HTTP client and gRPC-gateway.
// The purpose of the proxy is to handle HTTP requests and gRPC responses in such a way that:
//   - Eth2 API requests can be handled by gRPC-gateway correctly
//   - gRPC responses can be returned as spec-compliant Eth2 API responses
type ApiProxyMiddleware struct {
	GatewayAddress string
	ProxyAddress   string
	router         *mux.Router
}

type endpointData struct {
	postRequest interface{}
	getResponse interface{}
}

// Run starts the proxy, registering all proxy endpoints on ApiProxyMiddleware.ProxyAddress.
func (m *ApiProxyMiddleware) Run() error {
	m.router = mux.NewRouter()

	m.handleApiEndpoint("/eth/v1/beacon/genesis", endpointData{getResponse: &GenesisResponseJson{}})
	m.handleApiEndpoint("/eth/v1/beacon/states/{state_id}/root", endpointData{getResponse: &StateRootResponseJson{}})
	m.handleApiEndpoint("/eth/v1/beacon/states/{state_id}/fork", endpointData{getResponse: &StateForkResponseJson{}})
	m.handleApiEndpoint("/eth/v1/beacon/states/{state_id}/finality_checkpoints", endpointData{getResponse: &StateFinalityCheckpointResponseJson{}})
	m.handleApiEndpoint("/eth/v1/beacon/headers/{block_id}", endpointData{getResponse: &BlockHeaderResponseJson{}})
	m.handleApiEndpoint("/eth/v1/beacon/blocks", endpointData{postRequest: &BeaconBlockContainerJson{}})
	m.handleApiEndpoint("/eth/v1/beacon/blocks/{block_id}", endpointData{getResponse: &BlockResponseJson{}})
	m.handleApiEndpoint("/eth/v1/beacon/blocks/{block_id}/root", endpointData{getResponse: &BlockRootResponseJson{}})
	m.handleApiEndpoint("/eth/v1/beacon/blocks/{block_id}/attestations", endpointData{getResponse: &BlockAttestationsResponseJson{}})
	m.handleApiEndpoint("/eth/v1/beacon/pool/attestations", endpointData{postRequest: &SubmitAttestationRequestJson{}, getResponse: &BlockAttestationsResponseJson{}})
	m.handleApiEndpoint("/eth/v1/beacon/pool/attester_slashings", endpointData{postRequest: &AttesterSlashingJson{}, getResponse: &AttesterSlashingsPoolResponseJson{}})
	m.handleApiEndpoint("/eth/v1/beacon/pool/proposer_slashings", endpointData{postRequest: &ProposerSlashingJson{}, getResponse: &ProposerSlashingsPoolResponseJson{}})
	m.handleApiEndpoint("/eth/v1/beacon/pool/voluntary_exits", endpointData{postRequest: &SignedVoluntaryExitJson{}, getResponse: &VoluntaryExitsPoolResponseJson{}})
	m.handleApiEndpoint("/eth/v1/node/identity", endpointData{getResponse: &IdentityResponseJson{}})
	m.handleApiEndpoint("/eth/v1/node/peers", endpointData{getResponse: &PeersResponseJson{}})
	m.handleApiEndpoint("/eth/v1/node/peers/{peer_id}", endpointData{getResponse: &PeerResponseJson{}})
	m.handleApiEndpoint("/eth/v1/node/peer_count", endpointData{getResponse: &PeerCountResponseJson{}})
	m.handleApiEndpoint("/eth/v1/node/version", endpointData{getResponse: &VersionResponseJson{}})
	m.handleApiEndpoint("/eth/v1/node/health", endpointData{})
	m.handleApiEndpoint("/eth/v1/debug/beacon/states/{state_id}", endpointData{getResponse: &BeaconStateResponseJson{}})
	m.handleApiEndpoint("/eth/v1/debug/beacon/heads", endpointData{getResponse: &ForkChoiceHeadsResponseJson{}})
	m.handleApiEndpoint("/eth/v1/config/fork_schedule", endpointData{getResponse: &ForkScheduleResponseJson{}})

	return http.ListenAndServe(m.ProxyAddress, m.router)
}

func (m *ApiProxyMiddleware) handleApiEndpoint(endpoint string, data endpointData) {
	m.router.HandleFunc(endpoint, func(writer http.ResponseWriter, request *http.Request) {
		if request.Method == "POST" {
			// Deserialize the body into the 'm' struct, and post it to grpc-gateway.
			if err := json.NewDecoder(request.Body).Decode(&data.postRequest); err != nil {
				e := fmt.Errorf("could not decode request body: %w", err)
				writeError(writer, ErrorJson{Message: e.Error()})
				return
			}
			prepareGraffiti(data)
			// Encode all fields tagged 'hex' into a base64 string.
			if err := processHex(data.postRequest, func(v reflect.Value) error {
				b, err := bytesutil.FromHexString(v.String())
				if err != nil {
					return err
				}
				v.SetString(base64.StdEncoding.EncodeToString(b))
				return nil
			}); err != nil {
				e := fmt.Errorf("could not process request hex data: %w", err)
				writeError(writer, ErrorJson{Message: e.Error()})
				return
			}
			// Serialize the struct, which now includes a base64-encoded value, into JSON.
			j, err := json.Marshal(data.postRequest)
			if err != nil {
				e := fmt.Errorf("could not marshal request: %w", err)
				writeError(writer, ErrorJson{Message: e.Error()})
				return
			}
			// Set the body to the new JSON.
			request.Body = ioutil.NopCloser(bytes.NewReader(j))
			request.Header.Set("Content-Length", strconv.Itoa(len(j)))
			request.ContentLength = int64(len(j))
		}

		request.URL.Scheme = "http"
		request.URL.Host = m.GatewayAddress
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
			e := fmt.Errorf("could not proxy request: %w", err)
			writeError(writer, ErrorJson{Message: e.Error()})
			return
		}
		if grpcResp == nil {
			writeError(writer, ErrorJson{Message: "nil response from gRPC-gateway"})
			return
		}

		// Deserialize the output of grpc-gateway into the error struct.
		body, err := ioutil.ReadAll(grpcResp.Body)
		if err != nil {
			e := fmt.Errorf("could not read response body: %w", err)
			writeError(writer, ErrorJson{Message: e.Error()})
			return
		}
		errorJson := ErrorJson{}
		if err := json.Unmarshal(body, &errorJson); err != nil {
			e := fmt.Errorf("could not unmarshal error: %w", err)
			writeError(writer, ErrorJson{Message: e.Error()})
			return
		}
		var j []byte
		if errorJson.Message != "" {
			// Something went wrong, but the request completed, meaning we can write headers and the error message.
			for h, vs := range grpcResp.Header {
				for _, v := range vs {
					writer.Header().Set(h, v)
				}
			}
			writeError(writer, errorJson)
			return
		} else {
			// Don't do anything if the response is only a status code.
			if request.Method == "GET" && data.getResponse != nil {
				// Deserialize the output of grpc-gateway.
				if err := json.Unmarshal(body, &data.getResponse); err != nil {
					e := fmt.Errorf("could not unmarshal response: %w", err)
					writeError(writer, ErrorJson{Message: e.Error()})
					return
				}
				// Decode all fields tagged 'hex' from a base64 string.
				if err := processHex(data.getResponse, func(v reflect.Value) error {
					b, err := base64.StdEncoding.DecodeString(v.Interface().(string))
					if err != nil {
						return err
					}
					v.SetString(hexutil.Encode(b))
					return nil
				}); err != nil {
					e := fmt.Errorf("could not process response hex data: %w", err)
					writeError(writer, ErrorJson{Message: e.Error()})
					return
				}
				// Serialize the return value into JSON.
				j, err = json.Marshal(data.getResponse)
				if err != nil {
					e := fmt.Errorf("could not marshal response: %w", err)
					writeError(writer, ErrorJson{Message: e.Error()})
					return
				}
			}
		}

		// Write the response (headers + content) and PROFIT!
		for h, vs := range grpcResp.Header {
			for _, v := range vs {
				writer.Header().Set(h, v)
			}
		}
		if request.Method == "GET" {
			writer.Header().Set("Content-Length", strconv.Itoa(len(j)))
			writer.WriteHeader(grpcResp.StatusCode)
			if _, err := io.Copy(writer, ioutil.NopCloser(bytes.NewReader(j))); err != nil {
				e := fmt.Errorf("could not write response message: %w", err)
				writeError(writer, ErrorJson{Message: e.Error()})
				return
			}
		} else if request.Method == "POST" {
			writer.WriteHeader(grpcResp.StatusCode)
		}

		if err := grpcResp.Body.Close(); err != nil {
			e := fmt.Errorf("could not close response body: %w", err)
			writeError(writer, ErrorJson{Message: e.Error()})
			return
		}
	})
}

func prepareGraffiti(data endpointData) {
	// Posted graffiti needs to have length of 32 bytes.
	if block, ok := data.postRequest.(*BeaconBlockContainerJson); ok {
		b := bytesutil.ToBytes32([]byte(block.Message.Body.Graffiti))
		block.Message.Body.Graffiti = hexutil.Encode(b[:])
	}
}

func writeError(writer http.ResponseWriter, e ErrorJson) {
	j, err := json.Marshal(e)
	if err != nil {
		log.WithError(err).Error("could not marshal error message")
	}
	writer.Header().Set("Content-Length", strconv.Itoa(len(j)))
	writer.Header().Set("Content-Type", "application/json")
	writer.WriteHeader(http.StatusInternalServerError)
	if _, err := io.Copy(writer, ioutil.NopCloser(bytes.NewReader(j))); err != nil {
		log.WithError(err).Error("could not write error message")
	}
}

// processHex calls 'processor' on any field that has the 'hex' tag set.
// It is a recursive function.
func processHex(s interface{}, processor func(value reflect.Value) error) error {
	t := reflect.TypeOf(s).Elem()
	v := reflect.Indirect(reflect.ValueOf(s))

	for i := 0; i < t.NumField(); i++ {
		switch v.Field(i).Kind() {
		case reflect.Slice:
			sliceElem := t.Field(i).Type.Elem()
			kind := sliceElem.Kind()
			// Recursively process slices to struct pointers.
			if kind == reflect.Ptr && sliceElem.Elem().Kind() == reflect.Struct {
				for j := 0; j < v.Field(i).Len(); j++ {
					if err := processHex(v.Field(i).Index(j).Interface(), processor); err != nil {
						return fmt.Errorf("could not process field: %w", err)
					}
				}
			}
			// Process each string in string slices.
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
		// Recursively process struct pointers.
		case reflect.Ptr:
			if v.Field(i).Elem().Kind() == reflect.Struct {
				if err := processHex(v.Field(i).Interface(), processor); err != nil {
					return fmt.Errorf("could not process field: %w", err)
				}
			}
		default:
			f := t.Field(i)
			// Process fields with 'hex' tag.
			if _, isBytes := f.Tag.Lookup("hex"); isBytes {
				if err := processor(v.Field(i)); err != nil {
					return fmt.Errorf("could not process field: %w", err)
				}
			}
		}
	}
	return nil
}
