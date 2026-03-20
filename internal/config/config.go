package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

// Config contains the runtime contract for the console service.
type Config struct {
	ServiceName           string        `yaml:"service_name"`
	HTTPPort              string        `yaml:"http_port"`
	LogLevel              string        `yaml:"log_level"`
	FrontendMode          string        `yaml:"frontend_mode"`
	BuildEngineBaseURL    string        `yaml:"build_engine_base_url"`
	BuildEngineAdminToken string        `yaml:"build_engine_admin_token"`
	BuildEngineTimeout    time.Duration `yaml:"-"`
	AdminAllowedCIDRs     []string      `yaml:"admin_allowed_cidrs"`
	TrustedProxyCIDRs     []string      `yaml:"trusted_proxy_cidrs"`
	ChallengeProvider     string        `yaml:"challenge_provider"`
	ChallengeSiteKey      string        `yaml:"challenge_site_key"`
	AdminUsername         string        `yaml:"admin_username"`
	AdminPassword         string        `yaml:"admin_password"`
	AdminCookieName       string        `yaml:"admin_cookie_name"`
	AdminCSRFCookieName   string        `yaml:"admin_csrf_cookie_name"`
	AnonymousCookieName   string        `yaml:"anonymous_cookie_name"`
	AnonymousHourlyLimit  int           `yaml:"anonymous_hourly_limit"`
	CooldownSeconds       int           `yaml:"cooldown_seconds"`
	ReadTimeout           time.Duration `yaml:"-"`
	WriteTimeout          time.Duration `yaml:"-"`
	IdleTimeout           time.Duration `yaml:"-"`
}

type fileConfig struct {
	ServiceName           string   `yaml:"service_name"`
	HTTPPort              string   `yaml:"http_port"`
	LogLevel              string   `yaml:"log_level"`
	FrontendMode          string   `yaml:"frontend_mode"`
	BuildEngineBaseURL    string   `yaml:"build_engine_base_url"`
	BuildEngineAdminToken string   `yaml:"build_engine_admin_token"`
	BuildEngineTimeout    string   `yaml:"build_engine_timeout"`
	AdminAllowedCIDRs     []string `yaml:"admin_allowed_cidrs"`
	TrustedProxyCIDRs     []string `yaml:"trusted_proxy_cidrs"`
	ChallengeProvider     string   `yaml:"challenge_provider"`
	ChallengeSiteKey      string   `yaml:"challenge_site_key"`
	AdminUsername         string   `yaml:"admin_username"`
	AdminPassword         string   `yaml:"admin_password"`
	AdminCookieName       string   `yaml:"admin_cookie_name"`
	AdminCSRFCookieName   string   `yaml:"admin_csrf_cookie_name"`
	AnonymousCookieName   string   `yaml:"anonymous_cookie_name"`
	AnonymousHourlyLimit  int      `yaml:"anonymous_hourly_limit"`
	CooldownSeconds       int      `yaml:"cooldown_seconds"`
	ReadTimeout           string   `yaml:"read_timeout"`
	WriteTimeout          string   `yaml:"write_timeout"`
	IdleTimeout           string   `yaml:"idle_timeout"`
}

func DefaultPath() string {
	return filepath.Join("configs", "config.yaml")
}

func Load(path string) (Config, error) {
	if path == "" {
		path = DefaultPath()
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return Config{}, fmt.Errorf("read config file %s: %w", path, err)
	}
	var raw fileConfig
	if err := yaml.Unmarshal(data, &raw); err != nil {
		return Config{}, fmt.Errorf("parse config file %s: %w", path, err)
	}
	readTimeout, err := parseDuration(raw.ReadTimeout, 5*time.Second)
	if err != nil {
		return Config{}, err
	}
	writeTimeout, err := parseDuration(raw.WriteTimeout, 2*time.Minute)
	if err != nil {
		return Config{}, err
	}
	idleTimeout, err := parseDuration(raw.IdleTimeout, 30*time.Second)
	if err != nil {
		return Config{}, err
	}
	buildEngineTimeout, err := parseDuration(raw.BuildEngineTimeout, 35*time.Second)
	if err != nil {
		return Config{}, err
	}

	cfg := Config{
		ServiceName:           blankFallback(raw.ServiceName, "rigel-console"),
		HTTPPort:              blankFallback(raw.HTTPPort, "8080"),
		LogLevel:              blankFallback(raw.LogLevel, "info"),
		FrontendMode:          blankFallback(raw.FrontendMode, "embedded"),
		BuildEngineBaseURL:    blankFallback(raw.BuildEngineBaseURL, "http://rigel-build-engine:18082"),
		BuildEngineAdminToken: blankFallback(os.Getenv("RIGEL_BUILD_ENGINE_ADMIN_TOKEN"), raw.BuildEngineAdminToken),
		BuildEngineTimeout:    buildEngineTimeout,
		AdminAllowedCIDRs:     cidrFallback(raw.AdminAllowedCIDRs, []string{"127.0.0.0/8", "::1/128"}),
		TrustedProxyCIDRs:     cidrFallback(raw.TrustedProxyCIDRs, []string{"127.0.0.0/8", "::1/128"}),
		ChallengeProvider:     strings.TrimSpace(raw.ChallengeProvider),
		ChallengeSiteKey:      strings.TrimSpace(raw.ChallengeSiteKey),
		AdminUsername:         blankFallback(raw.AdminUsername, "admin"),
		AdminPassword:         blankFallback(raw.AdminPassword, "admin123456"),
		AdminCookieName:       blankFallback(raw.AdminCookieName, "rigel_admin_session"),
		AdminCSRFCookieName:   blankFallback(raw.AdminCSRFCookieName, "rigel_admin_csrf"),
		AnonymousCookieName:   blankFallback(raw.AnonymousCookieName, "rigel_anonymous_id"),
		AnonymousHourlyLimit:  intFallback(raw.AnonymousHourlyLimit, 5),
		CooldownSeconds:       intFallback(raw.CooldownSeconds, 60),
		ReadTimeout:           readTimeout,
		WriteTimeout:          writeTimeout,
		IdleTimeout:           idleTimeout,
	}
	if cfg.HTTPPort == "" {
		return Config{}, fmt.Errorf("RIGEL_HTTP_PORT must not be empty")
	}
	if err := validateBuildEngineAdminToken(cfg.BuildEngineAdminToken); err != nil {
		return Config{}, err
	}
	return cfg, nil
}

func parseDuration(value string, fallback time.Duration) (time.Duration, error) {
	if value == "" {
		return fallback, nil
	}
	parsed, err := time.ParseDuration(value)
	if err != nil {
		return 0, fmt.Errorf("parse duration %q: %w", value, err)
	}
	return parsed, nil
}

func blankFallback(value, fallback string) string {
	if value == "" {
		return fallback
	}
	return value
}

func intFallback(value, fallback int) int {
	if value <= 0 {
		return fallback
	}
	return value
}

func validateBuildEngineAdminToken(token string) error {
	trimmed := strings.TrimSpace(token)
	if trimmed == "" {
		return fmt.Errorf("build_engine_admin_token must not be empty")
	}
	if len(trimmed) < 24 {
		return fmt.Errorf("build_engine_admin_token must be at least 24 characters")
	}
	if strings.EqualFold(trimmed, "rigel-build-engine-admin-token") {
		return fmt.Errorf("build_engine_admin_token must not use the default development token")
	}
	return nil
}

func cidrFallback(values, fallback []string) []string {
	if len(values) == 0 {
		return append([]string{}, fallback...)
	}
	out := make([]string, 0, len(values))
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" {
			continue
		}
		out = append(out, value)
	}
	if len(out) == 0 {
		return append([]string{}, fallback...)
	}
	return out
}
