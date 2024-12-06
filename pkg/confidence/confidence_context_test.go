package confidence

import (
	"context"
	"log/slog"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

type MockEventUploader struct {
	expectedContext map[string]interface{}
	TestingT        *testing.T
}

// helper function to check if a function panics
func didPanic(f func()) (didPanic bool) {
	defer func() {
		if r := recover(); r != nil {
			didPanic = true
		}
	}()
	f()
	return
}

func (e MockEventUploader) upload(ctx context.Context, request EventBatchRequest) {
	event := request.Events[0]
	assert.True(e.TestingT, reflect.DeepEqual(e.expectedContext, event.Payload["context"]))
}

func TestContextExistsInPayload(t *testing.T) {
	eventUploader := MockEventUploader{
		expectedContext: map[string]interface{}{"hello": "hey"},
	}
	client := createConfidenceWithUploader(t, templateResponse(), eventUploader)
	client.PutContext("hello", "hey")
	wg := client.Track(context.Background(), "test", map[string]interface{}{})
	wg.Wait()
}

func TestContextExistsInDataAndPanic(t *testing.T) {
	eventUploader := MockEventUploader{
		expectedContext: map[string]interface{}{"hello": "hey"},
	}
	client := createConfidenceWithUploader(t, templateResponse(), eventUploader)
	client.PutContext("hello", "hey")
	assert.Panics(t, func() {
		wg := client.Track(context.Background(), "test", map[string]interface{}{"context": "hey"})
		wg.Wait()
	})
}

func TestContextDoesNotExistInDataAndDoesNotPanic(t *testing.T) {
	eventUploader := MockEventUploader{
		expectedContext: map[string]interface{}{"hello": "hey"},
	}
	client := createConfidenceWithUploader(t, templateResponse(), eventUploader)
	client.PutContext("hello", "hey")
	assert.NotPanics(t, func() {
		wg := client.Track(context.Background(), "test", map[string]interface{}{"not_context": "hey"})
		wg.Wait()
	})
}

func TestContextIsInConfidenceObject(t *testing.T) {
	client := create_confidence(t, templateResponse())
	client.PutContext("hello", "hey")
	assert.Equal(t, client.GetContext(), map[string]interface{}{"hello": "hey"})
}

func TestWithContextIsInChildContext(t *testing.T) {
	client := create_confidence(t, templateResponse())
	client.PutContext("hello", "hey")
	child := client.WithContext(map[string]interface{}{"west": "world"})
	assert.Equal(t, child.GetContext(), map[string]interface{}{"hello": "hey", "west": "world"})
	client.PutContext("hello2", "hey2")
	assert.Equal(t, child.GetContext(), map[string]interface{}{"hello": "hey", "west": "world", "hello2": "hey2"})
}

func TestChildContextOverrideParentContext(t *testing.T) {
	client := create_confidence(t, templateResponse())
	client.PutContext("hello", "hey")
	child := client.WithContext(map[string]interface{}{"hello": "boom"})
	assert.Equal(t, child.GetContext(), map[string]interface{}{"hello": "boom"})
	assert.Equal(t, client.GetContext(), map[string]interface{}{"hello": "hey"})
}

func TestChildContextRemoveParentContext(t *testing.T) {
	client := create_confidence(t, templateResponse())
	client.PutContext("hello", "hey")
	child := client.WithContext(map[string]interface{}{})
	child.PutContext("hello", nil)
	assert.Equal(t, child.GetContext(), map[string]interface{}{})
}

func create_confidence(t *testing.T, response ResolveResponse) *Confidence {
	config := APIConfig{
		APIKey: "apiKey",
	}
	return &Confidence{
		Config:        config,
		ResolveClient: MockResolveClient{MockedResponse: response, MockedError: nil, TestingT: t},
		contextMap:    make(map[string]interface{}),
		Logger:        slog.Default(),
	}
}

func createConfidenceWithUploader(t *testing.T, response ResolveResponse, uploader MockEventUploader) *Confidence {
	config := APIConfig{
		APIKey: "apiKey",
	}
	return &Confidence{
		Config:        config,
		EventUploader: uploader,
		ResolveClient: MockResolveClient{MockedResponse: response, MockedError: nil, TestingT: t},
		contextMap:    make(map[string]interface{}),
		Logger:        slog.Default(),
	}
}
