package server

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/miaoledor/lolicount/internal/config"
	"github.com/miaoledor/lolicount/internal/imgcore/theme"
	"github.com/rs/zerolog"
)

// multiFrameServer builds a Server whose "lian" theme is multi-frame
// (RandomPickLayer), so demo requests exercise the no-store branch of
// the cache policy (AGENTS.md Iron Rule 1: multi-frame demo is random
// per request and must NOT be long-cached).
func multiFrameServer(t *testing.T) *Server {
	t.Helper()
	th := makeMultiFrameTheme("lian", []struct{ W, H int }{{10, 20}, {10, 20}})
	reg := &stubRegistry{themes: map[string]*theme.Theme{"lian": th}}
	cfg := &config.Config{
		Host: "127.0.0.1", Port: 0, DBInterval: 10,
		RateLimitIPPerSec: 10000, RateLimitIPPerMin: 100000, RateLimitNamePerSec: 10000,
	}
	s := New(cfg, zerolog.Nop(), reg, nil, nil)
	t.Cleanup(func() {
		s.ipLimiter.Stop()
		s.nameLimiter.Stop()
	})
	return s
}

// TestDemoMultiFrameNoStore verifies that a demo request against a
// multi-frame theme gets Cache-Control: no-store. The output changes
// per request (random frame selection), so long caching would freeze a
// stale frame in GitHub's image proxy (Iron Rule 1).
func TestDemoMultiFrameNoStore(t *testing.T) {
	s := multiFrameServer(t)
	req := httptest.NewRequest(http.MethodGet, "/@demo?theme=lian", nil)
	resp, err := s.app.Test(req)
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status: %d", resp.StatusCode)
	}
	if cc := resp.Header.Get("Cache-Control"); cc != "no-store" {
		t.Errorf("multi-frame demo Cache-Control: %q want no-store", cc)
	}
}

// TestDemoNumberOverridesMultiFrameCache verifies that demo with an
// explicit number>0 is long-cached even for a multi-frame theme,
// because the output is the fixed number text (deterministic), not a
// random frame. The background frame is still random, but the count
// text is fixed — this mirrors counter.go's branch which treats
// number>0 as deterministic for caching.
func TestDemoNumberLongCacheOnMultiFrame(t *testing.T) {
	s := multiFrameServer(t)
	req := httptest.NewRequest(http.MethodGet, "/@demo?theme=lian&number=42", nil)
	resp, err := s.app.Test(req)
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status: %d", resp.StatusCode)
	}
	// counter.go: name=="demo" && (q.Number>0 || !multiFrame) -> long cache.
	// number>0 wins even when the theme is multi-frame.
	if cc := resp.Header.Get("Cache-Control"); !strings.Contains(cc, "max-age=31536000") {
		t.Errorf("demo+number Cache-Control: %q want max-age=31536000", cc)
	}
}

// adminServer builds a Server with an optional ADMIN_KEY so adminAuth
// behavior can be tested across the empty / wrong / correct key cases.
func adminServer(t *testing.T, adminKey string) *Server {
	t.Helper()
	cfg := &config.Config{
		Host: "127.0.0.1", Port: 0, DBInterval: 10,
		RateLimitIPPerSec: 10000, RateLimitIPPerMin: 100000, RateLimitNamePerSec: 10000,
		AdminKey: adminKey,
	}
	s := New(cfg, zerolog.Nop(), nil, nil, nil)
	t.Cleanup(func() {
		s.ipLimiter.Stop()
		s.nameLimiter.Stop()
	})
	return s
}

// TestAdminEmptyKeyReturns404 verifies that when ADMIN_KEY is unset,
// admin endpoints return 404 (invisible) rather than 401/403. This
// avoids leaking the existence of admin functionality (middleware.go).
func TestAdminEmptyKeyReturns404(t *testing.T) {
	s := adminServer(t, "")
	req := httptest.NewRequest(http.MethodGet, "/api/admin/ping", nil)
	resp, err := s.app.Test(req)
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("empty ADMIN_KEY: got %d want 404 (invisible)", resp.StatusCode)
	}
}

// TestAdminWrongKeyReturns403 verifies that a non-empty ADMIN_KEY with
// a mismatched X-Admin-Key header returns 403.
func TestAdminWrongKeyReturns403(t *testing.T) {
	s := adminServer(t, "secret-key")
	req := httptest.NewRequest(http.MethodGet, "/api/admin/ping", nil)
	req.Header.Set("X-Admin-Key", "wrong")
	resp, err := s.app.Test(req)
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}
	if resp.StatusCode != http.StatusForbidden {
		t.Errorf("wrong key: got %d want 403", resp.StatusCode)
	}
}

// TestAdminMissingKeyReturns403 verifies that a non-empty ADMIN_KEY
// with no X-Admin-Key header at all returns 403.
func TestAdminMissingKeyReturns403(t *testing.T) {
	s := adminServer(t, "secret-key")
	req := httptest.NewRequest(http.MethodGet, "/api/admin/ping", nil)
	resp, err := s.app.Test(req)
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}
	if resp.StatusCode != http.StatusForbidden {
		t.Errorf("missing key: got %d want 403", resp.StatusCode)
	}
}

// TestAdminCorrectKeyReturns200 verifies that the correct X-Admin-Key
// header grants access to admin endpoints.
func TestAdminCorrectKeyReturns200(t *testing.T) {
	s := adminServer(t, "secret-key")
	req := httptest.NewRequest(http.MethodGet, "/api/admin/ping", nil)
	req.Header.Set("X-Admin-Key", "secret-key")
	resp, err := s.app.Test(req)
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("correct key: got %d want 200", resp.StatusCode)
	}
}
