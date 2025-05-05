package confidence

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"google.golang.org/protobuf/proto"
)

type HttpResolveClient struct {
	Client *http.Client
	Config APIConfig
	traces chan *ProtoLibraryTraces_ProtoTrace
}

func NewHttpResolveClient(config APIConfig) *HttpResolveClient {
	return &HttpResolveClient{
		Client: &http.Client{
			Timeout: config.ResolveTimeout,
		},
		Config: config,
		traces: make(chan *ProtoLibraryTraces_ProtoTrace, 1000), // Buffer size of 1000 should be sufficient
	}
}

func (client *HttpResolveClient) GetTracesAndClear() []*ProtoLibraryTraces_ProtoTrace {
	traces := make([]*ProtoLibraryTraces_ProtoTrace, 0)
	for {
		select {
		case trace := <-client.traces:
			traces = append(traces, trace)
		default:
			return traces
		}
	}
}

func parseErrorMessage(body io.ReadCloser) string {
	var resolveError resolveErrorMessage
	decoder := json.NewDecoder(body)
	decoder.UseNumber()
	err := decoder.Decode(&resolveError)
	if err != nil {
		return ""
	}
	return resolveError.Message
}

func (client *HttpResolveClient) SendResolveRequest(ctx context.Context,
	request ResolveRequest) (ResolveResponse, error) {
	jsonRequest, err := json.Marshal(request)
	if err != nil {
		return ResolveResponse{}, fmt.Errorf("error when serializing request to the resolver service: %w", err)
	}

	payload := bytes.NewBuffer(jsonRequest)
	req, err := http.NewRequestWithContext(ctx,
		http.MethodPost, fmt.Sprintf("%s/v1/flags:resolve", client.Config.APIResolveBaseUrl), payload)
	if err != nil {
		return ResolveResponse{}, err
	}

	traces := client.GetTracesAndClear()

	monitoring := &ProtoMonitoring{
		Platform: ProtoPlatform_PROTO_PLATFORM_GO,
		LibraryTraces: []*ProtoLibraryTraces{
			{
				Library:        ProtoLibraryTraces_PROTO_LIBRARY_CONFIDENCE,
				LibraryVersion: SDK_VERSION,
				Traces:         traces,
			},
		},
	}

	monitoringBytes, err := proto.Marshal(monitoring)
	if err == nil {
		monitoringBase64 := base64.StdEncoding.EncodeToString(monitoringBytes)
		req.Header.Set("X-CONFIDENCE-TELEMETRY", monitoringBase64)
	}

	startTime := time.Now()
	resp, err := client.Client.Do(req)
	if err != nil {
		status := ProtoLibraryTraces_ProtoTrace_ProtoRequestTrace_PROTO_STATUS_ERROR
		if err, ok := err.(interface{ Timeout() bool }); ok && err.Timeout() {
			status = ProtoLibraryTraces_ProtoTrace_ProtoRequestTrace_PROTO_STATUS_TIMEOUT
		}
		select {
		case client.traces <- &ProtoLibraryTraces_ProtoTrace{
			Id: ProtoLibraryTraces_PROTO_TRACE_ID_RESOLVE_LATENCY,
			Trace: &ProtoLibraryTraces_ProtoTrace_RequestTrace{
				RequestTrace: &ProtoLibraryTraces_ProtoTrace_ProtoRequestTrace{
					MillisecondDuration: uint64(time.Since(startTime).Milliseconds()),
					Status:              status,
				},
			},
		}:
		default:
			// Channel is full, drop the trace
		}
		return ResolveResponse{}, fmt.Errorf("error when calling the resolver service: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		select {
		case client.traces <- &ProtoLibraryTraces_ProtoTrace{
			Id: ProtoLibraryTraces_PROTO_TRACE_ID_RESOLVE_LATENCY,
			Trace: &ProtoLibraryTraces_ProtoTrace_RequestTrace{
				RequestTrace: &ProtoLibraryTraces_ProtoTrace_ProtoRequestTrace{
					MillisecondDuration: uint64(time.Since(startTime).Milliseconds()),
					Status:              ProtoLibraryTraces_ProtoTrace_ProtoRequestTrace_PROTO_STATUS_ERROR,
				},
			},
		}:
		default:
			// Channel is full, drop the trace
		}
		return ResolveResponse{},
			fmt.Errorf("got '%s' error from the resolver service: %s", resp.Status, parseErrorMessage(resp.Body))
	}

	var result ResolveResponse
	decoder := json.NewDecoder(resp.Body)
	decoder.UseNumber()
	err = decoder.Decode(&result)
	if err != nil {
		select {
		case client.traces <- &ProtoLibraryTraces_ProtoTrace{
			Id: ProtoLibraryTraces_PROTO_TRACE_ID_RESOLVE_LATENCY,
			Trace: &ProtoLibraryTraces_ProtoTrace_RequestTrace{
				RequestTrace: &ProtoLibraryTraces_ProtoTrace_ProtoRequestTrace{
					MillisecondDuration: uint64(time.Since(startTime).Milliseconds()),
					Status:              ProtoLibraryTraces_ProtoTrace_ProtoRequestTrace_PROTO_STATUS_ERROR,
				},
			},
		}:
		default:
			// Channel is full, drop the trace
		}
		return ResolveResponse{}, fmt.Errorf("error when deserializing response from the resolver service: %w", err)
	}

	select {
	case client.traces <- &ProtoLibraryTraces_ProtoTrace{
		Id: ProtoLibraryTraces_PROTO_TRACE_ID_RESOLVE_LATENCY,
		Trace: &ProtoLibraryTraces_ProtoTrace_RequestTrace{
			RequestTrace: &ProtoLibraryTraces_ProtoTrace_ProtoRequestTrace{
				MillisecondDuration: uint64(time.Since(startTime).Milliseconds()),
				Status:              ProtoLibraryTraces_ProtoTrace_ProtoRequestTrace_PROTO_STATUS_SUCCESS,
			},
		},
	}:
	default:
		// Channel is full, drop the trace
	}

	return result, nil
}
