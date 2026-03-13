package goofishcollector

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
		httpClient: &http.Client{Timeout: 30 * time.Second},
	}
}

func (c *Client) ListStateFiles(ctx context.Context) ([]model.GoofishStateFile, error) {
	response, err := doJSON[struct {
		Count int                      `json:"count"`
		Items []model.GoofishStateFile `json:"items"`
	}](ctx, c.httpClient, http.MethodGet, c.baseURL+"/api/v1/state-files", nil)
	if err != nil {
		return nil, err
	}
	return response.Items, nil
}

func (c *Client) PromoteStateFile(ctx context.Context, fileName string) (model.GoofishStateFile, error) {
	response, err := doJSON[struct {
		Message string                 `json:"message"`
		Item    model.GoofishStateFile `json:"item"`
	}](ctx, c.httpClient, http.MethodPost, c.baseURL+"/api/v1/login-state/default", map[string]string{"file_name": fileName})
	if err != nil {
		return model.GoofishStateFile{}, err
	}
	return response.Item, nil
}

func (c *Client) ValidateState(ctx context.Context, req model.GoofishValidateRequest) (model.GoofishValidateResponse, error) {
	return doJSON[model.GoofishValidateResponse](ctx, c.httpClient, http.MethodPost, c.baseURL+"/api/v1/validate-state", req)
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
		return zero, fmt.Errorf("upstream goofish-collector returned %d", resp.StatusCode)
	}
	if err := json.NewDecoder(resp.Body).Decode(&zero); err != nil {
		return zero, fmt.Errorf("decode response: %w", err)
	}
	return zero, nil
}
