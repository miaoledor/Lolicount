package server

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/miaoledor/lolicount/internal/config"
	"github.com/miaoledor/lolicount/internal/counter"
	"github.com/miaoledor/lolicount/internal/imgcore/cardthemedrawer"
	"github.com/miaoledor/lolicount/internal/imgcore/fdrawer"
	"github.com/miaoledor/lolicount/internal/store"
	"github.com/rs/zerolog"
)

// stubFThemeRegistry is an in-memory fdrawer.Registry for handler tests.
type stubFThemeRegistry struct {
	styles map[string]fdrawer.Style
}

func (s *stubFThemeRegistry) Get(name string) (fdrawer.Style, bool) {
	st, ok := s.styles[name]
	return st, ok
}
func (s *stubFThemeRegistry) List() []string {
	out := make([]string, 0, len(s.styles))
	for k := range s.styles {
		out = append(out, k)
	}
	return out
}

// newFThemeServer builds a server with one theme and one f-theme style.
func newFThemeServer(t *testing.T) *Server {
	t.Helper()
	th := &cardthemedrawer.Theme{Name: "lian", Frames: []cardthemedrawer.Frame{{Width: 10, Height: 20, Data: "data:image/gif;base64,F0"}}}
	reg := &stubRegistry{cards: map[string]*cardthemedrawer.Theme{"lian": th}}
	ft := &stubFThemeRegistry{styles: map[string]fdrawer.Style{
		"pink": {Name: "pink", Family: "monospace", Color: "#e91e63", Weight: "bold"},
	}}
	repo, err := store.NewSQLite(context.Background(), ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if c, ok := repo.(interface{ Close() error }); ok {
			c.Close()
		}
	})
	buf := counter.New(repo, zerolog.Nop(), 3600)
	if err := buf.Start(context.Background()); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(buf.Stop)
	cfg := &config.Config{Host: "127.0.0.1", Port: 0, DBInterval: 10, RateLimitIPPerSec: 10000, RateLimitIPPerMin: 100000, RateLimitNamePerSec: 10000}
	s := New(cfg, zerolog.Nop(), reg, ft, buf)
	t.Cleanup(func() {
		s.ipLimiter.Stop()
		s.nameLimiter.Stop()
	})
	return s
}

// M6: ?ftheme= applies the style to the counter text.
func TestCounterFTheme(t *testing.T) {
	s := newFThemeServer(t)
	req := httptest.NewRequest(http.MethodGet, "/@demo?theme=lian&ftheme=pink", nil)
	resp, err := s.app.Test(req)
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status: %d", resp.StatusCode)
	}
	body := readBody(t, resp)
	if !strings.Contains(body, `fill="#e91e63"`) {
		t.Errorf("ftheme=pink should set color: %s", sub(body, "fill"))
	}
	if !strings.Contains(body, `font-weight="bold"`) {
		t.Errorf("ftheme=pink should set weight: %s", sub(body, "font-weight"))
	}
}

// M6: unknown ftheme 400s.
func TestCounterUnknownFTheme400(t *testing.T) {
	s := newFThemeServer(t)
	req := httptest.NewRequest(http.MethodGet, "/@demo?theme=lian&ftheme=nonexistent", nil)
	resp, _ := s.app.Test(req)
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("unknown ftheme should 400, got %d", resp.StatusCode)
	}
}

// M6: random ftheme picks from the registry.
func TestCounterRandomFTheme(t *testing.T) {
	s := newFThemeServer(t)
	req := httptest.NewRequest(http.MethodGet, "/@demo?theme=lian&ftheme=random", nil)
	resp, _ := s.app.Test(req)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("random ftheme should 200, got %d", resp.StatusCode)
	}
}
