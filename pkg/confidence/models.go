package confidence

import (
	"context"
	"errors"
)

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

const DefaultAPIResolveBaseUrl = "https://resolver.confidence.dev"

type APIConfig struct {
	APIKey            string
	APIResolveBaseUrl string
}

func NewAPIConfig(apiKey string) *APIConfig {
	return &APIConfig{
		APIKey:            apiKey,
		APIResolveBaseUrl: DefaultAPIResolveBaseUrl,
	}
}

func NewAPIConfigWithUrl(apiKey, APIResolveBaseUrl string) *APIConfig {
	return &APIConfig{
		APIKey:            apiKey,
		APIResolveBaseUrl: APIResolveBaseUrl,
	}
}

func (c APIConfig) Validate() error {
	if c.APIKey == "" {
		return errors.New("api key needs to be set")
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
