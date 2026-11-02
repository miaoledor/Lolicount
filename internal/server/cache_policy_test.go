package server

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/miaoledor/lolicount/internal/config"
	"github.com/miaoledor/lolicount/internal/imgcore/theme"
	"github.com/rs/zerolog"
)

// multiFrameCacheServer builds a Server whose "lian" theme is multi-frame
// (RandomPickLayer), so demo requests exercise the no-store branch of
// the cache policy (AGENTS.md Iron Rule 1: multi-frame demo is random
// per request and must NOT be long-cached).
func multiFrameCacheServer(t *testing.T) *Server {
	t.Helper()
	th := makeMultiFrameTheme("lian", []struct{ W, H int }{{10, 20}, {10, 20}})
	reg := &stubRegistry{themes: map[string]*theme.Theme{"lian": th}}
	cfg := &config.Config{
		Host: "127.0.0.1", Port: 0, DBInterval: 10,
		RateLimitIPPerSec: 10000, RateLimitIPPerMin: 100000, RateLimitNamePerSec: 10000,
	}
	s := New(cfg, zerolog.Nop(), reg, nil, nil, nil)
	t.Cleanup(func() {
		s.ipLimiter.Stop()
		s.nameLimiter.Stop()
	})
	return s
}

// TestDemoMultiFrameNoStore verifies that a demo request against a
// multi-frame theme gets the realCountCacheControl combo. The output
// changes per request (random frame selection), so any caching would
// freeze a stale frame in GitHub's image proxy (Iron Rule 1).
func TestDemoMultiFrameNoStore(t *testing.T) {
	s := multiFrameCacheServer(t)
	req := httptest.NewRequest(http.MethodGet, "/@demo?theme=lian", nil)
	resp, err := s.app.Test(req)
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status: %d", resp.StatusCode)
	}
	if cc := resp.Header.Get("Cache-Control"); cc != realCountCacheControl {
		t.Errorf("multi-frame demo Cache-Control: %q want %q", cc, realCountCacheControl)
	}
}

// TestDemoNumberMultiFrameNoStore verifies that number>0 does not make a
// multi-frame demo deterministic. The text is fixed, but the background
// still changes, so the full no-cache combination remains mandatory.
func TestDemoNumberMultiFrameNoStore(t *testing.T) {
	s := multiFrameCacheServer(t)
	req := httptest.NewRequest(http.MethodGet, "/@demo?theme=lian&number=42", nil)
	resp, err := s.app.Test(req)
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status: %d", resp.StatusCode)
	}
	if cc := resp.Header.Get("Cache-Control"); cc != realCountCacheControl {
		t.Errorf("multi-frame demo+number Cache-Control: %q want %q", cc, realCountCacheControl)
	}
}
