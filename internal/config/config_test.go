package config

import "testing"

func TestLoadDefaults(t *testing.T) {
	t.Setenv("RIGEL_SERVICE_NAME", "")
	t.Setenv("RIGEL_HTTP_PORT", "")
	t.Setenv("RIGEL_HTTP_READ_TIMEOUT", "")
	t.Setenv("RIGEL_HTTP_WRITE_TIMEOUT", "")
	t.Setenv("RIGEL_HTTP_IDLE_TIMEOUT", "")
	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.BuildEngineBaseURL == "" {
		t.Fatal("expected build-engine base url")
	}
}
