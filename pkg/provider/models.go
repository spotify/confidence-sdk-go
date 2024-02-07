package provider

import (
	"context"
	"errors"
)

type APIRegion int64

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

func (c APIConfig) validate() error {
	if c.APIKey == "" {
		return errors.New("api key needs to be set")
	}
	if c.Region.apiURL() == "" {
		return errors.New("api region needs to be set")
	}
	return nil
}

type resolveClient interface {
	sendResolveRequest(ctx context.Context, request resolveRequest) (resolveResponse, error)
}

var errFlagNotFound = errors.New("flag not found")

type resolveRequest struct {
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

type resolveResponse struct {
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
