package provider

import (
	"encoding/json"
	"errors"
	"github.com/open-feature/go-sdk/pkg/openfeature"
	"github.com/stretchr/testify/assert"
	"reflect"
	"testing"
)

func TestSpitFlagString(t *testing.T) {
	t.Run("FlagWithValue", func(t *testing.T) {
		val1, val2 := splitFlagString("name.value")
		assert.Equal(t, "name", val1)
		assert.Equal(t, "value", val2)
	})

	t.Run("FlagWithoutSecondPart", func(t *testing.T) {
		val1, val2 := splitFlagString("novalue")
		assert.Equal(t, "novalue", val1)
		assert.Equal(t, "", val2)
	})

	t.Run("FlagWithMultipleDots", func(t *testing.T) {
		val1, val2 := splitFlagString("double.dot.value")
		assert.Equal(t, "double", val1)
		assert.Equal(t, "dot.value", val2)
	})
}

func TestExtractPropertyValue(t *testing.T) {
	t.Run("PathFromMap", func(t *testing.T) {
		values := map[string]interface{}{
			"child": map[string]interface{}{
				"key": "value",
			},
			"key": "no-value",
		}

		got, err := extractPropertyValue("child.key", values)
		assert.NoError(t, err)
		assert.Equal(t, "value", got)
	})

	t.Run("DirectPathFromMap", func(t *testing.T) {
		values := map[string]interface{}{
			"key": "direct-value",
		}

		got, err := extractPropertyValue("key", values)
		assert.NoError(t, err)
		assert.Equal(t, "direct-value", got)
	})

	t.Run("PathNotFound", func(t *testing.T) {
		values := map[string]interface{}{
			"valid": map[string]interface{}{
				"path": "value",
			},
		}

		got, err := extractPropertyValue("invalid.path", values)
		assert.Error(t, err)
		assert.Equal(t, false, got)
	})
}

func TestGetTypeForPath(t *testing.T) {
	t.Run("EmptyPath", func(t *testing.T) {
		schema := map[string]interface{}{
			"key": "value",
		}

		got, err := getTypeForPath(schema, "")
		assert.NoError(t, err)
		assert.Equal(t, reflect.Map, got)
	})

	t.Run("BoolSchema", func(t *testing.T) {
		schema := map[string]interface{}{
			"key": map[string]interface{}{
				"boolSchema": true,
			},
		}

		got, err := getTypeForPath(schema, "key")
		assert.NoError(t, err)
		assert.Equal(t, reflect.Bool, got)
	})

	t.Run("StringSchema", func(t *testing.T) {
		schema := map[string]interface{}{
			"key": map[string]interface{}{
				"stringSchema": "value",
			},
		}

		got, err := getTypeForPath(schema, "key")
		assert.NoError(t, err)
		assert.Equal(t, reflect.String, got)
	})

	t.Run("IntSchema", func(t *testing.T) {
		schema := map[string]interface{}{
			"key": map[string]interface{}{
				"intSchema": 123,
			},
		}

		got, err := getTypeForPath(schema, "key")
		assert.NoError(t, err)
		assert.Equal(t, reflect.Int64, got)
	})

	t.Run("FloatSchema", func(t *testing.T) {
		schema := map[string]interface{}{
			"key": map[string]interface{}{
				"doubleSchema": 123.456,
			},
		}

		got, err := getTypeForPath(schema, "key")
		assert.NoError(t, err)
		assert.Equal(t, reflect.Float64, got)
	})

	t.Run("MapSchema", func(t *testing.T) {
		schema := map[string]interface{}{
			"key": map[string]interface{}{
				"structSchema": map[string]interface{}{
					"schema": map[string]interface{}{
						"nested": map[string]interface{}{
							"structSchema": map[string]interface{}{},
						},
					},
				},
			},
		}

		got, err := getTypeForPath(schema, "key.nested")
		assert.NoError(t, err)
		assert.Equal(t, reflect.Map, got)
	})

	t.Run("PropertyNotFound", func(t *testing.T) {
		schema := map[string]interface{}{
			"valid": map[string]interface{}{
				"boolSchema": true,
			},
		}

		_, err := getTypeForPath(schema, "invalid")
		assert.Error(t, err)
	})
}

func TestProcessResolveError(t *testing.T) {
	defaultValue := "default"

	t.Run("FlagNotFoundError", func(t *testing.T) {
		res := processResolveError(errFlagNotFound, defaultValue)
		assert.Equal(t, defaultValue, res.Value)
		assert.IsType(t, openfeature.ResolutionError{}, res.ProviderResolutionDetail.ResolutionError)

		resDetails := res.ProviderResolutionDetail.ResolutionDetail()
		assert.Equal(t, openfeature.FlagNotFoundCode, resDetails.ErrorCode)
		assert.Equal(t, openfeature.ErrorReason, resDetails.Reason)
	})

	t.Run("GeneralError", func(t *testing.T) {
		err := errors.New("unknown error")
		res := processResolveError(err, defaultValue)
		assert.Equal(t, defaultValue, res.Value)
		assert.IsType(t, openfeature.ResolutionError{}, res.ProviderResolutionDetail.ResolutionError)

		resDetails := res.ProviderResolutionDetail.ResolutionDetail()
		assert.Equal(t, openfeature.GeneralCode, resDetails.ErrorCode)
		assert.Equal(t, openfeature.ErrorReason, resDetails.Reason)
	})
}

func TestProcessResolvedFlag(t *testing.T) {
	t.Run("EmptyValue", func(t *testing.T) {
		defaultValue := "default"
		rf := resolvedFlag{
			Value:      map[string]interface{}{},
			FlagSchema: flagSchema{Schema: map[string]interface{}{}},
		}

		expected := openfeature.InterfaceResolutionDetail{
			Value: defaultValue,
			ProviderResolutionDetail: openfeature.ProviderResolutionDetail{
				Reason: openfeature.DefaultReason,
			},
		}

		assert.Equal(t, expected, processResolvedFlag(rf, defaultValue, reflect.String, ""))
	})

	t.Run("TypeMismatchError", func(t *testing.T) {
		defaultValue := "default"
		rf := resolvedFlag{
			Value:      map[string]interface{}{"key": "value"},
			FlagSchema: flagSchema{Schema: map[string]interface{}{"key": "wrongType"}},
		}

		expected := openfeature.InterfaceResolutionDetail{
			Value: defaultValue,
			ProviderResolutionDetail: openfeature.ProviderResolutionDetail{
				ResolutionError: openfeature.NewTypeMismatchResolutionError(
					"schema for property key does not match the expected type"),
				Reason: openfeature.ErrorReason,
			},
		}

		assert.Equal(t, expected, processResolvedFlag(rf, defaultValue, reflect.String, "key"))
	})

	t.Run("ExtractValueError", func(t *testing.T) {
		defaultValue := "default"
		rf := resolvedFlag{
			Value:      map[string]interface{}{"key": "value"},
			FlagSchema: flagSchema{Schema: map[string]interface{}{"key": "value"}},
		}

		expected := typeMismatchError(defaultValue)
		expected.ProviderResolutionDetail.ResolutionError =
			openfeature.NewTypeMismatchResolutionError("schema for property key.missing does not match the expected type")

		assert.Equal(t, expected, processResolvedFlag(rf, defaultValue, reflect.String, "key.missing"))
	})

	t.Run("Success", func(t *testing.T) {
		defaultValue := "default"
		rf := resolvedFlag{
			Value:      map[string]interface{}{"key": "value"},
			FlagSchema: flagSchema{Schema: map[string]interface{}{"key": map[string]interface{}{"stringSchema": "value"}}},
		}

		expected := openfeature.InterfaceResolutionDetail{
			Value: "value", // Success case excludes default value
			ProviderResolutionDetail: openfeature.ProviderResolutionDetail{
				Reason: openfeature.TargetingMatchReason,
			},
		}
		assert.Equal(t, expected, processResolvedFlag(rf, defaultValue, reflect.String, "key"))
	})
}

func TestReplaceNumbers(t *testing.T) {
	t.Run("SuccessfulFlowFloat64", func(t *testing.T) {
		schema := map[string]interface{}{
			"key": map[string]interface{}{
				"doubleSchema": 123.45,
			},
		}
		input := map[string]interface{}{"key": json.Number("123.45")}
		expected := map[string]interface{}{"key": float64(123.45)}
		updatedMap, err := replaceNumbers("", input, schema)

		assert.NoError(t, err)
		assert.Equal(t, expected, updatedMap)
	})

	t.Run("SuccessfulFlowInt64", func(t *testing.T) {
		schema := map[string]interface{}{
			"key": map[string]interface{}{
				"intSchema": 123,
			},
		}
		input := map[string]interface{}{"key": json.Number("123")}
		expected := map[string]interface{}{"key": int64(123)}
		updatedMap, err := replaceNumbers("", input, schema)

		assert.NoError(t, err)
		assert.Equal(t, expected, updatedMap)
	})

	t.Run("SuccessfulFlowMap", func(t *testing.T) {
		schema := map[string]interface{}{
			"key": map[string]interface{}{
				"structSchema": map[string]interface{}{
					"schema": map[string]interface{}{
						"subKey": map[string]interface{}{
							"doubleSchema": 123.45,
						},
					},
				},
			},
		}
		input := map[string]interface{}{
			"key": map[string]interface{}{
				"subKey": json.Number("123.45"),
			},
		}
		expected := map[string]interface{}{
			"key": map[string]interface{}{
				"subKey": float64(123.45),
			},
		}
		updatedMap, err := replaceNumbers("", input, schema)

		assert.NoError(t, err)
		assert.Equal(t, expected, updatedMap)
	})

	t.Run("SuccessfulNestedFlowMap", func(t *testing.T) {
		schema := map[string]interface{}{
			"structKey": map[string]interface{}{
				"structSchema": map[string]interface{}{
					"schema": map[string]interface{}{
						"nestedStructKey": map[string]interface{}{
							"structSchema": map[string]interface{}{
								"schema": map[string]interface{}{
									"nestedDoubleKey": map[string]interface{}{
										"doubleSchema": map[string]interface{}{},
									},
								},
							},
						},
					},
				},
			},
		}

		input := map[string]interface{}{
			"structKey": map[string]interface{}{
				"nestedStructKey": map[string]interface{}{
					"nestedDoubleKey": json.Number("123.45"),
				},
			},
		}

		expected := map[string]interface{}{
			"structKey": map[string]interface{}{
				"nestedStructKey": map[string]interface{}{
					"nestedDoubleKey": float64(123.45),
				},
			},
		}

		updatedMap, err := replaceNumbers("", input, schema)
		assert.NoError(t, err)
		assert.Equal(t, expected, updatedMap)
	})
}

func TestTypeMismatchError(t *testing.T) {
	t.Run("WithStringValue", func(t *testing.T) {
		defaultValue := "my default value"
		expected := openfeature.InterfaceResolutionDetail{
			Value: defaultValue,
			ProviderResolutionDetail: openfeature.ProviderResolutionDetail{
				ResolutionError: openfeature.NewTypeMismatchResolutionError(
					"Unable to extract property value from resolve response"),
				Reason: openfeature.ErrorReason,
			},
		}

		assert.Equal(t, expected, typeMismatchError(defaultValue))
	})

	t.Run("WithIntValue", func(t *testing.T) {
		defaultValue := 123
		expected := openfeature.InterfaceResolutionDetail{
			Value: defaultValue,
			ProviderResolutionDetail: openfeature.ProviderResolutionDetail{
				ResolutionError: openfeature.NewTypeMismatchResolutionError(
					"Unable to extract property value from resolve response"),
				Reason: openfeature.ErrorReason,
			},
		}

		assert.Equal(t, expected, typeMismatchError(defaultValue))
	})
}

func TestToBoolResolutionDetail(t *testing.T) {
	defaultValue := false

	t.Run("WhenValueIsBool", func(t *testing.T) {
		res := openfeature.InterfaceResolutionDetail{
			Value: true,
			ProviderResolutionDetail: openfeature.ProviderResolutionDetail{
				Reason: openfeature.TargetingMatchReason,
			},
		}

		expected := openfeature.BoolResolutionDetail{
			Value: true,
			ProviderResolutionDetail: openfeature.ProviderResolutionDetail{
				Reason: openfeature.TargetingMatchReason,
			},
		}

		assert.Equal(t, expected, toBoolResolutionDetail(res, defaultValue))
	})

	t.Run("WhenValueIsNotBool", func(t *testing.T) {
		res := openfeature.InterfaceResolutionDetail{
			Value: "not a bool",
			ProviderResolutionDetail: openfeature.ProviderResolutionDetail{
				Reason: openfeature.TargetingMatchReason,
			},
		}

		expected := openfeature.BoolResolutionDetail{
			Value: defaultValue,
			ProviderResolutionDetail: openfeature.ProviderResolutionDetail{
				ResolutionError: openfeature.NewTypeMismatchResolutionError("Unable to convert response property to boolean"),
				Reason:          openfeature.ErrorReason,
			},
		}

		assert.Equal(t, expected, toBoolResolutionDetail(res, defaultValue))
	})

	t.Run("WhenReasonIsNotTargetingMatchReason", func(t *testing.T) {
		res := openfeature.InterfaceResolutionDetail{
			Value: true,
			ProviderResolutionDetail: openfeature.ProviderResolutionDetail{
				Reason: openfeature.ErrorReason,
			},
		}

		expected := openfeature.BoolResolutionDetail{
			Value: defaultValue,
			ProviderResolutionDetail: openfeature.ProviderResolutionDetail{
				Reason: openfeature.ErrorReason,
			},
		}

		assert.Equal(t, expected, toBoolResolutionDetail(res, defaultValue))
	})
}

func TestToStringResolutionDetail(t *testing.T) {
	defaultValue := "default"

	t.Run("WhenValueIsString", func(t *testing.T) {
		res := openfeature.InterfaceResolutionDetail{
			Value: "hello",
			ProviderResolutionDetail: openfeature.ProviderResolutionDetail{
				Reason: openfeature.TargetingMatchReason,
			},
		}

		expected := openfeature.StringResolutionDetail{
			Value: "hello",
			ProviderResolutionDetail: openfeature.ProviderResolutionDetail{
				Reason: openfeature.TargetingMatchReason,
			},
		}

		assert.Equal(t, expected, toStringResolutionDetail(res, defaultValue))
	})

	t.Run("WhenValueIsNotString", func(t *testing.T) {
		res := openfeature.InterfaceResolutionDetail{
			Value: 123,
			ProviderResolutionDetail: openfeature.ProviderResolutionDetail{
				Reason: openfeature.TargetingMatchReason,
			},
		}

		expected := openfeature.StringResolutionDetail{
			Value: defaultValue,
			ProviderResolutionDetail: openfeature.ProviderResolutionDetail{
				ResolutionError: openfeature.NewTypeMismatchResolutionError("Unable to convert response property to boolean"),
				Reason:          openfeature.ErrorReason,
			},
		}

		assert.Equal(t, expected, toStringResolutionDetail(res, defaultValue))
	})

	t.Run("WhenReasonIsNotTargetingMatchReason", func(t *testing.T) {
		res := openfeature.InterfaceResolutionDetail{
			Value: "hello",
			ProviderResolutionDetail: openfeature.ProviderResolutionDetail{
				Reason: openfeature.ErrorReason,
			},
		}

		expected := openfeature.StringResolutionDetail{
			Value: defaultValue,
			ProviderResolutionDetail: openfeature.ProviderResolutionDetail{
				Reason: openfeature.ErrorReason,
			},
		}

		assert.Equal(t, expected, toStringResolutionDetail(res, defaultValue))
	})
}

func TestToFloatResolutionDetail(t *testing.T) {
	defaultValue := 42.0

	t.Run("WhenValueIsFloat", func(t *testing.T) {
		res := openfeature.InterfaceResolutionDetail{
			Value: 24.0,
			ProviderResolutionDetail: openfeature.ProviderResolutionDetail{
				Reason: openfeature.TargetingMatchReason,
			},
		}

		expected := openfeature.FloatResolutionDetail{
			Value: 24.0,
			ProviderResolutionDetail: openfeature.ProviderResolutionDetail{
				Reason: openfeature.TargetingMatchReason,
			},
		}

		assert.Equal(t, expected, toFloatResolutionDetail(res, defaultValue))
	})

	t.Run("WhenValueIsNotFloat", func(t *testing.T) {
		res := openfeature.InterfaceResolutionDetail{
			Value: "not a float",
			ProviderResolutionDetail: openfeature.ProviderResolutionDetail{
				Reason: openfeature.TargetingMatchReason,
			},
		}

		expected := openfeature.FloatResolutionDetail{
			Value: defaultValue,
			ProviderResolutionDetail: openfeature.ProviderResolutionDetail{
				ResolutionError: openfeature.NewTypeMismatchResolutionError("Unable to convert response property to float"),
				Reason:          openfeature.ErrorReason,
			},
		}

		assert.Equal(t, expected, toFloatResolutionDetail(res, defaultValue))
	})

	t.Run("WhenReasonIsNotTargetingMatchReason", func(t *testing.T) {
		res := openfeature.InterfaceResolutionDetail{
			Value: 24.0,
			ProviderResolutionDetail: openfeature.ProviderResolutionDetail{
				Reason: openfeature.ErrorReason,
			},
		}

		expected := openfeature.FloatResolutionDetail{
			Value: defaultValue,
			ProviderResolutionDetail: openfeature.ProviderResolutionDetail{
				Reason: openfeature.ErrorReason,
			},
		}

		assert.Equal(t, expected, toFloatResolutionDetail(res, defaultValue))
	})
}

func TestToIntResolutionDetail(t *testing.T) {
	defaultValue := int64(123)
	t.Run("WhenValueIsInt", func(t *testing.T) {
		res := openfeature.InterfaceResolutionDetail{
			Value: int64(456),
			ProviderResolutionDetail: openfeature.ProviderResolutionDetail{
				Reason: openfeature.TargetingMatchReason,
			},
		}

		expected := openfeature.IntResolutionDetail{
			Value: int64(456),
			ProviderResolutionDetail: openfeature.ProviderResolutionDetail{
				Reason: openfeature.TargetingMatchReason,
			},
		}

		assert.Equal(t, expected, toIntResolutionDetail(res, defaultValue))
	})

	t.Run("WhenValueIsNotInt", func(t *testing.T) {
		res := openfeature.InterfaceResolutionDetail{
			Value: "not an int",
			ProviderResolutionDetail: openfeature.ProviderResolutionDetail{
				Reason: openfeature.TargetingMatchReason,
			},
		}

		expected := openfeature.IntResolutionDetail{
			Value: defaultValue,
			ProviderResolutionDetail: openfeature.ProviderResolutionDetail{
				ResolutionError: openfeature.NewTypeMismatchResolutionError("Unable to convert response property to int"),
				Reason:          openfeature.ErrorReason,
			},
		}

		assert.Equal(t, expected, toIntResolutionDetail(res, defaultValue))
	})

	t.Run("WhenReasonIsNotTargetingMatchReason", func(t *testing.T) {
		res := openfeature.InterfaceResolutionDetail{
			Value: int64(456),
			ProviderResolutionDetail: openfeature.ProviderResolutionDetail{
				Reason: openfeature.ErrorReason,
			},
		}

		expected := openfeature.IntResolutionDetail{
			Value: defaultValue,
			ProviderResolutionDetail: openfeature.ProviderResolutionDetail{
				Reason: openfeature.ErrorReason,
			},
		}

		assert.Equal(t, expected, toIntResolutionDetail(res, defaultValue))
	})
}
