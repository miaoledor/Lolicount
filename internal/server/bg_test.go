package server

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/miaoledor/lolicount/internal/bg"
	"github.com/miaoledor/lolicount/internal/config"
	"github.com/miaoledor/lolicount/internal/counter"
	"github.com/miaoledor/lolicount/internal/store"
	"github.com/miaoledor/lolicount/internal/theme"
	"github.com/rs/zerolog"
)

// stubBgRegistry is an in-memory bg.Registry for handler tests.
type stubBgRegistry struct {
	bgs map[string]bg.Background
}

func (s *stubBgRegistry) Get(name string) (bg.Background, bool) {
	b, ok := s.bgs[name]
	return b, ok
}
func (s *stubBgRegistry) List() []string {
	out := make([]string, 0, len(s.bgs))
	for k := range s.bgs {
		out = append(out, k)
	}
	return out
}

// newBgServer builds a server with one theme frame and one background.
func newBgServer(t *testing.T) *Server {
	t.Helper()
	th := &theme.Theme{Name: "loli", Frames: []theme.Frame{{Width: 20, Height: 30, Data: "data:image/gif;base64,QQ"}}}
	reg := &stubRegistry{themes: map[string]*theme.Theme{"loli": th}}
	bgs := &stubBgRegistry{bgs: map[string]bg.Background{
		"loli-stand": {Name: "loli-stand", URL: "https://cdn.example.com/bg.png", Width: 400, Height: 300},
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
	s := New(cfg, zerolog.Nop(), reg, buf, bgs)
	t.Cleanup(func() {
		s.ipLimiter.Stop()
		s.nameLimiter.Stop()
	})
	return s
}

func TestCounterWithBackgroundOverlay(t *testing.T) {
	s := newBgServer(t)
	req := httptest.NewRequest(http.MethodGet, "/@bgtest?theme=loli&bg=loli-stand&x=20&y=180&fsize=40", nil)
	resp, err := s.app.Test(req)
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status: %d", resp.StatusCode)
	}
	body := readBody(t, resp)
	// Background external URL (Iron Rule 2).
	if !strings.Contains(body, `href="https://cdn.example.com/bg.png"`) {
		t.Errorf("missing background URL: %s", body)
	}
	// viewBox fixed to background dims.
	if !strings.Contains(body, `viewBox="0 0 400 300"`) {
		t.Errorf("viewBox should be bg dims: %s", body)
	}
	// Digit data URI present.
	if !strings.Contains(body, `data:image/gif;base64,QQ`) {
		t.Errorf("missing digit data URI: %s", body)
	}
}

func TestCounterUnknownBackground400(t *testing.T) {
	s := newBgServer(t)
	req := httptest.NewRequest(http.MethodGet, "/@bgtest?theme=loli&bg=nonexistent", nil)
	resp, _ := s.app.Test(req)
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("unknown bg should 400, got %d", resp.StatusCode)
	}
}

func TestCounterNoBgFallsBackToRender(t *testing.T) {
	s := newBgServer(t)
	req := httptest.NewRequest(http.MethodGet, "/@nobg?theme=loli", nil)
	resp, _ := s.app.Test(req)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status: %d", resp.StatusCode)
	}
	body := readBody(t, resp)
	// Pure-digit mode: no external bg URL.
	if strings.Contains(body, "cdn.example.com") {
		t.Errorf("non-bg mode should not reference bg URL: %s", body)
	}
}

func TestDemoWithBackgroundLongCache(t *testing.T) {
	s := newBgServer(t)
	req := httptest.NewRequest(http.MethodGet, "/@demo?theme=loli&bg=loli-stand", nil)
	resp, _ := s.app.Test(req)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status: %d", resp.StatusCode)
	}
	if cc := resp.Header.Get("Cache-Control"); !strings.Contains(cc, "max-age=31536000") {
		t.Errorf("demo+bg should still be long cache: %q", cc)
	}
}
