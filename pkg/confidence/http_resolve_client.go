package confidence

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type HttpResolveClient struct {
	Client *http.Client
	Config APIConfig
}

func NewHttpResolveClient(config APIConfig) HttpResolveClient {
	return HttpResolveClient{
		Client: &http.Client{
			Timeout: config.ResolveTimeout,
		},
		Config: config,
	}
}

func parseErrorMessage(body io.ReadCloser) string {
	var resolveError resolveErrorMessage
	decoder := json.NewDecoder(body)
	decoder.UseNumber()
	err := decoder.Decode(&resolveError)
	if err != nil {
		return ""
	}
	return resolveError.Message
}

func (client HttpResolveClient) SendResolveRequest(ctx context.Context,
	request ResolveRequest) (ResolveResponse, error) {
	jsonRequest, err := json.Marshal(request)
	if err != nil {
		return ResolveResponse{}, fmt.Errorf("error when serializing request to the resolver service: %w", err)
	}

	payload := bytes.NewBuffer(jsonRequest)
	req, err := http.NewRequestWithContext(ctx,
		http.MethodPost, fmt.Sprintf("%s/v1/flags:resolve", client.Config.APIResolveBaseUrl), payload)
	if err != nil {
		return ResolveResponse{}, err
	}

	resp, err := client.Client.Do(req)
	if err != nil {
		return ResolveResponse{}, fmt.Errorf("error when calling the resolver service: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return ResolveResponse{},
			fmt.Errorf("got '%s' error from the resolver service: %s", resp.Status, parseErrorMessage(resp.Body))
	}

	var result ResolveResponse
	decoder := json.NewDecoder(resp.Body)
	decoder.UseNumber()
	err = decoder.Decode(&result)
	if err != nil {
		return ResolveResponse{}, fmt.Errorf("error when deserializing response from the resolver service: %w", err)
	}
	return result, nil
}
