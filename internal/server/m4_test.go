package server

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/rs/zerolog"

	"github.com/miaoledor/lolicount/internal/config"
	"github.com/miaoledor/lolicount/internal/counter"
	"github.com/miaoledor/lolicount/internal/store"
	"github.com/miaoledor/lolicount/internal/theme"
)

// m4Server builds a server with a 1-frame theme and a real buffer, with
// rate limits intentionally low so tests can trip them.
func m4Server(t *testing.T, ipSec, ipMin, nameSec int) *Server {
	t.Helper()
	th := &theme.Theme{Name: "loli", Frames: []theme.Frame{{Width: 10, Height: 20, Data: "data:image/gif;base64,QQ"}}}
	reg := &stubRegistry{themes: map[string]*theme.Theme{"loli": th}}
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
	cfg := &config.Config{
		Host: "127.0.0.1", Port: 0, DBInterval: 10,
		RateLimitIPPerSec: ipSec, RateLimitIPPerMin: ipMin, RateLimitNamePerSec: nameSec,
	}
	s := New(cfg, zerolog.Nop(), reg, buf, nil)
	t.Cleanup(func() {
		s.ipLimiter.Stop()
		s.nameLimiter.Stop()
	})
	return s
}

// TestIPLimitReturns429: bursts past the IP per-second rate get 429.
func TestIPLimitReturns429(t *testing.T) {
	s := m4Server(t, 2, 1000, 1000) // 2 req/s IP, generous name
	for i := 0; i < 2; i++ {
		req := httptest.NewRequest(http.MethodGet, "/@iplimit?theme=loli", nil)
		resp, err := s.app.Test(req)
		if err != nil {
			t.Fatalf("iter %d: %v", i, err)
		}
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("iter %d: expected 200, got %d", i, resp.StatusCode)
		}
	}
	// 3rd request within the same second should be 429.
	req := httptest.NewRequest(http.MethodGet, "/@iplimit?theme=loli", nil)
	resp, err := s.app.Test(req)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusTooManyRequests {
		t.Errorf("expected 429, got %d", resp.StatusCode)
	}
}

// TestNameLimitDegradesReadOnly: a name past 5/s must still return 200
// with the CURRENT count (not +1), never 429 (Iron Rule 3).
func TestNameLimitDegradesReadOnly(t *testing.T) {
	s := m4Server(t, 1000, 100000, 2) // generous IP, 2/s name
	// Two increments: count becomes 1 then 2.
	for i := 1; i <= 2; i++ {
		req := httptest.NewRequest(http.MethodGet, "/@degrade?theme=loli", nil)
		resp, _ := s.app.Test(req)
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("iter %d: expected 200, got %d", i, resp.StatusCode)
		}
	}
	// 3rd request same second: degraded. Must be 200, text "2" (not 3).
	req := httptest.NewRequest(http.MethodGet, "/@degrade?theme=loli", nil)
	resp, err := s.app.Test(req)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("degraded request must be 200 (not 429), got %d", resp.StatusCode)
	}
	body := readBody(t, resp)
	if !strings.Contains(body, ">2<") {
		t.Errorf("degraded count should be 2 (no increment): %s", sub(body, "text"))
	}
}

// TestDemoLongCache: demo gets max-age=31536000, real counter no-store.
func TestDemoLongCache(t *testing.T) {
	s := m4Server(t, 1000, 100000, 1000)
	demo := httptest.NewRequest(http.MethodGet, "/@demo?theme=loli", nil)
	dresp, _ := s.app.Test(demo)
	if cc := dresp.Header.Get("Cache-Control"); !strings.Contains(cc, "max-age=31536000") {
		t.Errorf("demo Cache-Control: %q want max-age=31536000", cc)
	}
	real := httptest.NewRequest(http.MethodGet, "/@realcache?theme=loli", nil)
	rresp, _ := s.app.Test(real)
	if cc := rresp.Header.Get("Cache-Control"); cc != "no-store" {
		t.Errorf("real counter Cache-Control: %q want no-store", cc)
	}
}

// TestCORSSetsHeadersOnAPI: CORS headers appear on /api/* paths.
func TestCORSSetsHeadersOnAPI(t *testing.T) {
	s := m4Server(t, 1000, 100000, 1000)
	req := httptest.NewRequest(http.MethodOptions, "/api/themes", nil)
	req.Header.Set("Origin", "https://example.com")
	resp, err := s.app.Test(req)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusNoContent {
		t.Errorf("OPTIONS /api: expected 204, got %d", resp.StatusCode)
	}
	if ao := resp.Header.Get("Access-Control-Allow-Origin"); ao != "https://example.com" {
		t.Errorf("ACAO: %q want https://example.com", ao)
	}
}

// TestNoCORSOnCounter: counter SVG paths must NOT carry CORS headers.
func TestNoCORSOnCounter(t *testing.T) {
	s := m4Server(t, 1000, 100000, 1000)
	req := httptest.NewRequest(http.MethodGet, "/@nocors?theme=loli", nil)
	req.Header.Set("Origin", "https://example.com")
	resp, _ := s.app.Test(req)
	if ao := resp.Header.Get("Access-Control-Allow-Origin"); ao != "" {
		t.Errorf("counter SVG must not have CORS, got ACAO=%q", ao)
	}
}

// TestNameLimitResetsAfterWindow: after the 1s window expires, the name
// can increment again (degradation is temporary, not permanent).
func TestNameLimitResetsAfterWindow(t *testing.T) {
	s := m4Server(t, 1000, 100000, 1)
	// First request increments to 1.
	s.app.Test(httptest.NewRequest(http.MethodGet, "/@reset?theme=loli", nil))
	// Second within the window: degraded, still 1.
	resp, _ := s.app.Test(httptest.NewRequest(http.MethodGet, "/@reset?theme=loli", nil))
	if !strings.Contains(readBody(t, resp), ">1<") {
		t.Fatal("expected degraded count 1")
	}
	// Wait past the 1s window, then increment again -> 2.
	time.Sleep(1100 * time.Millisecond)
	resp2, _ := s.app.Test(httptest.NewRequest(http.MethodGet, "/@reset?theme=loli", nil))
	if !strings.Contains(readBody(t, resp2), ">2<") {
		t.Errorf("expected count 2 after window reset, got %s", sub(readBody(t, resp2), "text"))
	}
}

// TestInvalidParam400: invalid query (negative number) returns 400.
func TestInvalidParam400(t *testing.T) {
	s := m4Server(t, 1000, 100000, 1000)
	req := httptest.NewRequest(http.MethodGet, "/@demo?theme=loli&number=-5", nil)
	resp, _ := s.app.Test(req)
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("expected 400 for negative number, got %d", resp.StatusCode)
	}
}

// keep io imported if readBody moves
var _ = io.ReadAll
var _ fiber.Ctx
