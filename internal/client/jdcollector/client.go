package jdcollector

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/rigel-labs/rigel-console/internal/domain/model"
)

type Client struct {
	baseURL    string
	httpClient *http.Client
}

func New(baseURL string) *Client {
	return &Client{
		baseURL:    strings.TrimRight(baseURL, "/"),
		httpClient: &http.Client{Timeout: 15 * time.Second},
	}
}

func (c *Client) GetScheduleConfig(ctx context.Context) (model.CollectorScheduleResponse, error) {
	return doJSON[model.CollectorScheduleResponse](ctx, c.httpClient, http.MethodGet, c.baseURL+"/api/v1/admin/schedule", nil)
}

func (c *Client) UpdateScheduleConfig(ctx context.Context, payload model.CollectorScheduleUpsertRequest) (model.CollectorScheduleResponse, error) {
	return doJSON[model.CollectorScheduleResponse](ctx, c.httpClient, http.MethodPut, c.baseURL+"/api/v1/admin/schedule", payload)
}

func doJSON[T any](ctx context.Context, httpClient *http.Client, method, target string, payload any) (T, error) {
	var zero T
	var body *bytes.Reader
	if payload == nil {
		body = bytes.NewReader(nil)
	} else {
		encoded, err := json.Marshal(payload)
		if err != nil {
			return zero, fmt.Errorf("marshal request: %w", err)
		}
		body = bytes.NewReader(encoded)
	}

	req, err := http.NewRequestWithContext(ctx, method, target, body)
	if err != nil {
		return zero, fmt.Errorf("new request: %w", err)
	}
	if payload != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := httpClient.Do(req)
	if err != nil {
		return zero, fmt.Errorf("do request: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode >= http.StatusBadRequest {
		return zero, fmt.Errorf("upstream jd-collector returned %d", resp.StatusCode)
	}
	if err := json.NewDecoder(resp.Body).Decode(&zero); err != nil {
		return zero, fmt.Errorf("decode response: %w", err)
	}
	return zero, nil
}
