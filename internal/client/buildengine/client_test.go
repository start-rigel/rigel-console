package buildengine

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/rigel-labs/rigel-console/internal/domain/model"
)

func TestDoJSONIncludesUpstreamErrorMessage(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"error":"catalog.items must not be empty"}`))
	}))
	defer server.Close()

	_, err := doJSON[map[string]any](context.Background(), server.Client(), http.MethodGet, server.URL, nil, nil)
	if err == nil {
		t.Fatal("expected upstream error")
	}
	if !strings.Contains(err.Error(), "catalog.items must not be empty") {
		t.Fatalf("expected upstream error message to be preserved, got %v", err)
	}
}

func TestGetSystemSettingsAddsAdminTokenHeader(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-Rigel-Admin-Token") != "token-123" {
			t.Fatalf("expected admin token header, got %q", r.Header.Get("X-Rigel-Admin-Token"))
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"ai_runtime":{"base_url":"","model":"m","timeout_seconds":25,"enabled":true,"gateway_token_configured":false,"gateway_token_masked":"","api_token_configured":false,"api_token_masked":""},"catalog_ai_limits":{"max_models_per_category":5}}`))
	}))
	defer server.Close()

	client := New(server.URL, "token-123", "rigel_internal_service_token_123456")
	_, err := client.GetSystemSettings(context.Background())
	if err != nil {
		t.Fatalf("GetSystemSettings() error = %v", err)
	}
}

func TestRecommendBuildUsesServiceTokenButNotAdminToken(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if token := r.Header.Get("X-Rigel-Admin-Token"); token != "" {
			t.Fatalf("expected no admin token for non-admin endpoint, got %q", token)
		}
		if token := r.Header.Get("X-Rigel-Service-Token"); token != "service-token-123" {
			t.Fatalf("expected service token header, got %q", token)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"request_status":{"cache_hit":false,"remaining_ai_requests":4,"cooldown_seconds":0},"provider":"build-engine","fallback_used":true,"request":{"budget":6000,"use_case":"gaming","build_mode":"mixed"},"summary":"ok","estimated_total":3200,"within_budget":true,"build_items":[],"advice":{"reasons":[],"risks":[],"upgrade_advice":[]}}`))
	}))
	defer server.Close()

	client := New(server.URL, "token-123", "service-token-123")
	_, err := client.RecommendBuild(context.Background(), model.GenerateBuildRequest{Budget: 6000, UseCase: "gaming", BuildMode: "mixed"})
	if err != nil {
		t.Fatalf("RecommendBuild() error = %v", err)
	}
}

func TestNewWithTimeout(t *testing.T) {
	client := NewWithTimeout("http://example.com", "token-123", "service-token-123", 40*time.Second)
	if client.httpClient.Timeout != 40*time.Second {
		t.Fatalf("expected timeout 40s, got %s", client.httpClient.Timeout)
	}
}
