package provider

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type httpResolveClient struct {
	Client *http.Client
	Config APIConfig
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

func (client httpResolveClient) sendResolveRequest(ctx context.Context,
	request resolveRequest) (resolveResponse, error) {
	jsonRequest, err := json.Marshal(request)
	if err != nil {
		return resolveResponse{}, fmt.Errorf("error when serializing request to the resolver service: %w", err)
	}

	payload := bytes.NewBuffer(jsonRequest)
	req, err := http.NewRequestWithContext(ctx,
		http.MethodPost, fmt.Sprintf("%s/flags:resolve", client.Config.Region.apiURL()), payload)
	if err != nil {
		return resolveResponse{}, err
	}

	resp, err := client.Client.Do(req)
	if err != nil {
		return resolveResponse{}, fmt.Errorf("error when calling the resolver service: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return resolveResponse{},
			fmt.Errorf("got '%s' error from the resolver service: %s", resp.Status, parseErrorMessage(resp.Body))
	}

	var result resolveResponse
	decoder := json.NewDecoder(resp.Body)
	decoder.UseNumber()
	err = decoder.Decode(&result)
	if err != nil {
		return resolveResponse{}, fmt.Errorf("error when deserializing response from the resolver service: %w", err)
	}
	return result, nil
}
