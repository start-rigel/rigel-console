package buildengine

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
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

func (c *Client) RecommendBuild(ctx context.Context, req model.GenerateBuildRequest) (model.CatalogRecommendationResponse, error) {
	payload := map[string]any{
		"budget":     req.Budget,
		"use_case":   req.UseCase,
		"build_mode": req.BuildMode,
		"notes":      req.Notes,
	}
	return doJSON[model.CatalogRecommendationResponse](ctx, c.httpClient, http.MethodPost, c.baseURL+"/api/v1/recommend/build", payload)
}

func (c *Client) GenerateCatalogAdvice(ctx context.Context, req model.GenerateBuildRequest, catalog model.BuildEnginePriceCatalog) (model.CatalogAdviceResponse, error) {
	payload := map[string]any{
		"budget":               req.Budget,
		"use_case":             req.UseCase,
		"build_mode":           req.BuildMode,
		"brand_preference":     req.BrandPreference,
		"special_requirements": req.SpecialRequirements,
		"notes":                req.Notes,
		"catalog":              catalog,
	}
	return doJSON[model.CatalogAdviceResponse](ctx, c.httpClient, http.MethodPost, c.baseURL+"/api/v1/advice/catalog", payload)
}

func (c *Client) GetSystemSettings(ctx context.Context) (model.SystemSettingsResponse, error) {
	return doJSON[model.SystemSettingsResponse](ctx, c.httpClient, http.MethodGet, c.baseURL+"/admin/api/v1/settings/system", nil)
}

func (c *Client) UpdateSystemSettings(ctx context.Context, req model.UpdateSystemSettingsRequest) (model.SystemSettingsResponse, error) {
	return doJSON[model.SystemSettingsResponse](ctx, c.httpClient, http.MethodPut, c.baseURL+"/admin/api/v1/settings/system", req)
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
		message := strings.TrimSpace(readUpstreamError(resp.Body))
		if message == "" {
			return zero, fmt.Errorf("upstream build-engine returned %d", resp.StatusCode)
		}
		return zero, fmt.Errorf("upstream build-engine returned %d: %s", resp.StatusCode, message)
	}
	if err := json.NewDecoder(resp.Body).Decode(&zero); err != nil {
		return zero, fmt.Errorf("decode response: %w", err)
	}
	return zero, nil
}

func readUpstreamError(body io.Reader) string {
	data, err := io.ReadAll(body)
	if err != nil || len(data) == 0 {
		return ""
	}
	var payload struct {
		Error string `json:"error"`
	}
	if err := json.Unmarshal(data, &payload); err == nil && strings.TrimSpace(payload.Error) != "" {
		return payload.Error
	}
	return string(data)
}
