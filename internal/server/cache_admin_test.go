package server

import (
	"net/http"
	"net/http/httptest"
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
	s := multiFrameServer(t)
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
	s := multiFrameServer(t)
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

// adminServer builds a Server with an optional ADMIN_KEY so adminAuth
// behavior can be tested across the empty / wrong / correct key cases.
func adminServer(t *testing.T, adminKey string) *Server {
	t.Helper()
	cfg := &config.Config{
		Host: "127.0.0.1", Port: 0, DBInterval: 10,
		RateLimitIPPerSec: 10000, RateLimitIPPerMin: 100000, RateLimitNamePerSec: 10000,
		AdminKey: adminKey,
	}
	s := New(cfg, zerolog.Nop(), nil, nil, nil, nil)
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
