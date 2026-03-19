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
	ServiceName          string        `yaml:"service_name"`
	HTTPPort             string        `yaml:"http_port"`
	LogLevel             string        `yaml:"log_level"`
	FrontendMode         string        `yaml:"frontend_mode"`
	PostgresDSN          string        `yaml:"postgres_dsn"`
	BuildEngineBaseURL   string        `yaml:"build_engine_base_url"`
	BuildEngineToken     string        `yaml:"build_engine_token"`
	JDCollectorBaseURL   string        `yaml:"jd_collector_base_url"`
	AdminUsername        string        `yaml:"admin_username"`
	AdminPassword        string        `yaml:"admin_password"`
	AdminPasswordHash    string        `yaml:"admin_password_hash"`
	AdminCookieName      string        `yaml:"admin_cookie_name"`
	AdminCSRFCookieName  string        `yaml:"admin_csrf_cookie_name"`
	AnonymousCookieName  string        `yaml:"anonymous_cookie_name"`
	AnonymousHourlyLimit int           `yaml:"anonymous_hourly_limit"`
	IPHourlyLimit        int           `yaml:"ip_hourly_limit"`
	DeviceHourlyLimit    int           `yaml:"device_hourly_limit"`
	CooldownSeconds      int           `yaml:"cooldown_seconds"`
	ChallengePassSeconds int           `yaml:"challenge_pass_seconds"`
	SessionTTLSeconds    int           `yaml:"session_ttl_seconds"`
	RedisAddr            string        `yaml:"redis_addr"`
	ChallengeProvider    string        `yaml:"challenge_provider"`
	ChallengeSiteKey     string        `yaml:"challenge_site_key"`
	ChallengeSecret      string        `yaml:"challenge_secret"`
	ChallengeVerifyURL   string        `yaml:"challenge_verify_url"`
	TrustedProxyCIDRs    []string      `yaml:"trusted_proxy_cidrs"`
	AdminAllowedCIDRs    []string      `yaml:"admin_allowed_cidrs"`
	ReadTimeout          time.Duration `yaml:"-"`
	WriteTimeout         time.Duration `yaml:"-"`
	IdleTimeout          time.Duration `yaml:"-"`
}

type fileConfig struct {
	ServiceName          string   `yaml:"service_name"`
	HTTPPort             string   `yaml:"http_port"`
	LogLevel             string   `yaml:"log_level"`
	FrontendMode         string   `yaml:"frontend_mode"`
	PostgresDSN          string   `yaml:"postgres_dsn"`
	BuildEngineBaseURL   string   `yaml:"build_engine_base_url"`
	BuildEngineToken     string   `yaml:"build_engine_token"`
	JDCollectorBaseURL   string   `yaml:"jd_collector_base_url"`
	AdminUsername        string   `yaml:"admin_username"`
	AdminPassword        string   `yaml:"admin_password"`
	AdminPasswordHash    string   `yaml:"admin_password_hash"`
	AdminCookieName      string   `yaml:"admin_cookie_name"`
	AdminCSRFCookieName  string   `yaml:"admin_csrf_cookie_name"`
	AnonymousCookieName  string   `yaml:"anonymous_cookie_name"`
	AnonymousHourlyLimit int      `yaml:"anonymous_hourly_limit"`
	IPHourlyLimit        int      `yaml:"ip_hourly_limit"`
	DeviceHourlyLimit    int      `yaml:"device_hourly_limit"`
	CooldownSeconds      int      `yaml:"cooldown_seconds"`
	ChallengePassSeconds int      `yaml:"challenge_pass_seconds"`
	SessionTTLSeconds    int      `yaml:"session_ttl_seconds"`
	RedisAddr            string   `yaml:"redis_addr"`
	ChallengeProvider    string   `yaml:"challenge_provider"`
	ChallengeSiteKey     string   `yaml:"challenge_site_key"`
	ChallengeSecret      string   `yaml:"challenge_secret"`
	ChallengeVerifyURL   string   `yaml:"challenge_verify_url"`
	TrustedProxyCIDRs    []string `yaml:"trusted_proxy_cidrs"`
	AdminAllowedCIDRs    []string `yaml:"admin_allowed_cidrs"`
	ReadTimeout          string   `yaml:"read_timeout"`
	WriteTimeout         string   `yaml:"write_timeout"`
	IdleTimeout          string   `yaml:"idle_timeout"`
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

	cfg := Config{
		ServiceName:          blankFallback(raw.ServiceName, "rigel-console"),
		HTTPPort:             blankFallback(raw.HTTPPort, "8080"),
		LogLevel:             blankFallback(raw.LogLevel, "info"),
		FrontendMode:         blankFallback(raw.FrontendMode, "embedded"),
		PostgresDSN:          strings.TrimSpace(raw.PostgresDSN),
		BuildEngineBaseURL:   blankFallback(raw.BuildEngineBaseURL, "http://rigel-build-engine:18082"),
		BuildEngineToken:     raw.BuildEngineToken,
		JDCollectorBaseURL:   blankFallback(raw.JDCollectorBaseURL, "http://rigel-jd-collector:18081"),
		AdminUsername:        strings.TrimSpace(raw.AdminUsername),
		AdminPassword:        raw.AdminPassword,
		AdminPasswordHash:    strings.TrimSpace(raw.AdminPasswordHash),
		AdminCookieName:      blankFallback(raw.AdminCookieName, "rigel_admin_session"),
		AdminCSRFCookieName:  blankFallback(raw.AdminCSRFCookieName, "rigel_admin_csrf"),
		AnonymousCookieName:  blankFallback(raw.AnonymousCookieName, "rigel_anonymous_id"),
		AnonymousHourlyLimit: intFallback(raw.AnonymousHourlyLimit, 5),
		IPHourlyLimit:        intFallback(raw.IPHourlyLimit, 20),
		DeviceHourlyLimit:    intFallback(raw.DeviceHourlyLimit, 12),
		CooldownSeconds:      intFallback(raw.CooldownSeconds, 60),
		ChallengePassSeconds: intFallback(raw.ChallengePassSeconds, 900),
		SessionTTLSeconds:    intFallback(raw.SessionTTLSeconds, 2592000),
		RedisAddr:            strings.TrimSpace(raw.RedisAddr),
		ChallengeProvider:    strings.TrimSpace(raw.ChallengeProvider),
		ChallengeSiteKey:     strings.TrimSpace(raw.ChallengeSiteKey),
		ChallengeSecret:      strings.TrimSpace(raw.ChallengeSecret),
		ChallengeVerifyURL:   blankFallback(raw.ChallengeVerifyURL, "https://challenges.cloudflare.com/turnstile/v0/siteverify"),
		TrustedProxyCIDRs:    append([]string(nil), raw.TrustedProxyCIDRs...),
		AdminAllowedCIDRs:    normalizeCIDRDefaults(raw.AdminAllowedCIDRs),
		ReadTimeout:          readTimeout,
		WriteTimeout:         writeTimeout,
		IdleTimeout:          idleTimeout,
	}
	if cfg.HTTPPort == "" {
		return Config{}, fmt.Errorf("RIGEL_HTTP_PORT must not be empty")
	}
	if cfg.PostgresDSN == "" {
		return Config{}, fmt.Errorf("postgres_dsn must not be empty")
	}
	if cfg.AdminUsername == "" {
		return Config{}, fmt.Errorf("admin_username must not be empty")
	}
	if cfg.AdminPasswordHash == "" && strings.TrimSpace(cfg.AdminPassword) == "" {
		return Config{}, fmt.Errorf("admin_password_hash or admin_password must not be empty")
	}
	if cfg.AdminPassword == "admin123456" {
		return Config{}, fmt.Errorf("admin_password must not use default weak password")
	}
	if cfg.BuildEngineToken == "" {
		return Config{}, fmt.Errorf("build_engine_token must not be empty")
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

func normalizeCIDRDefaults(values []string) []string {
	if len(values) > 0 {
		return append([]string(nil), values...)
	}
	return []string{
		"127.0.0.1/32",
		"::1/128",
		"10.0.0.0/8",
		"172.16.0.0/12",
		"192.168.0.0/16",
		"100.64.0.0/10",
	}
}
