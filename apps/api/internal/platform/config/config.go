// Package config loads runtime configuration from environment variables.
package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	Env             string // development | test | production
	HTTPAddr        string
	DatabaseURL     string
	PublicWebURL    string // used to build share links and allowed redirect origins
	CookieDomain    string
	CookieSecure    bool
	SessionTTL      time.Duration
	ReservationTTL  time.Duration
	FeatureFlags    string
	LogLevel        string
	ShutdownTimeout time.Duration
}

func (c Config) IsProduction() bool { return c.Env == "production" }

func Load() (Config, error) {
	cfg := Config{
		Env:             env("APP_ENV", "development"),
		HTTPAddr:        env("API_ADDR", ":8080"),
		DatabaseURL:     os.Getenv("DATABASE_URL"),
		PublicWebURL:    strings.TrimRight(env("PUBLIC_WEB_URL", "http://localhost:3000"), "/"),
		CookieDomain:    os.Getenv("COOKIE_DOMAIN"),
		FeatureFlags:    os.Getenv("FEATURE_FLAGS"),
		LogLevel:        env("LOG_LEVEL", "info"),
		ShutdownTimeout: 15 * time.Second,
	}
	var err error
	if cfg.CookieSecure, err = envBool("COOKIE_SECURE", cfg.Env == "production"); err != nil {
		return Config{}, err
	}
	if cfg.SessionTTL, err = envDuration("SESSION_TTL", 30*24*time.Hour); err != nil {
		return Config{}, err
	}
	if cfg.ReservationTTL, err = envDuration("RESERVATION_TTL", 7*24*time.Hour); err != nil {
		return Config{}, err
	}
	if cfg.DatabaseURL == "" {
		return Config{}, fmt.Errorf("config: DATABASE_URL is required")
	}
	return cfg, nil
}

func env(key, fallback string) string {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		return v
	}
	return fallback
}

func envBool(key string, fallback bool) (bool, error) {
	v := os.Getenv(key)
	if v == "" {
		return fallback, nil
	}
	b, err := strconv.ParseBool(v)
	if err != nil {
		return false, fmt.Errorf("config: %s: %w", key, err)
	}
	return b, nil
}

func envDuration(key string, fallback time.Duration) (time.Duration, error) {
	v := os.Getenv(key)
	if v == "" {
		return fallback, nil
	}
	d, err := time.ParseDuration(v)
	if err != nil {
		return 0, fmt.Errorf("config: %s: %w", key, err)
	}
	return d, nil
}
