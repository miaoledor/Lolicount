package logger

import (
	"testing"

	"github.com/rs/zerolog"
)

func TestParseLevel(t *testing.T) {
	cases := []struct {
		in   string
		want zerolog.Level
	}{
		{"trace", zerolog.TraceLevel},
		{"DEBUG", zerolog.DebugLevel},
		{"info", zerolog.InfoLevel},
		{"warn", zerolog.WarnLevel},
		{"warning", zerolog.WarnLevel},
		{"error", zerolog.ErrorLevel},
		{"fatal", zerolog.FatalLevel},
		{"panic", zerolog.PanicLevel},
		{"disabled", zerolog.Disabled},
		{"  Info  ", zerolog.InfoLevel}, // trimmed + case-insensitive
		{"", zerolog.InfoLevel},         // empty -> default
		{"nope", zerolog.InfoLevel},     // unknown -> default
	}
	for _, tc := range cases {
		got := parseLevel(tc.in)
		if got != tc.want {
			t.Errorf("parseLevel(%q) = %v, want %v", tc.in, got, tc.want)
		}
	}
}

// Init must not panic and must leave the global logger at the requested level.
func TestInitSetsLevel(t *testing.T) {
	for _, lvl := range []string{"debug", "error", "disabled"} {
		Init(lvl)
		want := parseLevel(lvl)
		// zerolog exposes the level via a level field on the internal logger;
		// comparing the String form is the stable public surface.
		if got := L.GetLevel(); got != want {
			t.Errorf("after Init(%q): level=%v want=%v", lvl, got, want)
		}
	}
}
