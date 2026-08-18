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
	Port     int    `envconfig:"PORT"     default:"9721"`
	LogLevel string `envconfig:"LOG_LEVEL" default:"info"`

	// Storage: single path — buffer -> batched upsert -> SQLite (AGENTS.md Iron Rule 5).
	DBPath     string `envconfig:"DB_PATH"     default:"data/count.db"`
	DBInterval int    `envconfig:"DB_INTERVAL" default:"10"`

	// Rate limiting thresholds (applied in M4, reserved here to keep the contract stable).
	RateLimitIPPerSec    int `envconfig:"RATE_LIMIT_IP_PER_SEC"      default:"10"`
	RateLimitIPPerMin    int `envconfig:"RATE_LIMIT_IP_PER_MIN"      default:"300"`
	RateLimitNamePerSec  int `envconfig:"RATE_LIMIT_NAME_PER_SEC"    default:"5"`
	RateLimitUploadPerHr int `envconfig:"RATE_LIMIT_UPLOAD_PER_HOUR" default:"5"`

	// TrustProxy enables X-Forwarded-* header trust so c.IP() returns the
	// real client behind a reverse proxy (Caddy/Nginx). Trusts loopback
	// by default (same-host proxy). Set TRUST_PROXY_PRIVATE=1 to also
	// trust private ranges when the proxy is on another internal host.
	TrustProxy        bool `envconfig:"TRUST_PROXY"          default:"true"`
	TrustProxyPrivate bool `envconfig:"TRUST_PROXY_PRIVATE" default:"false"`
}

// Load reads configuration from environment variables with the given prefix,
// then validates it. It returns an error if a value fails to parse or is
// out of range. Failing fast at startup prevents silent misbehavior later
// (e.g. a zero DB_INTERVAL stalling the flush ticker in M3).
func Load(prefix string) (*Config, error) {
	var c Config
	if err := envconfig.Process(prefix, &c); err != nil {
		return nil, fmt.Errorf("config load: %w", err)
	}
	if err := c.Validate(); err != nil {
		return nil, fmt.Errorf("config validate: %w", err)
	}
	return &c, nil
}

// Validate checks semantic constraints that envconfig tags cannot express.
func (c *Config) Validate() error {
	if c.Port < 1 || c.Port > 65535 {
		return fmt.Errorf("port must be in 1..65535, got %d", c.Port)
	}
	if c.DBInterval < 1 {
		return fmt.Errorf("db_interval must be >= 1 second, got %d", c.DBInterval)
	}
	if c.DBPath == "" {
		return fmt.Errorf("db_path must not be empty")
	}
	if c.RateLimitIPPerSec < 1 || c.RateLimitIPPerMin < 1 ||
		c.RateLimitNamePerSec < 1 || c.RateLimitUploadPerHr < 1 {
		return fmt.Errorf("rate limit thresholds must be >= 1")
	}
	return nil
}

// Addr returns the HTTP listen address derived from Host and Port.
func (c *Config) Addr() string {
	return fmt.Sprintf("%s:%d", c.Host, c.Port)
}
