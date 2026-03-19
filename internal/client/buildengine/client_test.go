package buildengine

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestDoJSONIncludesUpstreamErrorMessage(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"error":"catalog.items must not be empty"}`))
	}))
	defer server.Close()

	_, err := doJSON[map[string]any](context.Background(), server.Client(), http.MethodGet, server.URL, nil)
	if err == nil {
		t.Fatal("expected upstream error")
	}
	if !strings.Contains(err.Error(), "catalog.items must not be empty") {
		t.Fatalf("expected upstream error message to be preserved, got %v", err)
	}
}
