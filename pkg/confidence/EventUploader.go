package confidence

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
)

type EventUploader interface {
	upload(ctx context.Context, request EventBatchRequest)
}

type HttpEventUploader struct {
	Client *http.Client
	Config APIConfig
}

func (e HttpEventUploader) upload(ctx context.Context, request EventBatchRequest) {
	jsonRequest, err := json.Marshal(request)
	if err != nil {
		return
	}

	payload := bytes.NewBuffer(jsonRequest)
	req, err := http.NewRequestWithContext(ctx,
		http.MethodPost, "https://events.eu.confidence.dev/v1/events:publish", payload)
	if err != nil {
		return
	}

	resp, err := e.Client.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()
}
