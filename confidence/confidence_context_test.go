package confidence

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

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

func create_confidence(t *testing.T, response ResolveResponse) *Confidence {
	config := APIConfig{
		APIKey: "apiKey",
		Region: APIRegionGlobal,
	}
	return &Confidence{
		Config:        config,
		ResolveClient: MockResolveClient{MockedResponse: response, MockedError: nil, TestingT: t},
		contextMap:    make(map[string]interface{}),
	}
}
