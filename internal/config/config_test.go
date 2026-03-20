package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLoadDefaults(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")
	content := []byte("" +
		"http_port: \"18084\"\n" +
		"build_engine_base_url: http://rigel-build-engine:18082\n" +
		"build_engine_admin_token: rigel_console_admin_token_123456\n")
	if err := os.WriteFile(path, content, 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.BuildEngineBaseURL == "" {
		t.Fatal("expected build-engine base url")
	}
	if len(cfg.AdminAllowedCIDRs) == 0 {
		t.Fatal("expected default admin_allowed_cidrs")
	}
}

func TestLoadRejectsWeakAdminToken(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")
	content := []byte("" +
		"http_port: \"18084\"\n" +
		"build_engine_base_url: http://rigel-build-engine:18082\n" +
		"build_engine_admin_token: short\n")
	if err := os.WriteFile(path, content, 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	_, err := Load(path)
	if err == nil {
		t.Fatal("expected weak token error")
	}
	if !strings.Contains(err.Error(), "at least 24 characters") {
		t.Fatalf("unexpected error: %v", err)
	}
}
