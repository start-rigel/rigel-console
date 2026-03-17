package buildengine

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
	return &Client{baseURL: strings.TrimRight(baseURL, "/"), httpClient: &http.Client{Timeout: 15 * time.Second}}
}

func (c *Client) GenerateBuild(ctx context.Context, req model.GenerateBuildRequest) (model.BuildEngineResponse, error) {
	return doJSON[model.BuildEngineResponse](ctx, c.httpClient, http.MethodPost, c.baseURL+"/api/v1/builds/generate", req)
}

func (c *Client) GetBuild(ctx context.Context, buildID string) (model.BuildEngineResponse, error) {
	return doJSON[model.BuildEngineResponse](ctx, c.httpClient, http.MethodGet, c.baseURL+"/api/v1/builds/"+buildID, nil)
}

func (c *Client) SearchParts(ctx context.Context, keyword string, limit int) ([]model.PartSearchResult, error) {
	query := url.Values{}
	query.Set("keyword", keyword)
	if limit > 0 {
		query.Set("limit", fmt.Sprintf("%d", limit))
	}
	response, err := doJSON[struct {
		Count int                      `json:"count"`
		Items []model.PartSearchResult `json:"items"`
	}](ctx, c.httpClient, http.MethodGet, c.baseURL+"/api/v1/parts/search?"+query.Encode(), nil)
	if err != nil {
		return nil, err
	}
	return response.Items, nil
}

func (c *Client) GetPriceCatalog(ctx context.Context, req model.GenerateBuildRequest) (model.BuildEnginePriceCatalog, error) {
	query := url.Values{}
	if req.UseCase != "" {
		query.Set("use_case", req.UseCase)
	}
	if req.BuildMode != "" {
		query.Set("build_mode", req.BuildMode)
	}
	query.Set("limit", "500")
	return doJSON[model.BuildEnginePriceCatalog](ctx, c.httpClient, http.MethodGet, c.baseURL+"/api/v1/catalog/prices?"+query.Encode(), nil)
}

func (c *Client) GenerateAdvice(ctx context.Context, build model.BuildEngineResponse) (model.AdviceResponse, error) {
	payload := map[string]any{"build": build}
	return doJSON[model.AdviceResponse](ctx, c.httpClient, http.MethodPost, c.baseURL+"/api/v1/advice/generate", payload)
}

func (c *Client) GenerateCatalogAdvice(ctx context.Context, req model.GenerateBuildRequest, catalog model.BuildEnginePriceCatalog) (model.CatalogAdviceResponse, error) {
	payload := map[string]any{
		"budget":     req.Budget,
		"use_case":   req.UseCase,
		"build_mode": req.BuildMode,
		"catalog":    catalog,
	}
	return doJSON[model.CatalogAdviceResponse](ctx, c.httpClient, http.MethodPost, c.baseURL+"/api/v1/advice/catalog", payload)
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
		return zero, fmt.Errorf("upstream build-engine returned %d", resp.StatusCode)
	}
	if err := json.NewDecoder(resp.Body).Decode(&zero); err != nil {
		return zero, fmt.Errorf("decode response: %w", err)
	}
	return zero, nil
}
