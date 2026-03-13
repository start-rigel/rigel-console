package aiadvisor

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
	return &Client{baseURL: strings.TrimRight(baseURL, "/"), httpClient: &http.Client{Timeout: 15 * time.Second}}
}

func (c *Client) GenerateAdvice(ctx context.Context, build model.BuildEngineResponse) (model.AIAdvisorResponse, error) {
	payload := map[string]any{"build": build}
	encoded, err := json.Marshal(payload)
	if err != nil {
		return model.AIAdvisorResponse{}, fmt.Errorf("marshal request: %w", err)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/api/v1/advice/generate", bytes.NewReader(encoded))
	if err != nil {
		return model.AIAdvisorResponse{}, fmt.Errorf("new request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return model.AIAdvisorResponse{}, fmt.Errorf("do request: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode >= http.StatusBadRequest {
		return model.AIAdvisorResponse{}, fmt.Errorf("upstream ai-advisor returned %d", resp.StatusCode)
	}
	var response model.AIAdvisorResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return model.AIAdvisorResponse{}, fmt.Errorf("decode response: %w", err)
	}
	return response, nil
}
