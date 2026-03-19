package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadDefaults(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")
	content := []byte("" +
		"http_port: \"18084\"\n" +
		"postgres_dsn: postgres://rigel:rigel@postgres:5432/rigel?sslmode=disable\n" +
		"build_engine_base_url: http://rigel-build-engine:18082\n" +
		"build_engine_token: test-token\n" +
		"admin_username: admin\n" +
		"admin_password_hash: $2a$10$8BcB9rJ93cxzaO6F6DRK4OV2CCOrZvXSInSV+wOOyy48WaG3K7U9K\n")
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
	if cfg.JDCollectorBaseURL == "" {
		t.Fatal("expected jd-collector base url")
	}
}
