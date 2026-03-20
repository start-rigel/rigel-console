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
	adminToken string
	httpClient *http.Client
}

func New(baseURL, adminToken string) *Client {
	return NewWithTimeout(baseURL, adminToken, 15*time.Second)
}

func NewWithTimeout(baseURL, adminToken string, timeout time.Duration) *Client {
	if timeout <= 0 {
		timeout = 15 * time.Second
	}
	return &Client{
		baseURL:    strings.TrimRight(baseURL, "/"),
		adminToken: strings.TrimSpace(adminToken),
		httpClient: &http.Client{Timeout: timeout},
	}
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
	return doJSON[model.BuildEnginePriceCatalog](ctx, c.httpClient, http.MethodGet, c.baseURL+"/api/v1/catalog/prices?"+query.Encode(), nil, nil)
}

func (c *Client) RecommendBuild(ctx context.Context, req model.GenerateBuildRequest) (model.CatalogRecommendationResponse, error) {
	payload := map[string]any{
		"budget":     req.Budget,
		"use_case":   req.UseCase,
		"build_mode": req.BuildMode,
		"notes":      req.Notes,
	}
	return doJSON[model.CatalogRecommendationResponse](ctx, c.httpClient, http.MethodPost, c.baseURL+"/api/v1/recommend/build", payload, nil)
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
	return doJSON[model.CatalogAdviceResponse](ctx, c.httpClient, http.MethodPost, c.baseURL+"/api/v1/advice/catalog", payload, nil)
}

func (c *Client) GetSystemSettings(ctx context.Context) (model.SystemSettingsResponse, error) {
	extraHeaders := map[string]string{}
	if c.adminToken != "" {
		extraHeaders["X-Rigel-Admin-Token"] = c.adminToken
	}
	return doJSON[model.SystemSettingsResponse](ctx, c.httpClient, http.MethodGet, c.baseURL+"/admin/api/v1/settings/system", nil, extraHeaders)
}

func (c *Client) UpdateSystemSettings(ctx context.Context, req model.UpdateSystemSettingsRequest) (model.SystemSettingsResponse, error) {
	extraHeaders := map[string]string{}
	if c.adminToken != "" {
		extraHeaders["X-Rigel-Admin-Token"] = c.adminToken
	}
	return doJSON[model.SystemSettingsResponse](ctx, c.httpClient, http.MethodPut, c.baseURL+"/admin/api/v1/settings/system", req, extraHeaders)
}

func doJSON[T any](ctx context.Context, httpClient *http.Client, method, target string, payload any, headers map[string]string) (T, error) {
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
	for key, value := range headers {
		if strings.TrimSpace(value) != "" {
			req.Header.Set(key, value)
		}
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
