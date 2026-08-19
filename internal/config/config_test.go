package config

import (
	"os"
	"strings"
	"testing"
)

// envVars is the full set of envconfig-managed variable names.
var envVars = []string{
	"HOST", "PORT", "LOG_LEVEL", "DB_PATH", "DB_INTERVAL",
	"RATE_LIMIT_IP_PER_SEC", "RATE_LIMIT_IP_PER_MIN",
	"RATE_LIMIT_NAME_PER_SEC", "RATE_LIMIT_UPLOAD_PER_HOUR",
	"BASE_URL",
}

// clearEnv unsets every managed variable and restores the originals on cleanup,
// so each test starts from defaults rather than the ambient environment
// (CI shells may export PORT etc.). Unsetting is required because envconfig
// treats an empty-but-set value as a real value, not a missing one.
func clearEnv(t *testing.T) {
	t.Helper()
	saved := map[string]string{}
	for _, k := range envVars {
		saved[k], _ = os.LookupEnv(k)
		os.Unsetenv(k)
	}
	t.Cleanup(func() {
		for _, k := range envVars {
			if v, ok := saved[k]; ok {
				os.Setenv(k, v)
			} else {
				os.Unsetenv(k)
			}
		}
	})
}

func TestLoadDefaults(t *testing.T) {
	clearEnv(t)
	c, err := Load("")
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if c.Host != "127.0.0.1" || c.Port != 9721 || c.LogLevel != "info" {
		t.Errorf("defaults: got %+v", c)
	}
	if c.DBPath != "data/count.db" || c.DBInterval != 10 {
		t.Errorf("storage defaults: got db=%q interval=%d", c.DBPath, c.DBInterval)
	}
	if c.Addr() != "127.0.0.1:9721" {
		t.Errorf("Addr: %q", c.Addr())
	}
}

func TestLoadOverrides(t *testing.T) {
	clearEnv(t)
	t.Setenv("HOST", "127.0.0.1")
	t.Setenv("PORT", "8080")
	t.Setenv("DB_INTERVAL", "5")
	t.Setenv("LOG_LEVEL", "debug")

	c, err := Load("")
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if c.Host != "127.0.0.1" || c.Port != 8080 || c.DBInterval != 5 || c.LogLevel != "debug" {
		t.Errorf("overrides: got %+v", c)
	}
	if c.Addr() != "127.0.0.1:8080" {
		t.Errorf("Addr: %q", c.Addr())
	}
}

func TestLoadInvalidPort(t *testing.T) {
	clearEnv(t)
	t.Setenv("PORT", "abc")
	if _, err := Load(""); err == nil {
		t.Fatal("expected parse error for non-integer port")
	}
}

func TestValidatePortRange(t *testing.T) {
	clearEnv(t)
	for _, p := range []string{"0", "-1", "65536", "100000"} {
		t.Setenv("PORT", p)
		if _, err := Load(""); err == nil {
			t.Errorf("port=%s: expected validation error", p)
		}
	}
}

func TestValidateDBInterval(t *testing.T) {
	clearEnv(t)
	for _, v := range []string{"0", "-5"} {
		t.Setenv("DB_INTERVAL", v)
		if _, err := Load(""); err == nil {
			t.Errorf("db_interval=%s: expected validation error", v)
		}
	}
}

func TestValidateRateLimits(t *testing.T) {
	clearEnv(t)
	t.Setenv("RATE_LIMIT_IP_PER_SEC", "0")
	if _, err := Load(""); err == nil {
		t.Fatal("expected validation error for zero ip rate limit")
	}
}

// Ensure error messages mention the offending field for operator clarity.
func TestLoadErrorMentionsField(t *testing.T) {
	clearEnv(t)
	t.Setenv("PORT", "0")
	_, err := Load("")
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "port") {
		t.Errorf("error should mention port: %v", err)
	}
}

// An empty DB_PATH must be rejected so the SQLite driver does not fail
// later with a confusing "unable to open database file" at flush time.
func TestValidateDBPathEmpty(t *testing.T) {
	clearEnv(t)
	t.Setenv("DB_PATH", "")
	if _, err := Load(""); err == nil {
		t.Fatal("expected validation error for empty db_path")
	}
}

// Each rate-limit threshold is guarded independently; verify the branches
// not covered by TestValidateRateLimits (per-min, per-name, upload).
func TestValidateEachRateLimit(t *testing.T) {
	cases := map[string]string{
		"RATE_LIMIT_IP_PER_MIN":      "0",
		"RATE_LIMIT_NAME_PER_SEC":    "0",
		"RATE_LIMIT_UPLOAD_PER_HOUR": "-1",
	}
	for k, v := range cases {
		clearEnv(t)
		t.Setenv(k, v)
		if _, err := Load(""); err == nil {
			t.Errorf("%s=%s: expected validation error", k, v)
		}
	}
}

// Port boundaries: 1 and 65535 are valid, 0 and 65536 are not.
func TestValidatePortBoundaries(t *testing.T) {
	clearEnv(t)
	for _, p := range []string{"1", "65535"} {
		t.Setenv("PORT", p)
		if _, err := Load(""); err != nil {
			t.Errorf("port=%s: expected ok, got %v", p, err)
		}
	}
}


// BaseURL defaults to empty (front-end falls back to same-origin).
func TestBaseURLDefaultEmpty(t *testing.T) {
	clearEnv(t)
	c, err := Load("")
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if c.BaseURL != "" {
		t.Errorf("BaseURL default should be empty, got %q", c.BaseURL)
	}
}

// A valid BASE_URL is accepted and trailing slashes are stripped.
func TestBaseURLValid(t *testing.T) {
	cases := []string{
		"https://umi7.top",
		"https://umi7.top/",
		"https://umi7.top///",
		"http://example.com:8080",
	}
	for _, v := range cases {
		clearEnv(t)
		t.Setenv("BASE_URL", v)
		c, err := Load("")
		if err != nil {
			t.Errorf("BASE_URL=%q: %v", v, err)
			continue
		}
		want := strings.TrimRight(v, "/")
		if c.BaseURL != want {
			t.Errorf("BASE_URL=%q: got %q want %q", v, c.BaseURL, want)
		}
	}
}

// BASE_URL without an http(s) scheme must be rejected so the front-end
// never emits a relative or javascript: embed link.
func TestBaseURLInvalidScheme(t *testing.T) {
	cases := []string{
		"umi7.top",
		"ftp://example.com",
		"//example.com",
		"javascript:alert(1)",
	}
	for _, v := range cases {
		clearEnv(t)
		t.Setenv("BASE_URL", v)
		if _, err := Load(""); err == nil {
			t.Errorf("BASE_URL=%q: expected validation error", v)
		}
	}
}
