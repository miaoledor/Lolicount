package server

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/rs/zerolog"

	"github.com/miaoledor/lolicount/internal/config"
)

// newTestServer builds a Server on an ephemeral port without listening.
func newTestServer(t *testing.T) *Server {
	t.Helper()
	cfg := &config.Config{Host: "127.0.0.1", Port: 0, DBInterval: 10}
	return New(cfg, zerolog.Nop(), nil, nil, nil)
}

func TestHeartbeatStatusAndNoStore(t *testing.T) {
	s := newTestServer(t)
	req := httptest.NewRequest(http.MethodGet, "/heart-beat", nil)
	resp, err := s.app.Test(req)
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusOK)
	}
	if cc := resp.Header.Get("Cache-Control"); cc != "no-store" {
		t.Errorf("Cache-Control: got %q want no-store", cc)
	}
	if ct := resp.Header.Get("Content-Type"); !strings.HasPrefix(ct, "application/json") {
		t.Errorf("Content-Type: got %q want application/json*", ct)
	}
}

func TestHeartbeatBody(t *testing.T) {
	s := newTestServer(t)
	req := httptest.NewRequest(http.MethodGet, "/heart-beat", nil)
	resp, err := s.app.Test(req)
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}
	body, _ := io.ReadAll(resp.Body)
	var got map[string]any
	if err := json.Unmarshal(body, &got); err != nil {
		t.Fatalf("unmarshal: %v (body=%s)", err, body)
	}
	if got["status"] != "alive" {
		t.Errorf("status field: got %v want alive", got["status"])
	}
	if ts, ok := got["timestamp"].(string); !ok || ts == "" {
		t.Errorf("timestamp missing or empty: %v", got["timestamp"])
	}
}

// Unknown non-asset paths are served the SPA fallback (index.html) so
// Vue Router can handle client-side routing. This is intentional since
// the embedded Nuxt SSG dist was registered as a catch-all (M8). A 404
// here would break deep links into the single-page front-end.
func TestUnknownRouteSPAFallback(t *testing.T) {
	s := newTestServer(t)
	req := httptest.NewRequest(http.MethodGet, "/does-not-exist", nil)
	resp, err := s.app.Test(req)
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("status: got %d want 200 (SPA fallback)", resp.StatusCode)
	}
	body, _ := io.ReadAll(resp.Body)
	if !strings.Contains(string(body), "<html") {
		t.Errorf("SPA fallback should serve index.html, got: %q", body[:min(64, len(body))])
	}
}

func TestMethodNotAllowed(t *testing.T) {
	s := newTestServer(t)
	req := httptest.NewRequest(http.MethodPost, "/heart-beat", nil)
	resp, err := s.app.Test(req)
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}
	if resp.StatusCode != http.StatusMethodNotAllowed {
		t.Errorf("status: got %d want 405", resp.StatusCode)
	}
}

// Shutdown on a server that never listened must not panic; it lets main
// call it unconditionally in error paths.
func TestShutdownNeverListened(t *testing.T) {
	s := newTestServer(t)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if err := s.Shutdown(ctx); err != nil {
		t.Fatalf("Shutdown on unstarted server: %v", err)
	}
}
