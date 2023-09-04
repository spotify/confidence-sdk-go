package provider

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"testing"

	"github.com/open-feature/go-sdk/pkg/openfeature"
	"github.com/stretchr/testify/assert"
)

type MockResolveClient struct {
	MockedResponse resolveResponse
	MockedError    error
	ApplyRequests  *[]applyFlagRequest
}

func (r MockResolveClient) sendResolveRequest(_ context.Context,
	_ resolveRequest) (resolveResponse, error) {
	return r.MockedResponse, r.MockedError
}

func (r MockResolveClient) sendApplyRequest(_ context.Context, request applyFlagRequest) error {
	*r.ApplyRequests = append(*r.ApplyRequests, request)
	return nil
}

func TestResolveBoolValue(t *testing.T) {
	client := client(templateResponse(), nil)
	attributes := make(map[string]interface{})

	evalDetails, _ := client.BooleanValueDetails(
		context.Background(), "test-flag.boolean-key", false, openfeature.NewEvaluationContext(
			"dennis",
			attributes))

	assert.Equal(t, true, evalDetails.Value)
	assert.Equal(t, openfeature.TargetingMatchReason, evalDetails.Reason)
	assert.Equal(t, "flags/test-flag/variants/treatment", evalDetails.Variant)
	assert.Equal(t, "test-flag.boolean-key", evalDetails.FlagKey)
}

func TestResolveIntValue(t *testing.T) {
	client := client(templateResponse(), nil)
	attributes := make(map[string]interface{})

	evalDetails, _ := client.IntValueDetails(
		context.Background(), "test-flag.integer-key", 99, openfeature.NewEvaluationContext(
			"dennis",
			attributes))

	assert.Equal(t, int64(40), evalDetails.Value)
}

func TestResolveDoubleValue(t *testing.T) {
	client := client(templateResponse(), nil)
	attributes := make(map[string]interface{})

	evalDetails, _ := client.FloatValueDetails(
		context.Background(), "test-flag.double-key", 99.99, openfeature.NewEvaluationContext(
			"dennis",
			attributes))

	assert.Equal(t, 20.203, evalDetails.Value)
}

func TestResolveStringValue(t *testing.T) {
	client := client(templateResponse(), nil)
	attributes := make(map[string]interface{})

	evalDetails, _ := client.StringValueDetails(
		context.Background(), "test-flag.string-key", "default", openfeature.NewEvaluationContext(
			"dennis",
			attributes))

	assert.Equal(t, "treatment", evalDetails.Value)
}

func TestResolveObjectValue(t *testing.T) {
	client := client(templateResponse(), nil)
	attributes := make(map[string]interface{})

	evalDetails, _ := client.ObjectValueDetails(
		context.Background(), "test-flag.struct-key", "default", openfeature.NewEvaluationContext(
			"dennis",
			attributes))

	_, ok := evalDetails.Value.(map[string]interface{})
	assert.True(t, ok)
}

func TestResolveNestedValue(t *testing.T) {
	client := client(templateResponse(), nil)
	attributes := make(map[string]interface{})

	evalDetails, _ := client.BooleanValueDetails(
		context.Background(), "test-flag.struct-key.boolean-key", true, openfeature.NewEvaluationContext(
			"dennis",
			attributes))

	assert.Equal(t, false, evalDetails.Value)
}

func TestResolveWholeFlagAsObject(t *testing.T) {
	client := client(templateResponse(), nil)
	attributes := make(map[string]interface{})

	evalDetails, _ := client.ObjectValueDetails(
		context.Background(), "test-flag", "default", openfeature.NewEvaluationContext(
			"dennis",
			attributes))

	_, ok := evalDetails.Value.(map[string]interface{})
	assert.True(t, ok)
}

func TestResolveWholeFlagAsObjectWithInts(t *testing.T) {
	client := client(templateResponse(), nil)
	attributes := make(map[string]interface{})

	evalDetails, _ := client.ObjectValueDetails(
		context.Background(), "test-flag", "default", openfeature.NewEvaluationContext(
			"dennis",
			attributes))

	value, _ := evalDetails.Value.(map[string]interface{})
	rootIntValue := value["integer-key"]

	assert.Equal(t, reflect.Int64, reflect.ValueOf(rootIntValue).Kind())
	assert.Equal(t, int64(40), rootIntValue)

	nestedIntValue := value["struct-key"].(map[string]interface{})["integer-key"]

	assert.Equal(t, reflect.Int64, reflect.ValueOf(nestedIntValue).Kind())
	assert.Equal(t, int64(23), nestedIntValue)
}

func TestResolveWithWrongType(t *testing.T) {
	client := client(templateResponse(), nil)
	attributes := make(map[string]interface{})

	evalDetails, _ := client.BooleanValueDetails(
		context.Background(), "test-flag.integer-key", false, openfeature.NewEvaluationContext(
			"dennis",
			attributes))

	assert.Equal(t, false, evalDetails.Value)
	assert.Equal(t, openfeature.ErrorReason, evalDetails.Reason)
	assert.Equal(t, openfeature.TypeMismatchCode, evalDetails.ErrorCode)
}

func TestResolveWithUnexpectedFlag(t *testing.T) {
	client := client(templateResponseWithFlagName("wrong-flag"), nil)
	attributes := make(map[string]interface{})

	evalDetails, _ := client.BooleanValueDetails(
		context.Background(), "test-flag.boolean-key", true, openfeature.NewEvaluationContext(
			"dennis",
			attributes))

	assert.Equal(t, true, evalDetails.Value)
	assert.Equal(t, openfeature.ErrorReason, evalDetails.Reason)
	assert.Equal(t, openfeature.FlagNotFoundCode, evalDetails.ErrorCode)
	assert.Equal(t, "unexpected flag 'wrong-flag' from remote", evalDetails.ErrorMessage)
}

func TestResolveWithNonExistingFlag(t *testing.T) {
	client := client(emptyResponse(), nil)
	attributes := make(map[string]interface{})

	evalDetails, _ := client.BooleanValueDetails(
		context.Background(), "test-flag.boolean-key", true, openfeature.NewEvaluationContext(
			"dennis",
			attributes))

	assert.Equal(t, true, evalDetails.Value)
	assert.Equal(t, openfeature.ErrorReason, evalDetails.Reason)
	assert.Equal(t, openfeature.FlagNotFoundCode, evalDetails.ErrorCode)
	assert.Equal(t, "no active flag 'test-flag' was found", evalDetails.ErrorMessage)
}

func client(response resolveResponse, errorToReturn error) *openfeature.Client {
	provider := FlagProvider{Config: APIConfig{APIKey: "apikey",
		Region: APIRegionEU}, ResolveClient: MockResolveClient{MockedResponse: response, MockedError: errorToReturn}}
	openfeature.SetProvider(provider)
	return openfeature.NewClient("testApp")
}

func templateResponse() resolveResponse {
	return templateResponseWithFlagName("test-flag")
}

func templateResponseWithFlagName(flagName string) resolveResponse {
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
    "integer-key": 23
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
	var result resolveResponse
	decoder := json.NewDecoder(bytes.NewBuffer([]byte(templateResolveResponse)))
	decoder.UseNumber()
	_ = decoder.Decode(&result)
	return result
}

func emptyResponse() resolveResponse {
	templateResolveResponse :=
		`
{
 "resolvedFlags": [],
 "resolveToken": ""
}
`
	var result resolveResponse
	decoder := json.NewDecoder(bytes.NewBuffer([]byte(templateResolveResponse)))
	decoder.UseNumber()
	_ = decoder.Decode(&result)
	return result
}
