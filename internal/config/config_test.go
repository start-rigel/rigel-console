package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestLoadDefaults(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")
	content := []byte("" +
		"http_port: \"18084\"\n" +
		"build_engine_base_url: http://rigel-build-engine:18082\n" +
		"build_engine_admin_token: rigel_console_admin_token_123456\n" +
		"build_engine_service_token: rigel_internal_service_token_123456\n")
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
	if cfg.BuildEngineTimeout != 35*time.Second {
		t.Fatalf("expected default build_engine_timeout 35s, got %s", cfg.BuildEngineTimeout)
	}
}

func TestLoadRejectsWeakAdminToken(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")
	content := []byte("" +
		"http_port: \"18084\"\n" +
		"build_engine_base_url: http://rigel-build-engine:18082\n" +
		"build_engine_admin_token: short\n" +
		"build_engine_service_token: rigel_internal_service_token_123456\n")
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

func TestLoadBuildEngineTimeout(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")
	content := []byte("" +
		"http_port: \"18084\"\n" +
		"build_engine_base_url: http://rigel-build-engine:18082\n" +
		"build_engine_admin_token: rigel_console_admin_token_123456\n" +
		"build_engine_service_token: rigel_internal_service_token_123456\n" +
		"build_engine_timeout: 45s\n")
	if err := os.WriteFile(path, content, 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.BuildEngineTimeout != 45*time.Second {
		t.Fatalf("expected build_engine_timeout 45s, got %s", cfg.BuildEngineTimeout)
	}
}

func TestLoadRejectsWeakServiceToken(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")
	content := []byte("" +
		"http_port: \"18084\"\n" +
		"build_engine_base_url: http://rigel-build-engine:18082\n" +
		"build_engine_admin_token: rigel_console_admin_token_123456\n" +
		"build_engine_service_token: short\n")
	if err := os.WriteFile(path, content, 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	_, err := Load(path)
	if err == nil {
		t.Fatal("expected weak service token error")
	}
	if !strings.Contains(err.Error(), "build_engine_service_token must be at least 24 characters") {
		t.Fatalf("unexpected error: %v", err)
	}
}
