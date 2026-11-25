// Package config loads application configuration from environment variables.
package config

import (
	"fmt"
	"os"
)

// Config holds the runtime configuration of the backend.
type Config struct {
	AppEnv      string
	PublicURL   string
	LogLevel    string
	HTTPAddr    string
	DatabaseURL string
}

// Load reads configuration from the environment. DATABASE_URL is required;
// every other value has a local-development default.
func Load() (Config, error) {
	cfg := Config{
		AppEnv:      envOrDefault("APP_ENV", "local"),
		PublicURL:   os.Getenv("APP_PUBLIC_URL"),
		LogLevel:    envOrDefault("LOG_LEVEL", "info"),
		HTTPAddr:    envOrDefault("HTTP_ADDR", ":8080"),
		DatabaseURL: os.Getenv("DATABASE_URL"),
	}
	if cfg.DatabaseURL == "" {
		return Config{}, fmt.Errorf("config: DATABASE_URL is required")
	}
	return cfg, nil
}

func envOrDefault(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
