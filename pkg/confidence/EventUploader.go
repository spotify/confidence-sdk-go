package confidence

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
)

type EventUploader interface {
	upload(ctx context.Context, request EventBatchRequest)
}

type HttpEventUploader struct {
	Client *http.Client
	Config APIConfig
	Logger *slog.Logger
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
		e.Logger.Warn("Failed to perform upload request", "error", err)
		return
	}
	if resp.StatusCode != http.StatusOK {
		e.Logger.Warn("Failed to upload event", "status", resp.Status)
	}
	defer resp.Body.Close()
}
