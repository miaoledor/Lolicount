// Package logger initializes the global zerolog logger used across the project.
package logger

import (
	"os"
	"strings"
	"time"

	"github.com/rs/zerolog"
)

// L is the package-level logger. Call Init once at startup before use.
var L zerolog.Logger

// Init configures the global logger with the given level string.
// Unknown levels fall back to info. Output is pretty-printed on a TTY,
// raw JSON otherwise (production-friendly).
func Init(level string) {
	zerolog.TimeFieldFormat = time.RFC3339
	lvl := parseLevel(level)

	if isTTY() {
		L = zerolog.New(zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: time.RFC3339}).
			Level(lvl).
			With().Timestamp().Logger()
	} else {
		L = zerolog.New(os.Stdout).Level(lvl).With().Timestamp().Logger()
	}
}

// parseLevel maps a config string to a zerolog level, defaulting to info.
func parseLevel(s string) zerolog.Level {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "trace":
		return zerolog.TraceLevel
	case "debug":
		return zerolog.DebugLevel
	case "warn", "warning":
		return zerolog.WarnLevel
	case "error":
		return zerolog.ErrorLevel
	case "fatal":
		return zerolog.FatalLevel
	case "panic":
		return zerolog.PanicLevel
	case "disabled":
		return zerolog.Disabled
	default:
		return zerolog.InfoLevel
	}
}

// isTTY reports whether stdout is a terminal, to pick pretty vs JSON output.
func isTTY() bool {
	fi, err := os.Stdout.Stat()
	if err != nil {
		return false
	}
	return (fi.Mode() & os.ModeCharDevice) != 0
}
