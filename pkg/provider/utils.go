package provider

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strings"

	"github.com/open-feature/go-sdk/openfeature"
)

func splitFlagString(flag string) (string, string) {
	splittedFlag := strings.SplitN(flag, ".", 2)
	if len(splittedFlag) == 2 {
		return splittedFlag[0], splittedFlag[1]
	}

	return splittedFlag[0], ""
}

func extractPropertyValue(path string, values map[string]interface{}) (interface{}, error) {
	if path == "" {
		return values, nil
	}

	firstPartAndRest := strings.SplitN(path, ".", 2)
	if len(firstPartAndRest) == 1 {
		value := values[firstPartAndRest[0]]
		return value, nil
	}

	childMap, ok := values[firstPartAndRest[0]].(map[string]interface{})
	if ok {
		return extractPropertyValue(firstPartAndRest[1], childMap)
	}

	return false, fmt.Errorf("unable to find property in path %s", path)
}

func getTypeForPath(schema map[string]interface{}, path string) (reflect.Kind, error) {
	if path == "" {
		return reflect.Map, nil
	}

	firstPartAndRest := strings.SplitN(path, ".", 2)
	if len(firstPartAndRest) == 1 {
		value, ok := schema[firstPartAndRest[0]].(map[string]interface{})
		if !ok {
			return 0, fmt.Errorf("schema was not in the expected format")
		}

		if _, isBool := value["boolSchema"]; isBool {
			return reflect.Bool, nil
		} else if _, isString := value["stringSchema"]; isString {
			return reflect.String, nil
		} else if _, isInt := value["intSchema"]; isInt {
			return reflect.Int64, nil
		} else if _, isFloat := value["doubleSchema"]; isFloat {
			return reflect.Float64, nil
		} else if _, isMap := value["structSchema"]; isMap {
			return reflect.Map, nil
		}

		return 0, fmt.Errorf("unable to find property type in schema %s", path)
	}

	// If we are here, the property path contains multiple entries -> this must be a struct -> recurse down the tree.
	childMap, ok := schema[firstPartAndRest[0]].(map[string]interface{})
	if !ok {
		return 0, fmt.Errorf("unexpected error when parsing resolve response schema")
	}

	if structMap, isStruct := childMap["structSchema"]; isStruct {
		structSchema, _ := structMap.(map[string]interface{})["schema"].(map[string]interface{})
		return getTypeForPath(structSchema, firstPartAndRest[1])
	}

	return 0, fmt.Errorf("unable to find property in schema %s", path)
}

func processResolveError(err error, defaultValue interface{}) openfeature.InterfaceResolutionDetail {
	switch {
	case errors.Is(err, errFlagNotFound):
		return openfeature.InterfaceResolutionDetail{
			Value: defaultValue,
			ProviderResolutionDetail: openfeature.ProviderResolutionDetail{
				ResolutionError: openfeature.NewFlagNotFoundResolutionError("error when resolving, flag not found"),
				Reason:          openfeature.ErrorReason}}
	default:
		return openfeature.InterfaceResolutionDetail{
			Value: defaultValue,
			ProviderResolutionDetail: openfeature.ProviderResolutionDetail{
				ResolutionError: openfeature.NewGeneralResolutionError("error when resolving, returning default value"),
				Reason:          openfeature.ErrorReason}}
	}
}

func processResolvedFlag(resolvedFlag resolvedFlag, defaultValue interface{},
	expectedKind reflect.Kind, propertyPath string) openfeature.InterfaceResolutionDetail {
	if len(resolvedFlag.Value) == 0 {
		return openfeature.InterfaceResolutionDetail{
			Value: defaultValue,
			ProviderResolutionDetail: openfeature.ProviderResolutionDetail{
				Reason: openfeature.DefaultReason}}
	}

	actualKind, schemaErr := getTypeForPath(resolvedFlag.FlagSchema.Schema, propertyPath)
	if schemaErr != nil || actualKind != expectedKind {
		return openfeature.InterfaceResolutionDetail{
			Value: defaultValue,
			ProviderResolutionDetail: openfeature.ProviderResolutionDetail{
				ResolutionError: openfeature.NewTypeMismatchResolutionError(
					fmt.Sprintf("schema for property %s does not match the expected type",
						propertyPath)),
				Reason: openfeature.ErrorReason}}
	}

	updatedMap, err := replaceNumbers("", resolvedFlag.Value, resolvedFlag.FlagSchema.Schema)
	if err != nil {
		return typeMismatchError(defaultValue)
	}

	extractedValue, extractValueError := extractPropertyValue(propertyPath, updatedMap)
	if extractValueError != nil {
		return typeMismatchError(defaultValue)
	}

	return openfeature.InterfaceResolutionDetail{
		Value: extractedValue,
		ProviderResolutionDetail: openfeature.ProviderResolutionDetail{
			Reason:  openfeature.TargetingMatchReason,
			Variant: resolvedFlag.Variant}}
}

func replaceNumbers(basePath string, input map[string]interface{},
	schema map[string]interface{}) (map[string]interface{}, error) {
	updatedMap := make(map[string]interface{})
	for key, value := range input {
		kind, typeErr := getTypeForPath(schema, fmt.Sprintf("%s%s", basePath, key))
		if typeErr != nil {
			return updatedMap, fmt.Errorf("unable to get type for path %w", typeErr)
		}

		switch kind {
		case reflect.Float64:
			floatValue, err := value.(json.Number).Float64()
			if err != nil {
				return updatedMap, fmt.Errorf("unable to convert to float")
			}

			updatedMap[key] = floatValue
		case reflect.Int64:
			intValue, err := value.(json.Number).Int64()
			if err != nil {
				return updatedMap, fmt.Errorf("unable to convert to int")
			}

			updatedMap[key] = intValue
		case reflect.Map:
			asMap, ok := value.(map[string]interface{})
			if !ok {
				return updatedMap, fmt.Errorf("unable to convert map")
			}

			childMap, err := replaceNumbers(fmt.Sprintf("%s%s.", basePath, key), asMap, schema)
			if err != nil {
				return updatedMap, fmt.Errorf("unable to convert map")
			}

			updatedMap[key] = childMap
		default:
			updatedMap[key] = value
		}
	}

	return updatedMap, nil
}

func typeMismatchError(defaultValue interface{}) openfeature.InterfaceResolutionDetail {
	return openfeature.InterfaceResolutionDetail{
		Value: defaultValue,
		ProviderResolutionDetail: openfeature.ProviderResolutionDetail{
			ResolutionError: openfeature.NewTypeMismatchResolutionError(
				"Unable to extract property value from resolve response"),
			Reason: openfeature.ErrorReason}}
}

func toBoolResolutionDetail(res openfeature.InterfaceResolutionDetail,
	defaultValue bool) openfeature.BoolResolutionDetail {
	if res.ProviderResolutionDetail.Reason == openfeature.TargetingMatchReason {
		v, ok := res.Value.(bool)
		if ok {
			return openfeature.BoolResolutionDetail{
				Value:                    v,
				ProviderResolutionDetail: res.ProviderResolutionDetail,
			}
		}

		return openfeature.BoolResolutionDetail{
			Value: defaultValue,
			ProviderResolutionDetail: openfeature.ProviderResolutionDetail{
				ResolutionError: openfeature.NewTypeMismatchResolutionError("Unable to convert response property to boolean"),
				Reason:          openfeature.ErrorReason,
			},
		}
	}

	return openfeature.BoolResolutionDetail{
		Value:                    defaultValue,
		ProviderResolutionDetail: res.ProviderResolutionDetail,
	}
}

func toStringResolutionDetail(res openfeature.InterfaceResolutionDetail,
	defaultValue string) openfeature.StringResolutionDetail {
	if res.ProviderResolutionDetail.Reason == openfeature.TargetingMatchReason {
		v, ok := res.Value.(string)
		if ok {
			return openfeature.StringResolutionDetail{
				Value:                    v,
				ProviderResolutionDetail: res.ProviderResolutionDetail,
			}
		}

		return openfeature.StringResolutionDetail{
			Value: defaultValue,
			ProviderResolutionDetail: openfeature.ProviderResolutionDetail{
				ResolutionError: openfeature.NewTypeMismatchResolutionError("Unable to convert response property to boolean"),
				Reason:          openfeature.ErrorReason,
			},
		}
	}

	return openfeature.StringResolutionDetail{
		Value:                    defaultValue,
		ProviderResolutionDetail: res.ProviderResolutionDetail,
	}
}

func toFloatResolutionDetail(res openfeature.InterfaceResolutionDetail,
	defaultValue float64) openfeature.FloatResolutionDetail {
	if res.ProviderResolutionDetail.Reason == openfeature.TargetingMatchReason {
		v, ok := res.Value.(float64)
		if ok {
			return openfeature.FloatResolutionDetail{
				Value:                    v,
				ProviderResolutionDetail: res.ProviderResolutionDetail,
			}
		}

		return openfeature.FloatResolutionDetail{
			Value: defaultValue,
			ProviderResolutionDetail: openfeature.ProviderResolutionDetail{
				ResolutionError: openfeature.NewTypeMismatchResolutionError("Unable to convert response property to float"),
				Reason:          openfeature.ErrorReason,
			},
		}
	}

	return openfeature.FloatResolutionDetail{
		Value:                    defaultValue,
		ProviderResolutionDetail: res.ProviderResolutionDetail,
	}
}

func toIntResolutionDetail(res openfeature.InterfaceResolutionDetail,
	defaultValue int64) openfeature.IntResolutionDetail {
	if res.ProviderResolutionDetail.Reason == openfeature.TargetingMatchReason {
		v, ok := res.Value.(int64)
		if ok {
			return openfeature.IntResolutionDetail{
				Value:                    v,
				ProviderResolutionDetail: res.ProviderResolutionDetail,
			}
		}

		return openfeature.IntResolutionDetail{
			Value: defaultValue,
			ProviderResolutionDetail: openfeature.ProviderResolutionDetail{
				ResolutionError: openfeature.NewTypeMismatchResolutionError("Unable to convert response property to int"),
				Reason:          openfeature.ErrorReason,
			},
		}
	}

	return openfeature.IntResolutionDetail{
		Value:                    defaultValue,
		ProviderResolutionDetail: res.ProviderResolutionDetail,
	}
}
