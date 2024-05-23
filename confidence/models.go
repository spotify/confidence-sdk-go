package confidence

import (
	"context"
	"errors"
)

type APIRegion int64

func NewFlagNotFoundResolutionError(msg string) ResolutionError {
	return ResolutionError{
		code:    FlagNotFoundCode,
		message: msg,
	}
}

func NewParseErrorResolutionError(msg string) ResolutionError {
	return ResolutionError{
		code:    ParseErrorCode,
		message: msg,
	}
}

// NewTypeMismatchResolutionError constructs a resolution error with code TYPE_MISMATCH
//
// Explanation - The type of the flag value does not match the expected type.
func NewTypeMismatchResolutionError(msg string) ResolutionError {
	return ResolutionError{
		code:    TypeMismatchCode,
		message: msg,
	}
}

// NewTargetingKeyMissingResolutionError constructs a resolution error with code TARGETING_KEY_MISSING
//
// Explanation - The provider requires a targeting key and one was not provided in the evaluation context.
func NewTargetingKeyMissingResolutionError(msg string) ResolutionError {
	return ResolutionError{
		code:    TargetingKeyMissingCode,
		message: msg,
	}
}

func NewInvalidContextResolutionError(msg string) ResolutionError {
	return ResolutionError{
		code:    InvalidContextCode,
		message: msg,
	}
}

// NewGeneralResolutionError constructs a resolution error with code GENERAL
//
// Explanation - The error was for a reason not enumerated above.
func NewGeneralResolutionError(msg string) ResolutionError {
	return ResolutionError{
		code:    GeneralCode,
		message: msg,
	}
}

type APIConfig struct {
	APIKey string
	Region APIRegion
}

func NewAPIConfig(apiKey string) *APIConfig {
	return &APIConfig{
		APIKey: apiKey,
		Region: APIRegionGlobal,
	}
}

const (
	APIRegionEU     = iota
	APIRegionUS     = iota
	APIRegionGlobal = iota
)

// Private types below

const euAPIURL = "https://resolver.eu.confidence.dev/v1"
const usAPIURL = "https://resolver.us.confidence.dev/v1"
const globalAPIURL = "https://resolver.confidence.dev/v1"

func (r APIRegion) apiURL() string {
	if r == APIRegionEU {
		return euAPIURL
	} else if r == APIRegionUS {
		return usAPIURL
	} else if r == APIRegionGlobal {
		return globalAPIURL
	}
	return ""
}

func (c APIConfig) Validate() error {
	if c.APIKey == "" {
		return errors.New("api key needs to be set")
	}
	if c.Region.apiURL() == "" {
		return errors.New("api region needs to be set")
	}
	return nil
}

type ResolveClient interface {
	SendResolveRequest(ctx context.Context, request ResolveRequest) (ResolveResponse, error)
}

var errFlagNotFound = errors.New("flag not found")

type EventBatchRequest struct {
	CclientSecret string  `json:"clientSecret"`
	Sdk           sdk     `json:"sdk"`
	SendTime      string  `json:"sendTime"`
	Events        []Event `json:"events"`
}

type Event struct {
	EventDefinition string                 `json:"eventDefinition"`
	EventTime       string                 `json:"eventTime"`
	Payload         map[string]interface{} `json:"payload"`
}

type ResolveRequest struct {
	ClientSecret      string                 `json:"client_secret"`
	Apply             bool                   `json:"apply"`
	EvaluationContext map[string]interface{} `json:"evaluation_context"`
	Flags             []string               `json:"flags"`
	Sdk               sdk                    `json:"sdk"`
}

type sdk struct {
	Id      string `json:"id"`
	Version string `json:"version"`
}

type ResolveResponse struct {
	ResolvedFlags []resolvedFlag `json:"resolvedFlags"`
	ResolveToken  string         `json:"resolveToken"`
}

type resolveErrorMessage struct {
	Code    int64  `json:"code"`
	Message string `json:"message"`
}

type resolvedFlag struct {
	Flag       string                 `json:"flag"`
	Variant    string                 `json:"variant"`
	Reason     string                 `json:"reason"`
	Value      map[string]interface{} `json:"value"`
	FlagSchema flagSchema             `json:"flagSchema"`
}

type flagSchema struct {
	Schema map[string]interface{} `json:"schema"`
}
