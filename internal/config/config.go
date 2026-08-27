// Package config loads application configuration from environment variables.
// All fields carry sane defaults so the server runs without a .env file.
package config

import (
	"fmt"
	"strings"

	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"
)

// Config holds all runtime configuration loaded from the environment.
type Config struct {
	Host     string `envconfig:"HOST"     default:"127.0.0.1"`
	Port     int    `envconfig:"PORT"     default:"9721"`
	LogLevel string `envconfig:"LOG_LEVEL" default:"info"`

	// Storage: single path — buffer -> batched upsert -> SQLite (AGENTS.md Iron Rule 5).
	DBPath     string `envconfig:"DB_PATH"     default:"data/count.db"`
	DBInterval int    `envconfig:"DB_INTERVAL" default:"10"`

	// Rate limiting thresholds (applied in M4, reserved here to keep the contract stable).
	RateLimitIPPerSec    int `envconfig:"RATE_LIMIT_IP_PER_SEC"      default:"60"`
	RateLimitIPPerMin    int `envconfig:"RATE_LIMIT_IP_PER_MIN"      default:"3000"`
	RateLimitNamePerSec  int `envconfig:"RATE_LIMIT_NAME_PER_SEC"    default:"20"`
	RateLimitUploadPerHr int `envconfig:"RATE_LIMIT_UPLOAD_PER_HOUR" default:"10"`

	// TrustProxy enables X-Forwarded-* header trust so c.IP() returns the
	// real client behind a reverse proxy (Caddy/Nginx). Trusts loopback
	// by default (same-host proxy). Set TRUST_PROXY_PRIVATE=1 to also
	// trust private ranges when the proxy is on another internal host.
	TrustProxy        bool `envconfig:"TRUST_PROXY"          default:"true"`
	TrustProxyPrivate bool `envconfig:"TRUST_PROXY_PRIVATE" default:"false"`

	// BaseURL is the public origin (scheme://host[:port]) used to build
	// embed links shown on the web UI. It does NOT affect where the server
	// listens — only the URLs the front-end suggests users paste into
	// READMEs. Empty = fall back to the request's own origin. Trailing
	// slashes are stripped at validation time.
	BaseURL string `envconfig:"BASE_URL" default:""`

	// AdminKey protects admin-only endpoints (theme review, approve,
	// reject, delete). Empty = admin endpoints disabled (return 404).
	AdminKey string `envconfig:"ADMIN_KEY" default:""`
}

// Load reads configuration from environment variables with the given prefix,
// then validates it. It returns an error if a value fails to parse or is
// out of range. Failing fast at startup prevents silent misbehavior later
// (e.g. a zero DB_INTERVAL stalling the flush ticker in M3).
func Load(prefix string) (*Config, error) {
	// Load .env if present (dev convenience). In production env vars are
	// usually set directly; a missing file is not an error. godotenv never
	// overrides vars that already exist in the environment, so explicit
	// env settings always win over .env.
	_ = godotenv.Load()
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
	if err := c.validateBaseURL(); err != nil {
		return err
	}
	return nil
}

// validateBaseURL normalizes BaseURL (strips trailing slashes) and, when
// set, requires an http(s) scheme so the front-end never emits a relative
// or javascript: embed link.
func (c *Config) validateBaseURL() error {
	c.BaseURL = strings.TrimRight(c.BaseURL, "/")
	if c.BaseURL == "" {
		return nil
	}
	if !strings.HasPrefix(c.BaseURL, "http://") && !strings.HasPrefix(c.BaseURL, "https://") {
		return fmt.Errorf("base_url must start with http:// or https://, got %q", c.BaseURL)
	}
	return nil
}

// Addr returns the HTTP listen address derived from Host and Port.
func (c *Config) Addr() string {
	return fmt.Sprintf("%s:%d", c.Host, c.Port)
}
