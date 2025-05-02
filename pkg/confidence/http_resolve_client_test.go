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
		time.Sleep(10 * time.Millisecond) // Simulate network delay
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

	// Second request will contain telemetry with the data of the first request
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
		time.Sleep(10 * time.Millisecond) // Simulate network delay
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

	// Verify we received three headers
	assert.Equal(t, 3, len(receivedHeaders))

	monitoringBytes, err := base64.StdEncoding.DecodeString(receivedHeaders[2])
	assert.NoError(t, err)

	var thirdMonitoring ProtoMonitoring
	err = proto.Unmarshal(monitoringBytes, &thirdMonitoring)
	assert.NoError(t, err)

	assert.Equal(t, 1, len(thirdMonitoring.LibraryTraces[0].Traces))
}
