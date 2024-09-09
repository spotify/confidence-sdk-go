package confidence

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/stretchr/testify/assert"
	"reflect"
	"testing"
)

type MockResolveClient struct {
	MockedResponse ResolveResponse
	MockedError    error
	TestingT       *testing.T
}

func (r MockResolveClient) SendResolveRequest(_ context.Context,
	request ResolveRequest) (ResolveResponse, error) {
	assert.Equal(r.TestingT, "user1", request.EvaluationContext["targeting_key"])
	return r.MockedResponse, r.MockedError
}

func TestResolveBoolValue(t *testing.T) {
	client := client(t, templateResponse(), nil)
	client.PutContext("targeting_key", "user1")
	evalDetails := client.GetBoolFlag(context.Background(), "test-flag.boolean-key", false)

	assert.Equal(t, true, evalDetails.Value)
	assert.Equal(t, TargetingMatchReason, evalDetails.Reason)
	assert.Equal(t, "flags/test-flag/variants/treatment", evalDetails.Variant)
}

func TestResolveIntValue(t *testing.T) {
	client := client(t, templateResponse(), nil)
	client.PutContext("targeting_key", "user1")

	evalDetails := client.GetIntFlag(context.Background(), "test-flag.integer-key", 99)

	assert.Equal(t, int64(40), evalDetails.Value)
}

func TestResolveDoubleValue(t *testing.T) {
	client := client(t, templateResponse(), nil)
	client.PutContext("targeting_key", "user1")

	evalDetails := client.GetDoubleFlag(context.Background(), "test-flag.double-key", 99.99)

	assert.Equal(t, 20.203, evalDetails.Value)
}

func TestResolveStringValue(t *testing.T) {
	client := client(t, templateResponse(), nil)
	client.PutContext("targeting_key", "user1")

	evalDetails := client.GetStringFlag(context.Background(), "test-flag.string-key", "default")

	assert.Equal(t, "treatment", evalDetails.Value)
}

func TestResolveObjectValue(t *testing.T) {
	client := client(t, templateResponse(), nil)
	client.PutContext("targeting_key", "user1")

	evalDetails := client.GetObjectFlag(context.Background(), "test-flag.struct-key", map[string]interface{}{})
	_, ok := evalDetails.Value.(map[string]interface{})
	assert.True(t, ok)
}

func TestResolveNestedValue(t *testing.T) {
	client := client(t, templateResponse(), nil)
	client.PutContext("targeting_key", "user1")

	evalDetails := client.GetBoolFlag(context.Background(), "test-flag.struct-key.boolean-key", true)
	assert.Equal(t, false, evalDetails.Value)
}

func TestResolveDoubleNestedValue(t *testing.T) {
	client := client(t, templateResponse(), nil)
	client.PutContext("targeting_key", "user1")

	evalDetails := client.GetBoolFlag(context.Background(), "test-flag.struct-key.nested-struct-key.nested-boolean-key", true)
	assert.Equal(t, false, evalDetails.Value)
}

func TestResolveWholeFlagAsObject(t *testing.T) {
	client := client(t, templateResponse(), nil)
	client.PutContext("targeting_key", "user1")

	evalDetails := client.GetObjectFlag(context.Background(), "test-flag", map[string]interface{}{})
	_, ok := evalDetails.Value.(map[string]interface{})
	assert.True(t, ok)
}

func TestResolveWholeFlagAsObjectWithInts(t *testing.T) {
	client := client(t, templateResponse(), nil)
	client.PutContext("targeting_key", "user1")

	evalDetails := client.GetObjectFlag(context.Background(), "test-flag", map[string]interface{}{})

	value, _ := evalDetails.Value.(map[string]interface{})
	rootIntValue := value["integer-key"]

	assert.Equal(t, reflect.Int64, reflect.ValueOf(rootIntValue).Kind())
	assert.Equal(t, int64(40), rootIntValue)

	nestedIntValue := value["struct-key"].(map[string]interface{})["integer-key"]

	assert.Equal(t, reflect.Int64, reflect.ValueOf(nestedIntValue).Kind())
	assert.Equal(t, int64(23), nestedIntValue)
}

func TestResolveWithWrongType(t *testing.T) {
	client := client(t, templateResponse(), nil)
	client.PutContext("targeting_key", "user1")

	evalDetails := client.GetBoolFlag(context.Background(), "test-flag.integer-key", false)

	assert.Equal(t, false, evalDetails.Value)
	assert.Equal(t, ErrorReason, evalDetails.Reason)
	assert.Equal(t, TypeMismatchCode, evalDetails.ErrorCode)
}

func TestResolveWithUnexpectedFlag(t *testing.T) {
	client := client(t, templateResponseWithFlagName("wrong-flag"), nil)
	client.PutContext("targeting_key", "user1")

	evalDetails := client.GetBoolFlag(context.Background(), "test-flag.boolean-key", true)

	assert.Equal(t, true, evalDetails.Value)
	assert.Equal(t, ErrorReason, evalDetails.Reason)
	assert.Equal(t, FlagNotFoundCode, evalDetails.ErrorCode)
	assert.Equal(t, "unexpected flag 'wrong-flag' from remote", evalDetails.ErrorMessage)
}

func TestResolveWithNonExistingFlag(t *testing.T) {
	client := client(t, emptyResponse(), nil)
	client.PutContext("targeting_key", "user1")

	evalDetails := client.GetBoolFlag(context.Background(), "test-flag.boolean-key", true)

	assert.Equal(t, true, evalDetails.Value)
	assert.Equal(t, ErrorReason, evalDetails.Reason)
	assert.Equal(t, FlagNotFoundCode, evalDetails.ErrorCode)
	assert.Equal(t, "Flag not found", evalDetails.ErrorMessage)
}

func client(t *testing.T, response ResolveResponse, errorToReturn error) *Confidence {
	confidence := newConfidence("apiKey", MockResolveClient{MockedResponse: response, MockedError: errorToReturn, TestingT: t})
	return confidence
}

func templateResponse() ResolveResponse {
	return templateResponseWithFlagName("test-flag")
}

func templateResponseWithFlagName(flagName string) ResolveResponse {
	templateResolveResponse := fmt.Sprintf(`
{
 "resolvedFlags": [
 {
  "flag": "flags/%[1]s",
  "variant": "flags/%[1]s/variants/treatment",
  "value": {
   "struct-key": {
    "boolean-key": false,
    "string-key": "treatment-struct",
    "double-key": 123.23,
    "integer-key": 23,
	"nested-struct-key": {
		"nested-boolean-key": false
	}
   },
   "boolean-key": true,
   "string-key": "treatment",
   "double-key": 20.203,
   "integer-key": 40
  },
  "flagSchema": {
   "schema": {
    "struct-key": {
     "structSchema": {
      "schema": {
       "boolean-key": {
        "boolSchema": {}
       },
       "string-key": {
        "stringSchema": {}
       },
       "double-key": {
        "doubleSchema": {}
       },
       "integer-key": {
        "intSchema": {}
       },
	   "nested-struct-key": {
		"structSchema": {
			"schema": {
				"nested-boolean-key": {
					"boolSchema": {}
				}
			}
		}
	   }
      }
     }
    },
    "boolean-key": {
     "boolSchema": {}
    },
    "string-key": {
     "stringSchema": {}
    },
    "double-key": {
     "doubleSchema": {}
    },
    "integer-key": {
     "intSchema": {}
    }
   }
  },
  "reason": "RESOLVE_REASON_MATCH"
 }],
 "resolveToken": ""
}
`, flagName)
	var result ResolveResponse
	decoder := json.NewDecoder(bytes.NewBuffer([]byte(templateResolveResponse)))
	decoder.UseNumber()
	_ = decoder.Decode(&result)
	return result
}

func emptyResponse() ResolveResponse {
	templateResolveResponse :=
		`
{
 "resolvedFlags": [],
 "resolveToken": ""
}
`
	var result ResolveResponse
	decoder := json.NewDecoder(bytes.NewBuffer([]byte(templateResolveResponse)))
	decoder.UseNumber()
	_ = decoder.Decode(&result)
	return result
}

func newConfidence(apiKey string, client ResolveClient) *Confidence {
	config := APIConfig{
		APIKey: apiKey,
	}
	return &Confidence{
		Config:        config,
		ResolveClient: client,
		contextMap:    make(map[string]interface{}),
	}
}
