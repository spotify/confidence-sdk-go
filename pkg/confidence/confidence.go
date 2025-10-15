package confidence

import (
	"context"
	"fmt"
	"reflect"
	"strings"
	"sync"
	"time"

	"golang.org/x/exp/slog"
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
	SDK_VERSION = "0.4.6" // x-release-please-version
)

type Confidence struct {
	parent        ContextProvider
	EventUploader EventUploader
	contextMap    map[string]interface{}
	Config        APIConfig
	ResolveClient ResolveClient
	Logger        *slog.Logger
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
		if value == nil {
			delete(currentMap, key)
		} else {
			currentMap[key] = value
		}
	}
	return currentMap
}

type ConfidenceBuilder struct {
	confidence Confidence
}

func (e ConfidenceBuilder) SetLogger(logger *slog.Logger) ConfidenceBuilder {
	e.confidence.Logger = logger
	return e
}

func (e ConfidenceBuilder) SetAPIConfig(config APIConfig) ConfidenceBuilder {
	e.confidence.Config = config
	if config.APIResolveBaseUrl == "" {
		e.confidence.Config.APIResolveBaseUrl = DefaultAPIResolveBaseUrl
	}
	return e
}

func (e ConfidenceBuilder) SetResolveClient(client ResolveClient) ConfidenceBuilder {
	e.confidence.ResolveClient = client
	return e
}

func (e ConfidenceBuilder) Build() Confidence {
	if e.confidence.Logger == nil {
		e.confidence.Logger = slog.Default()
	}
	if e.confidence.ResolveClient == nil {
		e.confidence.ResolveClient = NewHttpResolveClient(e.confidence.Config)
	}
	if e.confidence.EventUploader == nil {
		e.confidence.EventUploader = NewHttpEventUploader(e.confidence.Config, e.confidence.Logger)
	}

	e.confidence.contextMap = make(map[string]interface{})
	e.confidence.Logger.Info("Confidence created", "config", e.confidence.Config)
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

func (e Confidence) Track(ctx context.Context, eventName string, data map[string]interface{}) *sync.WaitGroup {
	newMap := make(map[string]interface{})
	newMap["context"] = e.GetContext()

	for key, value := range data {
		if key == "context" {
			panic("invalid key \"context\" inside the data")
		}
		newMap[key] = value
	}

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		currentTime := time.Now()
		iso8601Time := currentTime.Format(time.RFC3339)
		event := Event{
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
		e.Logger.Debug("EventUploading started", "eventName", eventName)
		e.EventUploader.upload(ctx, batch)
		wg.Done()
		e.Logger.Debug("EventUploading completed", "eventName", eventName)
	}()
	return &wg
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
		Logger:        e.Logger,
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

func (e Confidence) GetObjectFlag(ctx context.Context, flag string, defaultValue map[string]interface{}) InterfaceResolutionDetail {
	resp := e.ResolveFlag(ctx, flag, defaultValue, reflect.Map)
	return resp
}

func (e Confidence) GetObjectValue(ctx context.Context, flag string, defaultValue map[string]interface{}) interface{} {
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
		slog.Warn("Error in resolving flag", "flag", flag, "error", err)
		return processResolveError(err, defaultValue)
	}
	logResolveTesterHint(e.Logger, flagName, e.Config.APIKey, e.contextMap)

	if len(resp.ResolvedFlags) == 0 {
		slog.Debug("Flag not found", "flag", flag)
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
		slog.Warn("Unexpected flag from remote", "flag", resolvedFlag.Flag)
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
