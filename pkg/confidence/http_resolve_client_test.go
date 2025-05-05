package confidence

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/proto"
)

func TestHttpResolveClient_TelemetryHeader(t *testing.T) {
	var receivedHeaders []string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedHeaders = append(receivedHeaders, r.Header.Get("X-CONFIDENCE-TELEMETRY"))
		time.Sleep(10 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(ResolveResponse{})
	}))
	defer server.Close()

	config := APIConfig{
		APIKey:            "test-key",
		APIResolveBaseUrl: server.URL,
		ResolveTimeout:    10 * time.Second,
	}
	client := NewHttpResolveClient(config)

	request := ResolveRequest{
		ClientSecret:      "test-secret",
		EvaluationContext: map[string]interface{}{"targeting_key": "user1"},
		Flags:             []string{"test-flag"},
		Sdk:               sdk{SDK_ID, SDK_VERSION},
	}

	_, err := client.SendResolveRequest(context.Background(), request)
	assert.NoError(t, err)

	_, err = client.SendResolveRequest(context.Background(), request)
	assert.NoError(t, err)

	assert.Equal(t, 2, len(receivedHeaders))

	monitoringBytes, err := base64.StdEncoding.DecodeString(receivedHeaders[0])
	assert.NoError(t, err)

	var firstMonitoring ProtoMonitoring
	err = proto.Unmarshal(monitoringBytes, &firstMonitoring)
	assert.NoError(t, err)
	assert.Equal(t, 0, len(firstMonitoring.LibraryTraces[0].Traces))

	monitoringBytes, err = base64.StdEncoding.DecodeString(receivedHeaders[1])
	assert.NoError(t, err)

	var secondMonitoring ProtoMonitoring
	err = proto.Unmarshal(monitoringBytes, &secondMonitoring)
	assert.NoError(t, err)

	assert.Equal(t, ProtoPlatform_PROTO_PLATFORM_GO, secondMonitoring.Platform)
	assert.Equal(t, 1, len(secondMonitoring.LibraryTraces))
	assert.Equal(t, ProtoLibraryTraces_PROTO_LIBRARY_CONFIDENCE, secondMonitoring.LibraryTraces[0].Library)
	assert.Equal(t, SDK_VERSION, secondMonitoring.LibraryTraces[0].LibraryVersion)
	assert.Equal(t, 1, len(secondMonitoring.LibraryTraces[0].Traces))
	assert.Equal(t, ProtoLibraryTraces_PROTO_TRACE_ID_RESOLVE_LATENCY, secondMonitoring.LibraryTraces[0].Traces[0].Id)
	assert.NotNil(t, secondMonitoring.LibraryTraces[0].Traces[0].GetRequestTrace())
	assert.Equal(t, ProtoLibraryTraces_ProtoTrace_ProtoRequestTrace_PROTO_STATUS_SUCCESS, secondMonitoring.LibraryTraces[0].Traces[0].GetRequestTrace().Status)
	assert.GreaterOrEqual(t, secondMonitoring.LibraryTraces[0].Traces[0].GetRequestTrace().MillisecondDuration, uint64(10))
}

func TestHttpResolveClient_TelemetryHeader_LatenciesAreCleared(t *testing.T) {
	var receivedHeaders []string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedHeaders = append(receivedHeaders, r.Header.Get("X-CONFIDENCE-TELEMETRY"))
		time.Sleep(10 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(ResolveResponse{})
	}))
	defer server.Close()

	config := APIConfig{
		APIKey:            "test-key",
		APIResolveBaseUrl: server.URL,
		ResolveTimeout:    10 * time.Second,
	}
	client := NewHttpResolveClient(config)

	request := ResolveRequest{
		ClientSecret:      "test-secret",
		EvaluationContext: map[string]interface{}{"targeting_key": "user1"},
		Flags:             []string{"test-flag"},
		Sdk:               sdk{SDK_ID, SDK_VERSION},
	}

	_, err := client.SendResolveRequest(context.Background(), request)
	assert.NoError(t, err)

	_, err = client.SendResolveRequest(context.Background(), request)
	assert.NoError(t, err)

	_, err = client.SendResolveRequest(context.Background(), request)
	assert.NoError(t, err)

	assert.Equal(t, 3, len(receivedHeaders))

	monitoringBytes, err := base64.StdEncoding.DecodeString(receivedHeaders[2])
	assert.NoError(t, err)

	var thirdMonitoring ProtoMonitoring
	err = proto.Unmarshal(monitoringBytes, &thirdMonitoring)
	assert.NoError(t, err)

	assert.Equal(t, 1, len(thirdMonitoring.LibraryTraces[0].Traces))
}

func TestHttpResolveClient_TelemetryHeader_ErrorStatus(t *testing.T) {
	var receivedHeaders []string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedHeaders = append(receivedHeaders, r.Header.Get("X-CONFIDENCE-TELEMETRY"))
		time.Sleep(10 * time.Millisecond)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(ResolveResponse{})
	}))
	defer server.Close()

	config := APIConfig{
		APIKey:            "test-key",
		APIResolveBaseUrl: server.URL,
		ResolveTimeout:    10 * time.Second,
	}
	client := NewHttpResolveClient(config)

	request := ResolveRequest{
		ClientSecret:      "test-secret",
		EvaluationContext: map[string]interface{}{"targeting_key": "user1"},
		Flags:             []string{"test-flag"},
		Sdk:               sdk{SDK_ID, SDK_VERSION},
	}

	_, err := client.SendResolveRequest(context.Background(), request)
	assert.Error(t, err)

	_, err = client.SendResolveRequest(context.Background(), request)
	assert.Error(t, err)

	assert.Equal(t, 2, len(receivedHeaders))

	monitoringBytes, err := base64.StdEncoding.DecodeString(receivedHeaders[0])
	assert.NoError(t, err)

	var firstMonitoring ProtoMonitoring
	err = proto.Unmarshal(monitoringBytes, &firstMonitoring)
	assert.NoError(t, err)
	assert.Equal(t, 0, len(firstMonitoring.LibraryTraces[0].Traces))

	monitoringBytes, err = base64.StdEncoding.DecodeString(receivedHeaders[1])
	assert.NoError(t, err)

	var secondMonitoring ProtoMonitoring
	err = proto.Unmarshal(monitoringBytes, &secondMonitoring)
	assert.NoError(t, err)

	assert.Equal(t, ProtoPlatform_PROTO_PLATFORM_GO, secondMonitoring.Platform)
	assert.Equal(t, 1, len(secondMonitoring.LibraryTraces))
	assert.Equal(t, ProtoLibraryTraces_PROTO_LIBRARY_CONFIDENCE, secondMonitoring.LibraryTraces[0].Library)
	assert.Equal(t, SDK_VERSION, secondMonitoring.LibraryTraces[0].LibraryVersion)
	assert.Equal(t, 1, len(secondMonitoring.LibraryTraces[0].Traces))
	assert.Equal(t, ProtoLibraryTraces_PROTO_TRACE_ID_RESOLVE_LATENCY, secondMonitoring.LibraryTraces[0].Traces[0].Id)
	assert.NotNil(t, secondMonitoring.LibraryTraces[0].Traces[0].GetRequestTrace())
	assert.Equal(t, ProtoLibraryTraces_ProtoTrace_ProtoRequestTrace_PROTO_STATUS_ERROR, secondMonitoring.LibraryTraces[0].Traces[0].GetRequestTrace().Status)
	assert.GreaterOrEqual(t, secondMonitoring.LibraryTraces[0].Traces[0].GetRequestTrace().MillisecondDuration, uint64(10))
}

func TestHttpResolveClient_TelemetryHeader_NetworkError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hj, ok := w.(http.Hijacker)
		if !ok {
			t.Fatal("webserver doesn't support hijacking")
		}
		conn, _, err := hj.Hijack()
		if err != nil {
			t.Fatal(err)
		}
		conn.Close()
	}))
	defer server.Close()

	config := APIConfig{
		APIKey:            "test-key",
		APIResolveBaseUrl: server.URL,
		ResolveTimeout:    10 * time.Second,
	}
	client := NewHttpResolveClient(config)

	request := ResolveRequest{
		ClientSecret:      "test-secret",
		EvaluationContext: map[string]interface{}{"targeting_key": "user1"},
		Flags:             []string{"test-flag"},
		Sdk:               sdk{SDK_ID, SDK_VERSION},
	}

	_, err := client.SendResolveRequest(context.Background(), request)
	assert.Error(t, err)

	traces := client.GetTracesAndClear()
	assert.Equal(t, 1, len(traces))
	assert.Equal(t, ProtoLibraryTraces_ProtoTrace_ProtoRequestTrace_PROTO_STATUS_ERROR, traces[0].GetRequestTrace().Status)
}

func TestHttpResolveClient_TelemetryHeader_MixedStatuses(t *testing.T) {
	var receivedHeaders []string
	successCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedHeaders = append(receivedHeaders, r.Header.Get("X-CONFIDENCE-TELEMETRY"))
		time.Sleep(10 * time.Millisecond)

		if successCount%2 == 0 {
			w.WriteHeader(http.StatusOK)
		} else {
			w.WriteHeader(http.StatusInternalServerError)
		}
		successCount++
		json.NewEncoder(w).Encode(ResolveResponse{})
	}))
	defer server.Close()

	config := APIConfig{
		APIKey:            "test-key",
		APIResolveBaseUrl: server.URL,
		ResolveTimeout:    10 * time.Second,
	}
	client := NewHttpResolveClient(config)

	request := ResolveRequest{
		ClientSecret:      "test-secret",
		EvaluationContext: map[string]interface{}{"targeting_key": "user1"},
		Flags:             []string{"test-flag"},
		Sdk:               sdk{SDK_ID, SDK_VERSION},
	}

	for i := 0; i < 4; i++ {
		_, err := client.SendResolveRequest(context.Background(), request)
		if i%2 == 0 {
			assert.NoError(t, err)
		} else {
			assert.Error(t, err)
		}

		traces := client.GetTracesAndClear()
		assert.Equal(t, 1, len(traces), "Expected 1 trace after request %d", i)
		if i%2 == 0 {
			assert.Equal(t, ProtoLibraryTraces_ProtoTrace_ProtoRequestTrace_PROTO_STATUS_SUCCESS, traces[0].GetRequestTrace().Status, "Expected success status at index %d", i)
		} else {
			assert.Equal(t, ProtoLibraryTraces_ProtoTrace_ProtoRequestTrace_PROTO_STATUS_ERROR, traces[0].GetRequestTrace().Status, "Expected error status at index %d", i)
		}
	}
}

func TestHttpResolveClient_TelemetryHeader_DeserializationError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(10 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("{"))
	}))
	defer server.Close()

	config := APIConfig{
		APIKey:            "test-key",
		APIResolveBaseUrl: server.URL,
		ResolveTimeout:    10 * time.Second,
	}
	client := NewHttpResolveClient(config)

	request := ResolveRequest{
		ClientSecret:      "test-secret",
		EvaluationContext: map[string]interface{}{"targeting_key": "user1"},
		Flags:             []string{"test-flag"},
		Sdk:               sdk{SDK_ID, SDK_VERSION},
	}

	_, err := client.SendResolveRequest(context.Background(), request)
	assert.Error(t, err)

	traces := client.GetTracesAndClear()
	assert.Equal(t, 1, len(traces))
	assert.Equal(t, ProtoLibraryTraces_ProtoTrace_ProtoRequestTrace_PROTO_STATUS_ERROR, traces[0].GetRequestTrace().Status)
}

func TestHttpResolveClient_TelemetryHeader_TimeoutError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Simulate a timeout by sleeping longer than the client's timeout
		time.Sleep(100 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(ResolveResponse{})
	}))
	defer server.Close()

	config := APIConfig{
		APIKey:            "test-key",
		APIResolveBaseUrl: server.URL,
		ResolveTimeout:    1 * time.Millisecond, // Set a very short timeout to force timeout
	}
	client := NewHttpResolveClient(config)

	request := ResolveRequest{
		ClientSecret:      "test-secret",
		EvaluationContext: map[string]interface{}{"targeting_key": "user1"},
		Flags:             []string{"test-flag"},
		Sdk:               sdk{SDK_ID, SDK_VERSION},
	}

	_, err := client.SendResolveRequest(context.Background(), request)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Client.Timeout exceeded while awaiting headers")

	traces := client.GetTracesAndClear()
	assert.Equal(t, 1, len(traces))
	assert.Equal(t, ProtoLibraryTraces_PROTO_TRACE_ID_RESOLVE_LATENCY, traces[0].Id)
	assert.NotNil(t, traces[0].GetRequestTrace())
	assert.Equal(t, ProtoLibraryTraces_ProtoTrace_ProtoRequestTrace_PROTO_STATUS_TIMEOUT, traces[0].GetRequestTrace().Status)
	assert.GreaterOrEqual(t, traces[0].GetRequestTrace().MillisecondDuration, uint64(1))
}
