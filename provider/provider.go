package provider

import (
	"context"
	"github.com/open-feature/go-sdk/openfeature"
	c "github.com/spotify/confidence-openfeature-provider-go/confidence"
	"reflect"
)

type FlagProvider struct {
	confidence c.Confidence
}

func NewFlagProvider(confidence c.Confidence) *FlagProvider {
	return &FlagProvider{
		confidence: confidence,
	}
}

func (e FlagProvider) Metadata() openfeature.Metadata {
	return openfeature.Metadata{Name: "ConfidenceFlagProvider"}
}

func (e FlagProvider) BooleanEvaluation(ctx context.Context, flag string, defaultValue bool,
	evalCtx openfeature.FlattenedContext) openfeature.BoolResolutionDetail {
	confidence := e.confidence.WithContext(processTargetingKey(evalCtx))
	res := confidence.ResolveFlag(ctx, flag, defaultValue, reflect.Bool)
	boolDetail := c.ToBoolResolutionDetail(res, defaultValue)
	return openfeature.BoolResolutionDetail{
		Value:                    boolDetail.Value,
		ProviderResolutionDetail: toOFResolutionDetail(boolDetail.ResolutionDetail),
	}
}

func (e FlagProvider) StringEvaluation(ctx context.Context, flag string, defaultValue string,
	evalCtx openfeature.FlattenedContext) openfeature.StringResolutionDetail {
	confidence := e.confidence.WithContext(processTargetingKey(evalCtx))
	res := confidence.ResolveFlag(ctx, flag, defaultValue, reflect.String)
	detail := c.ToStringResolutionDetail(res, defaultValue)
	return openfeature.StringResolutionDetail{
		Value:                    detail.Value,
		ProviderResolutionDetail: toOFResolutionDetail(detail.ResolutionDetail),
	}
}



func (e FlagProvider) FloatEvaluation(ctx context.Context, flag string, defaultValue float64,
	evalCtx openfeature.FlattenedContext) openfeature.FloatResolutionDetail {
	confidence := e.confidence.WithContext(processTargetingKey(evalCtx))
	res := confidence.ResolveFlag(ctx, flag, defaultValue, reflect.Float64)
	detail := c.ToFloatResolutionDetail(res, defaultValue)
	return openfeature.FloatResolutionDetail{
		Value:                    detail.Value,
		ProviderResolutionDetail: toOFResolutionDetail(detail.ResolutionDetail),
	}
}

func (e FlagProvider) IntEvaluation(ctx context.Context, flag string, defaultValue int64,
	evalCtx openfeature.FlattenedContext) openfeature.IntResolutionDetail {
	confidence := e.confidence.WithContext(processTargetingKey(evalCtx))
	res := confidence.ResolveFlag(ctx, flag, defaultValue, reflect.Int64)
	detail := c.ToIntResolutionDetail(res, defaultValue)
	return openfeature.IntResolutionDetail{
		Value:                    detail.Value,
		ProviderResolutionDetail: toOFResolutionDetail(detail.ResolutionDetail),
	}
}

func (e FlagProvider) ObjectEvaluation(ctx context.Context, flag string, defaultValue interface{},
	evalCtx openfeature.FlattenedContext) openfeature.InterfaceResolutionDetail {
	confidence := e.confidence.WithContext(processTargetingKey(evalCtx))
	res := confidence.ResolveFlag(ctx, flag, defaultValue, reflect.Map)
	detail := c.ToObjectResolutionDetail(res, defaultValue)
	return openfeature.InterfaceResolutionDetail{
		Value:                    detail.Value,
		ProviderResolutionDetail: toOFResolutionDetail(detail.ResolutionDetail),
	}
}

func (e FlagProvider) Hooks() []openfeature.Hook {
	return []openfeature.Hook{}
}

func toOFResolutionDetail(detail c.ResolutionDetail) openfeature.ProviderResolutionDetail {
	return openfeature.ProviderResolutionDetail{
		ResolutionError: toOFResolutionError(detail.ErrorCode, detail.ErrorMessage),
		Reason:          toOFReason(detail.Reason),
		Variant:         detail.Variant,
		FlagMetadata:    toOFFlagMetadata(detail.FlagMetadata),
	}
}

func toOFResolutionError(code c.ErrorCode, message string) openfeature.ResolutionError {
	switch code {
	case c.TypeMismatchCode:
		return openfeature.NewTypeMismatchResolutionError(message)
	case c.FlagNotFoundCode:
		return openfeature.NewFlagNotFoundResolutionError(message)
	case c.GeneralCode:
		return openfeature.NewGeneralResolutionError(message)
	case c.InvalidContextCode:
		return openfeature.NewInvalidContextResolutionError(message)
	case c.ProviderNotReadyCode:
		return openfeature.NewProviderNotReadyResolutionError(message)
	case c.ParseErrorCode:
		return openfeature.NewParseErrorResolutionError(message)
	}
	return openfeature.ResolutionError{}
}

func processTargetingKey(evalCtx openfeature.FlattenedContext) openfeature.FlattenedContext {
	newEvalContext := openfeature.FlattenedContext{}
	newEvalContext = evalCtx
	if targetingKey, exists := evalCtx["targetingKey"]; exists {
		newEvalContext["targeting_key"] = targetingKey
	}
	delete(newEvalContext, "targetingKey")
	return newEvalContext
}

func toOFFlagMetadata(metadata c.FlagMetadata) map[string]interface{} {
	return metadata
}

func toOFReason(reason c.Reason) openfeature.Reason {
	switch reason {
	case c.TargetingMatchReason:
		return openfeature.TargetingMatchReason
	case c.DefaultReason:
		return openfeature.DefaultReason
	default:
		return openfeature.ErrorReason
	}
}
