// Package config loads application configuration from environment variables.
// All fields carry sane defaults so the server runs without a .env file.
package config

import (
	"fmt"

	"github.com/kelseyhightower/envconfig"
)

// Config holds all runtime configuration loaded from the environment.
type Config struct {
	Host     string `envconfig:"HOST"     default:"0.0.0.0"`
	Port     int    `envconfig:"PORT"     default:"3000"`
	LogLevel string `envconfig:"LOG_LEVEL" default:"info"`

	// Storage: single path — buffer -> batched upsert -> SQLite (AGENTS.md Iron Rule 5).
	DBPath     string `envconfig:"DB_PATH"     default:"data/count.db"`
	DBInterval int    `envconfig:"DB_INTERVAL" default:"10"`

	// Rate limiting thresholds (applied in M4, reserved here to keep the contract stable).
	RateLimitIPPerSec     int `envconfig:"RATE_LIMIT_IP_PER_SEC"     default:"10"`
	RateLimitIPPerMin     int `envconfig:"RATE_LIMIT_IP_PER_MIN"     default:"300"`
	RateLimitNamePerSec   int `envconfig:"RATE_LIMIT_NAME_PER_SEC"   default:"5"`
	RateLimitUploadPerHr  int `envconfig:"RATE_LIMIT_UPLOAD_PER_HOUR" default:"5"`
}

// Load reads configuration from environment variables with the given prefix.
// It returns an error if any value fails to parse into its declared type.
func Load(prefix string) (*Config, error) {
	var c Config
	if err := envconfig.Process(prefix, &c); err != nil {
		return nil, fmt.Errorf("config load: %w", err)
	}
	return &c, nil
}

// Addr returns the HTTP listen address derived from Host and Port.
func (c *Config) Addr() string {
	return fmt.Sprintf("%s:%d", c.Host, c.Port)
}
