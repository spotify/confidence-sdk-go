package confidence

import (
	"fmt"
)

type ResolutionError struct {
	// fields are unexported, this means providers are forced to create structs of this type using one of the constructors below.
	// this effectively emulates an enum
	code    ErrorCode
	message string
}

func (r ResolutionError) Error() string {
	return fmt.Sprintf("%s: %s", r.code, r.message)
}

type ResolutionDetail struct {
	Variant      string
	Reason       Reason
	ErrorCode    ErrorCode
	ErrorMessage string
	FlagMetadata FlagMetadata
}

// FlagMetadata is a structure which supports definition of arbitrary properties, with keys of type string, and values
// of type boolean, string, int64 or float64. This structure is populated by a provider for use by an Application
// Author (via the Evaluation API) or an Application Integrator (via hooks).
type FlagMetadata map[string]interface{}
type Reason string

type ErrorCode string

const (
	// ProviderNotReadyCode - the value was resolved before the provider was ready.
	ProviderNotReadyCode ErrorCode = "PROVIDER_NOT_READY"
	// FlagNotFoundCode - the flag could not be found.
	FlagNotFoundCode ErrorCode = "FLAG_NOT_FOUND"
	// ParseErrorCode - an error was encountered parsing data, such as a flag configuration.
	ParseErrorCode ErrorCode = "PARSE_ERROR"
	// TypeMismatchCode - the type of the flag value does not match the expected type.
	TypeMismatchCode ErrorCode = "TYPE_MISMATCH"
	// TargetingKeyMissingCode - the provider requires a targeting key and one was not provided in the evaluation context.
	TargetingKeyMissingCode ErrorCode = "TARGETING_KEY_MISSING"
	// InvalidContextCode - the evaluation context does not meet provider requirements.
	InvalidContextCode ErrorCode = "INVALID_CONTEXT"
	// TimeoutCode - the request timed out.
	TimeoutCode ErrorCode = "TIMEOUT"
	// GeneralCode - the error was for a reason not enumerated above.
	GeneralCode ErrorCode = "GENERAL"
)

// BoolResolutionDetail provides a resolution detail with boolean type
type BoolResolutionDetail struct {
	Value bool
	ResolutionDetail
}

// StringResolutionDetail provides a resolution detail with string type
type StringResolutionDetail struct {
	Value string
	ResolutionDetail
}

// FloatResolutionDetail provides a resolution detail with float64 type
type FloatResolutionDetail struct {
	Value float64
	ResolutionDetail
}

// IntResolutionDetail provides a resolution detail with int64 type
type IntResolutionDetail struct {
	Value int64
	ResolutionDetail
}

// InterfaceResolutionDetail provides a resolution detail with interface{} type
type InterfaceResolutionDetail struct {
	Value interface{}
	ResolutionDetail
}

// Metadata provides provider name
type Metadata struct {
	Name string
}
