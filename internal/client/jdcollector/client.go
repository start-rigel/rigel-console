package jdcollector

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
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
		httpClient: &http.Client{Timeout: 2 * time.Minute},
	}
}

func (c *Client) ListProducts(ctx context.Context, filter model.AdminProductFilter) ([]model.AdminProduct, error) {
	query := url.Values{}
	query.Set("keyword", filter.Keyword)
	if filter.Limit > 0 {
		query.Set("limit", fmt.Sprintf("%d", filter.Limit))
	}
	if filter.ShopType != "" {
		query.Set("shop_type", filter.ShopType)
	}
	if filter.RealOnly {
		query.Set("real_only", "true")
	}
	if filter.SelfOperatedOnly {
		query.Set("self_operated_only", "true")
	}
	response, err := doJSON[struct {
		Count int                  `json:"count"`
		Items []model.AdminProduct `json:"items"`
	}](ctx, c.httpClient, http.MethodGet, c.baseURL+"/api/v1/products?"+query.Encode(), nil)
	if err != nil {
		return nil, err
	}
	return response.Items, nil
}

func (c *Client) ListJobs(ctx context.Context, limit int) ([]model.AdminJob, error) {
	query := url.Values{}
	if limit > 0 {
		query.Set("limit", fmt.Sprintf("%d", limit))
	}
	response, err := doJSON[struct {
		Count int              `json:"count"`
		Items []model.AdminJob `json:"items"`
	}](ctx, c.httpClient, http.MethodGet, c.baseURL+"/api/v1/jobs?"+query.Encode(), nil)
	if err != nil {
		return nil, err
	}
	return response.Items, nil
}

func (c *Client) TriggerCollection(ctx context.Context, req model.AdminCollectRequest) (model.AdminCollectResponse, error) {
	return doJSON[model.AdminCollectResponse](ctx, c.httpClient, http.MethodPost, c.baseURL+"/api/v1/collect/search", req)
}

func (c *Client) TriggerBatchCollection(ctx context.Context, req model.AdminCollectBatchRequest) (model.AdminCollectBatchResponse, error) {
	return doJSON[model.AdminCollectBatchResponse](ctx, c.httpClient, http.MethodPost, c.baseURL+"/api/v1/collect/batch", req)
}

func (c *Client) RetryJob(ctx context.Context, jobID string) (model.AdminCollectResponse, error) {
	return doJSON[model.AdminCollectResponse](ctx, c.httpClient, http.MethodPost, c.baseURL+"/api/v1/jobs/"+jobID+"/retry", nil)
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
