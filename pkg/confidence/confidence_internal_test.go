package confidence

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"testing"

	"golang.org/x/exp/slog"

	"github.com/stretchr/testify/assert"
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

func TestResolveBoolNullValue(t *testing.T) {
	client := client(t, templateResponse(), nil)
	client.PutContext("targeting_key", "user1")

	evalDetails := client.GetBoolFlag(context.Background(), "test-flag.boolean-key-null-value", true)

	assert.Equal(t, true, evalDetails.Value)
	assert.Equal(t, ErrorCode(""), evalDetails.ErrorCode)
	assert.Equal(t, "", evalDetails.ErrorMessage)
	assert.Equal(t, TargetingMatchReason, evalDetails.Reason)
	assert.Equal(t, "flags/test-flag/variants/treatment", evalDetails.Variant)
}

func TestResolveIntValue(t *testing.T) {
	client := client(t, templateResponse(), nil)
	client.PutContext("targeting_key", "user1")

	evalDetails := client.GetIntFlag(context.Background(), "test-flag.integer-key", 99)

	assert.Equal(t, int64(40), evalDetails.Value)
}

func TestResolveIntNullValue(t *testing.T) {
	client := client(t, templateResponse(), nil)
	client.PutContext("targeting_key", "user1")

	evalDetails := client.GetIntFlag(context.Background(), "test-flag.integer-key-null-value", 99)

	assert.Equal(t, int64(99), evalDetails.Value)
	assert.Equal(t, ErrorCode(""), evalDetails.ErrorCode)
	assert.Equal(t, "", evalDetails.ErrorMessage)
	assert.Equal(t, TargetingMatchReason, evalDetails.Reason)
	assert.Equal(t, "flags/test-flag/variants/treatment", evalDetails.Variant)
}

func TestResolveDoubleValue(t *testing.T) {
	client := client(t, templateResponse(), nil)
	client.PutContext("targeting_key", "user1")

	evalDetails := client.GetDoubleFlag(context.Background(), "test-flag.double-key", 99.99)

	assert.Equal(t, 20.203, evalDetails.Value)
}

func TestResolveDoubleNullValue(t *testing.T) {
	client := client(t, templateResponse(), nil)
	client.PutContext("targeting_key", "user1")

	evalDetails := client.GetDoubleFlag(context.Background(), "test-flag.double-key-null-value", 99.99)

	assert.Equal(t, 99.99, evalDetails.Value)
	assert.Equal(t, ErrorCode(""), evalDetails.ErrorCode)
	assert.Equal(t, "", evalDetails.ErrorMessage)
	assert.Equal(t, TargetingMatchReason, evalDetails.Reason)
	assert.Equal(t, "flags/test-flag/variants/treatment", evalDetails.Variant)
}

func TestResolveStringValue(t *testing.T) {
	client := client(t, templateResponse(), nil)
	client.PutContext("targeting_key", "user1")

	evalDetails := client.GetStringFlag(context.Background(), "test-flag.string-key", "default")

	assert.Equal(t, "treatment", evalDetails.Value)
}

func TestResolveStringValueNullValue(t *testing.T) {
	client := client(t, templateResponse(), nil)
	client.PutContext("targeting_key", "user1")

	evalDetails := client.GetStringFlag(context.Background(), "test-flag.string-key-null-value", "default")

	assert.Equal(t, "default", evalDetails.Value)
	assert.Equal(t, ErrorCode(""), evalDetails.ErrorCode)
	assert.Equal(t, "", evalDetails.ErrorMessage)
	assert.Equal(t, TargetingMatchReason, evalDetails.Reason)
	assert.Equal(t, "flags/test-flag/variants/treatment", evalDetails.Variant)
}

func TestResolveObjectValue(t *testing.T) {
	client := client(t, templateResponse(), nil)
	client.PutContext("targeting_key", "user1")

	evalDetails := client.GetObjectFlag(context.Background(), "test-flag.struct-key", map[string]interface{}{})
	_, ok := evalDetails.Value.(map[string]interface{})
	assert.True(t, ok)
}

func TestResolveBoolNestedValue(t *testing.T) {
	client := client(t, templateResponse(), nil)
	client.PutContext("targeting_key", "user1")

	evalDetails := client.GetBoolFlag(context.Background(), "test-flag.struct-key.boolean-key", true)
	assert.Equal(t, false, evalDetails.Value)
}

func TestResolveBoolNestedNullValue(t *testing.T) {
	client := client(t, templateResponse(), nil)
	client.PutContext("targeting_key", "user1")

	evalDetails := client.GetBoolFlag(context.Background(), "test-flag.struct-key.boolean-key-null-value", true)
	assert.Equal(t, true, evalDetails.Value)
	assert.Equal(t, ErrorCode(""), evalDetails.ErrorCode)
	assert.Equal(t, "", evalDetails.ErrorMessage)
	assert.Equal(t, TargetingMatchReason, evalDetails.Reason)
	assert.Equal(t, "flags/test-flag/variants/treatment", evalDetails.Variant)
}

func TestResolveStringNestedValue(t *testing.T) {
	client := client(t, templateResponse(), nil)
	client.PutContext("targeting_key", "user1")

	evalDetails := client.GetStringFlag(context.Background(), "test-flag.struct-key.string-key", "default")
	assert.Equal(t, "treatment-struct", evalDetails.Value)
}

func TestResolveStringNestedNullValue(t *testing.T) {
	client := client(t, templateResponse(), nil)
	client.PutContext("targeting_key", "user1")

	evalDetails := client.GetStringFlag(context.Background(), "test-flag.struct-key.string-key-null-value", "default")
	assert.Equal(t, "default", evalDetails.Value)
	assert.Equal(t, ErrorCode(""), evalDetails.ErrorCode)
	assert.Equal(t, "", evalDetails.ErrorMessage)
	assert.Equal(t, TargetingMatchReason, evalDetails.Reason)
	assert.Equal(t, "flags/test-flag/variants/treatment", evalDetails.Variant)
}

func TestResolveDoubleNestedValue(t *testing.T) {
	client := client(t, templateResponse(), nil)
	client.PutContext("targeting_key", "user1")

	evalDetails := client.GetDoubleFlag(context.Background(), "test-flag.struct-key.double-key", 99.99)
	assert.Equal(t, 123.23, evalDetails.Value)
}

func TestResolveDoubleNestedNullValue(t *testing.T) {
	client := client(t, templateResponse(), nil)
	client.PutContext("targeting_key", "user1")

	evalDetails := client.GetDoubleFlag(context.Background(), "test-flag.struct-key.double-key-null-value", 99.99)
	assert.Equal(t, 99.99, evalDetails.Value)
	assert.Equal(t, ErrorCode(""), evalDetails.ErrorCode)
	assert.Equal(t, "", evalDetails.ErrorMessage)
	assert.Equal(t, TargetingMatchReason, evalDetails.Reason)
	assert.Equal(t, "flags/test-flag/variants/treatment", evalDetails.Variant)
}

func TestResolveIntNestedValue(t *testing.T) {
	client := client(t, templateResponse(), nil)
	client.PutContext("targeting_key", "user1")

	evalDetails := client.GetIntFlag(context.Background(), "test-flag.struct-key.integer-key", 99)
	assert.Equal(t, int64(23), evalDetails.Value)
}

func TestResolveIntNestedNullValue(t *testing.T) {
	client := client(t, templateResponse(), nil)
	client.PutContext("targeting_key", "user1")

	evalDetails := client.GetIntFlag(context.Background(), "test-flag.struct-key.integer-key-null-value", 99)
	assert.Equal(t, int64(99), evalDetails.Value)
	assert.Equal(t, ErrorCode(""), evalDetails.ErrorCode)
	assert.Equal(t, "", evalDetails.ErrorMessage)
	assert.Equal(t, TargetingMatchReason, evalDetails.Reason)
	assert.Equal(t, "flags/test-flag/variants/treatment", evalDetails.Variant)
}

// Struct - In - Struct
func TestResolveBoolNestedStructValue(t *testing.T) {
	client := client(t, templateResponse(), nil)
	client.PutContext("targeting_key", "user1")

	evalDetails := client.GetBoolFlag(context.Background(), "test-flag.struct-key.nested-struct-key.nested-boolean-key", true)
	assert.Equal(t, false, evalDetails.Value)
}

func TestResolveBoolNestedStructNullValue(t *testing.T) {
	client := client(t, templateResponse(), nil)
	client.PutContext("targeting_key", "user1")

	evalDetails := client.GetBoolFlag(context.Background(), "test-flag.struct-key.nested-struct-key.nested-boolean-key-null-value", true)
	assert.Equal(t, true, evalDetails.Value)
	assert.Equal(t, ErrorCode(""), evalDetails.ErrorCode)
	assert.Equal(t, "", evalDetails.ErrorMessage)
	assert.Equal(t, TargetingMatchReason, evalDetails.Reason)
	assert.Equal(t, "flags/test-flag/variants/treatment", evalDetails.Variant)
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

func TestResolveWholeFlagAsObjectWithNulls(t *testing.T) {
	client := client(t, templateResponse(), nil)
	client.PutContext("targeting_key", "user1")

	evalDetails := client.GetObjectFlag(context.Background(), "test-flag", map[string]interface{}{})

	value, _ := evalDetails.Value.(map[string]interface{})
	rootNullValue := value["boolean-key-null-value"]

	assert.Nil(t, rootNullValue)

	nestedNullValue := value["struct-key"].(map[string]interface{})["string-key-null-value"]

	assert.Nil(t, nestedNullValue)
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
                    "boolean-key-null-value": null,
                    "string-key": "treatment-struct",
                    "string-key-null-value": null,
                    "double-key": 123.23,
                    "double-key-null-value": null,
                    "integer-key": 23,
                    "integer-key-null-value": null,
                    "nested-struct-key": {
                        "nested-boolean-key": false,
                        "nested-boolean-key-null-value": null
                    }
                },
                "boolean-key": true,
                "boolean-key-null-value": null,
                "string-key": "treatment",
                "string-key-null-value": null,
                "double-key": 20.203,
                "double-key-null-value": null,
                "integer-key": 40,
                "integer-key-null-value": null
            },
            "flagSchema": {
                "schema": {
                    "struct-key": {
                        "structSchema": {
                            "schema": {
                                "boolean-key": {
                                    "boolSchema": {}
                                },
                                "boolean-key-null-value": {
                                    "boolSchema": {}
                                },
                                "string-key": {
                                    "stringSchema": {}
                                },
                                "string-key-null-value": {
                                    "stringSchema": {}
                                },
                                "double-key": {
                                    "doubleSchema": {}
                                },
                                "double-key-null-value": {
                                    "doubleSchema": {}
                                },
                                "integer-key": {
                                    "intSchema": {}
                                },
                                "integer-key-null-value": {
                                    "intSchema": {}
                                },
                                "nested-struct-key": {
                                    "structSchema": {
                                        "schema": {
                                            "nested-boolean-key": {
                                                "boolSchema": {}
                                            },
                                            "nested-boolean-key-null-value": {
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
                    "boolean-key-null-value": {
                        "boolSchema": {}
                    },
                    "string-key": {
                        "stringSchema": {}
                    },
                    "string-key-null-value": {
                        "stringSchema": {}
                    },
                    "double-key": {
                        "doubleSchema": {}
                    },
                    "double-key-null-value": {
                        "doubleSchema": {}
                    },
                    "integer-key": {
                        "intSchema": {}
                    },
                    "integer-key-null-value": {
                        "intSchema": {}
                    }
                }
            },
            "reason": "RESOLVE_REASON_MATCH"
        }
    ],
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
		Logger:        slog.Default(),
	}
}
