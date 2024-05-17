package confidence

import (
	"context"
	"fmt"
	"time"
	"net/http"
	"reflect"
	"strings"
)

type FlagResolver interface {
	resolveFlag(ctx context.Context, flag string, defaultValue interface{},
		evalCtx map[string]interface{}, expectedKind reflect.Kind) InterfaceResolutionDetail
}

type ContextProvider interface {
	GetContext() map[string]interface{}
}

var (
	SDK_ID      = "SDK_ID_GO_CONFIDENCE"
	SDK_VERSION = "0.1.8" // x-release-please-version
)

type Confidence struct {
	parent        ContextProvider
	uploader      EventUploader
	contextMap    map[string]interface{}
	Config        APIConfig
	ResolveClient ResolveClient
}

func (e Confidence) GetContext() map[string]interface{} {
	currentMap := map[string]interface{}{}
	parentMap := make(map[string]interface{})
	if e.parent != nil {
		parentMap = e.parent.GetContext()
	}
for key, value := range parentMap {
		currentMap[key] = value
	}
	for key, value := range e.contextMap {
		currentMap[key] = value
	}
	return currentMap
}

type ConfidenceBuilder struct {
	confidence Confidence
}

func (e ConfidenceBuilder) SetAPIConfig(config APIConfig) ConfidenceBuilder {
	e.confidence.Config = config
	return e
}

func (e ConfidenceBuilder) SetResolveClient(client ResolveClient) ConfidenceBuilder {
	e.confidence.ResolveClient = client
	return e
}

func (e ConfidenceBuilder) Build() Confidence {
	if e.confidence.ResolveClient == nil {
		e.confidence.ResolveClient = HttpResolveClient{Client: &http.Client{}, Config: e.confidence.Config}
	}
	e.confidence.contextMap = make(map[string]interface{})
	return e.confidence
}

func NewConfidenceBuilder() ConfidenceBuilder {
	return ConfidenceBuilder{
		confidence: Confidence{},
	}
}

func (e Confidence) PutContext(key string, value interface{}) {
	e.contextMap[key] = value
}

func (e Confidence) Track(ctx context.Context, eventName string, message map[string]interface{}) {
	newMap := e.GetContext()

	for key, value := range message {
		newMap[key] = value
	}

	go func() {
		currentTime := time.Now()
		iso8601Time := currentTime.Format(time.RFC3339)
		event := Event {
			EventDefinition: fmt.Sprintf("eventDefinitions/%s", eventName),
			EventTime:       iso8601Time,
			Payload:         newMap,
		}
		batch := EventBatchRequest{
			CclientSecret: e.Config.APIKey,
			Sdk:           sdk{SDK_ID, SDK_VERSION},
			SendTime:      iso8601Time,
			Events:        []Event{event},
		}
		e.uploader.upload(ctx, batch)
	}()
}

func (e Confidence) WithContext(context map[string]interface{}) Confidence {
	newMap := map[string]interface{}{}
	for key, value := range e.GetContext() {
		newMap[key] = value
	}

	for key, value := range context {
		newMap[key] = value
	}

	return Confidence{
		parent:        &e,
		contextMap:    newMap,
		Config:        e.Config,
		ResolveClient: e.ResolveClient,
	}
}

func (e Confidence) GetBoolFlag(ctx context.Context, flag string, defaultValue bool) BoolResolutionDetail {
	resp := e.ResolveFlag(ctx, flag, defaultValue, reflect.Bool)
	return ToBoolResolutionDetail(resp, defaultValue)
}

func (e Confidence) GetBoolValue(ctx context.Context, flag string, defaultValue bool) bool {
	return e.GetBoolFlag(ctx, flag, defaultValue).Value
}

func (e Confidence) GetIntFlag(ctx context.Context, flag string, defaultValue int64) IntResolutionDetail {
	resp := e.ResolveFlag(ctx, flag, defaultValue, reflect.Int64)
	return ToIntResolutionDetail(resp, defaultValue)
}

func (e Confidence) GetIntValue(ctx context.Context, flag string, defaultValue int64) int64 {
	return e.GetIntFlag(ctx, flag, defaultValue).Value
}

func (e Confidence) GetDoubleFlag(ctx context.Context, flag string, defaultValue float64) FloatResolutionDetail {
	resp := e.ResolveFlag(ctx, flag, defaultValue, reflect.Float64)
	return ToFloatResolutionDetail(resp, defaultValue)
}

func (e Confidence) GetDoubleValue(ctx context.Context, flag string, defaultValue float64) float64 {
	return e.GetDoubleFlag(ctx, flag, defaultValue).Value
}

func (e Confidence) GetStringFlag(ctx context.Context, flag string, defaultValue string) StringResolutionDetail {
	resp := e.ResolveFlag(ctx, flag, defaultValue, reflect.String)
	return ToStringResolutionDetail(resp, defaultValue)
}

func (e Confidence) GetStringValue(ctx context.Context, flag string, defaultValue string) string {
	return e.GetStringFlag(ctx, flag, defaultValue).Value
}

func (e Confidence) GetObjectFlag(ctx context.Context, flag string, defaultValue string) InterfaceResolutionDetail {
	resp := e.ResolveFlag(ctx, flag, defaultValue, reflect.Map)
	return resp
}

func (e Confidence) GetObjectValue(ctx context.Context, flag string, defaultValue string) interface{} {
	return e.GetObjectFlag(ctx, flag, defaultValue).Value
}

func (e Confidence) ResolveFlag(ctx context.Context, flag string, defaultValue interface{}, expectedKind reflect.Kind) InterfaceResolutionDetail {
	flagName, propertyPath := splitFlagString(flag)

	requestFlagName := fmt.Sprintf("flags/%s", flagName)
	resp, err := e.ResolveClient.SendResolveRequest(ctx,
		ResolveRequest{ClientSecret: e.Config.APIKey,
			Flags: []string{requestFlagName}, Apply: true, EvaluationContext: e.contextMap,
			Sdk: sdk{Id: SDK_ID, Version: SDK_VERSION}})

	if err != nil {
		return processResolveError(err, defaultValue)
	}
	if len(resp.ResolvedFlags) == 0 {
		return InterfaceResolutionDetail{
			Value: defaultValue,
			ResolutionDetail: ResolutionDetail{
				Variant:      "",
				Reason:       ErrorReason,
				ErrorCode:    FlagNotFoundCode,
				ErrorMessage: "Flag not found",
				FlagMetadata: nil,
			},
		}
	}

	resolvedFlag := resp.ResolvedFlags[0]
	if resolvedFlag.Flag != requestFlagName {
		return InterfaceResolutionDetail{
			Value: defaultValue,
			ResolutionDetail: ResolutionDetail{
				Variant:      "",
				Reason:       ErrorReason,
				ErrorCode:    FlagNotFoundCode,
				ErrorMessage: fmt.Sprintf("unexpected flag '%s' from remote", strings.TrimPrefix(resolvedFlag.Flag, "flags/")),
				FlagMetadata: nil,
			},
		}
	}

	return processResolvedFlag(resolvedFlag, defaultValue, expectedKind, propertyPath)
}
