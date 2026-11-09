package main

import (
	"errors"
	"flag"
	"testing"

	"github.com/miaoledor/lolicount/internal/config"
)

// validConfig builds a configuration that already passes Validate, so each
// test isolates the startup-flag behavior instead of unrelated defaults.
func validConfig() *config.Config {
	return &config.Config{
		Host:                "127.0.0.1",
		Port:                9721,
		LogLevel:            "info",
		DBPath:              "data/count.db",
		DBInterval:          10,
		RateLimitIPPerSec:   60,
		RateLimitIPPerMin:   3000,
		RateLimitNamePerSec: 20,
		ThemeResidentLimit:  10,
	}
}

// TestApplyRuntimeFlagsDefault pins the built-in default: with no arguments
// the loaded value (10 when the env var is absent) must survive untouched.
func TestApplyRuntimeFlagsDefault(t *testing.T) {
	cfg := validConfig()
	if err := applyRuntimeFlags(cfg, nil); err != nil {
		t.Fatalf("applyRuntimeFlags: %v", err)
	}
	if cfg.ThemeResidentLimit != 10 {
		t.Errorf("got %d, want 10", cfg.ThemeResidentLimit)
	}
}

// TestApplyRuntimeFlagsEnvValueWinsWithoutFlag verifies that an environment
// override is still honored when the CLI flag is not supplied.
func TestApplyRuntimeFlagsEnvValueWinsWithoutFlag(t *testing.T) {
	cfg := validConfig()
	cfg.ThemeResidentLimit = 32
	if err := applyRuntimeFlags(cfg, nil); err != nil {
		t.Fatalf("applyRuntimeFlags: %v", err)
	}
	if cfg.ThemeResidentLimit != 32 {
		t.Errorf("got %d, want 32", cfg.ThemeResidentLimit)
	}
}

// TestApplyRuntimeFlagsOverridesEnv verifies the documented precedence and
// both accepted flag spellings.
func TestApplyRuntimeFlagsOverridesEnv(t *testing.T) {
	for _, args := range [][]string{
		{"-theme-resident-limit", "5"},
		{"--theme-resident-limit=5"},
	} {
		cfg := validConfig()
		cfg.ThemeResidentLimit = 32
		if err := applyRuntimeFlags(cfg, args); err != nil {
			t.Fatalf("args %v: %v", args, err)
		}
		if cfg.ThemeResidentLimit != 5 {
			t.Errorf("args %v: got %d, want 5", args, cfg.ThemeResidentLimit)
		}
	}
}

// TestApplyRuntimeFlagsZeroKeepsAllThemes pins the escape hatch used by
// thumbnail generation and debugging.
func TestApplyRuntimeFlagsZeroKeepsAllThemes(t *testing.T) {
	cfg := validConfig()
	if err := applyRuntimeFlags(cfg, []string{"-theme-resident-limit=0"}); err != nil {
		t.Fatalf("applyRuntimeFlags: %v", err)
	}
	if cfg.ThemeResidentLimit != 0 {
		t.Errorf("got %d, want 0", cfg.ThemeResidentLimit)
	}
}

// TestApplyRuntimeFlagsRejectsInvalidValues covers negative values, parse
// failures and unknown flags so a typo fails fast at startup.
func TestApplyRuntimeFlagsRejectsInvalidValues(t *testing.T) {
	for _, args := range [][]string{
		{"-theme-resident-limit=-1"},
		{"-theme-resident-limit=abc"},
		{"-unknown-flag"},
	} {
		cfg := validConfig()
		if err := applyRuntimeFlags(cfg, args); err == nil {
			t.Errorf("args %v: expected an error", args)
		}
	}
}

// TestApplyRuntimeFlagsHelp keeps -h recognizable so main can exit cleanly
// instead of reporting a configuration failure.
func TestApplyRuntimeFlagsHelp(t *testing.T) {
	cfg := validConfig()
	if err := applyRuntimeFlags(cfg, []string{"-h"}); !errors.Is(err, flag.ErrHelp) {
		t.Errorf("got %v, want flag.ErrHelp", err)
	}
}
