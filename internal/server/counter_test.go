package server

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/rs/zerolog"

	"github.com/miaoledor/lolicount/internal/config"
	"github.com/miaoledor/lolicount/internal/counter"
	"github.com/miaoledor/lolicount/internal/imgcore"
	"github.com/miaoledor/lolicount/internal/imgcore/cardthemedrawer"
	"github.com/miaoledor/lolicount/internal/imgcore/characterthemedrawer"
	"github.com/miaoledor/lolicount/internal/imgcore/renderer"
	"github.com/miaoledor/lolicount/internal/store"
)

// stubRegistry is an in-memory renderer.ThemeRegistry for handler tests.
type stubRegistry struct {
	cards      map[string]*cardthemedrawer.Theme
	characters map[string]*characterthemedrawer.Character
}

func (s *stubRegistry) GetCard(name string) (*cardthemedrawer.Theme, bool) {
	t, ok := s.cards[name]
	return t, ok
}

func (s *stubRegistry) GetCharacter(name string) (*characterthemedrawer.Character, bool) {
	c, ok := s.characters[name]
	return c, ok
}

func (s *stubRegistry) Get(name string) (renderer.ThemeEntry, bool) {
	if _, ok := s.cards[name]; ok {
		return renderer.ThemeEntry{Name: name, Kind: imgcore.LegacyKindFrame}, true
	}
	if _, ok := s.characters[name]; ok {
		return renderer.ThemeEntry{Name: name, Kind: imgcore.LegacyKindCharacter}, true
	}
	return renderer.ThemeEntry{}, false
}

func (s *stubRegistry) List() []renderer.ThemeEntry {
	var out []renderer.ThemeEntry
	for k := range s.cards {
		out = append(out, renderer.ThemeEntry{Name: k, Kind: imgcore.LegacyKindFrame})
	}
	for k := range s.characters {
		out = append(out, renderer.ThemeEntry{Name: k, Kind: imgcore.LegacyKindCharacter})
	}
	return out
}

// newCounterServer builds a Server with a single fake "lian" theme of 3
// uniform 10x20 frames and a real counter buffer backed by an in-memory
// SQLite store.
func newCounterServer(t *testing.T) *Server {
	t.Helper()
	th := &cardthemedrawer.Theme{Name: "lian", Frames: make([]cardthemedrawer.Frame, 3)}
	for i := 0; i < 3; i++ {
		th.Frames[i] = cardthemedrawer.Frame{Width: 10, Height: 20, Data: fmt.Sprintf("data:image/gif;base64,F%d", i)}
	}
	reg := &stubRegistry{cards: map[string]*cardthemedrawer.Theme{"lian": th}}

	repo, err := store.NewSQLite(context.Background(), ":memory:")
	if err != nil {
		t.Fatalf("NewSQLite: %v", err)
	}
	t.Cleanup(func() {
		if c, ok := repo.(interface{ Close() error }); ok {
			c.Close()
		}
	})
	buf := counter.New(repo, zerolog.Nop(), 3600)
	if err := buf.Start(context.Background()); err != nil {
		t.Fatalf("buffer start: %v", err)
	}
	t.Cleanup(buf.Stop)

	cfg := &config.Config{Host: "127.0.0.1", Port: 0, DBInterval: 10, RateLimitIPPerSec: 10000, RateLimitIPPerMin: 100000, RateLimitNamePerSec: 10000}
	s := New(cfg, zerolog.Nop(), reg, nil, buf)
	t.Cleanup(func() {
		s.ipLimiter.Stop()
		s.nameLimiter.Stop()
	})
	return s
}

func TestCounterDemoSVG(t *testing.T) {
	s := newCounterServer(t)
	req := httptest.NewRequest(http.MethodGet, "/@demo?theme=lian", nil)
	resp, err := s.app.Test(req)
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status: %d", resp.StatusCode)
	}
	if ct := resp.Header.Get("Content-Type"); !strings.HasPrefix(ct, "image/svg+xml") {
		t.Errorf("Content-Type: %q", ct)
	}
	// Iron Rule 1: demo is the ONLY path allowed long cache.
	if cc := resp.Header.Get("Cache-Control"); !strings.Contains(cc, "max-age=31536000") {
		t.Errorf("demo Cache-Control: %q want max-age=31536000", cc)
	}
	body := readBody(t, resp)
	if !strings.HasPrefix(body, "<?xml") || !strings.Contains(body, "<svg") {
		t.Errorf("body is not SVG: %q", trunc(body, 80))
	}
	if !strings.Contains(body, `viewBox="0 0 200 420"`) {
		t.Errorf("viewBox wrong: %s", sub(body, "viewBox"))
	}
}

func TestCounterNumberShowsValue(t *testing.T) {
	s := newCounterServer(t)
	req := httptest.NewRequest(http.MethodGet, "/@demo?theme=lian&number=2", nil)
	resp, err := s.app.Test(req)
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status: %d", resp.StatusCode)
	}
	body := readBody(t, resp)
	if !strings.Contains(body, `>2<`) {
		t.Errorf("number text 2 missing: %s", sub(body, "text"))
	}
}

func TestCounterGetAlias(t *testing.T) {
	s := newCounterServer(t)
	req := httptest.NewRequest(http.MethodGet, "/get/@demo?theme=lian", nil)
	resp, err := s.app.Test(req)
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status: %d", resp.StatusCode)
	}
}

func TestCounterUnknownTheme400(t *testing.T) {
	s := newCounterServer(t)
	req := httptest.NewRequest(http.MethodGet, "/@demo?theme=nonexistent", nil)
	resp, err := s.app.Test(req)
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("status: %d want 400", resp.StatusCode)
	}
}

func TestCounterInvalidParam400(t *testing.T) {
	s := newCounterServer(t)
	req := httptest.NewRequest(http.MethodGet, "/@demo?theme=lian&number=-1", nil)
	resp, err := s.app.Test(req)
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("status: %d want 400", resp.StatusCode)
	}
}

// readBody reads the full response body as a string.
func readBody(t *testing.T, resp *http.Response) string {
	t.Helper()
	b := make([]byte, 0, 4096)
	buf := make([]byte, 4096)
	for {
		n, err := resp.Body.Read(buf)
		if n > 0 {
			b = append(b, buf[:n]...)
		}
		if err != nil {
			break
		}
	}
	return string(b)
}

func trunc(s string, n int) string {
	if len(s) > n {
		return s[:n]
	}
	return s
}

func sub(s, marker string) string {
	i := strings.Index(s, marker)
	if i < 0 {
		return "(not found)"
	}
	end := i + 60
	if end > len(s) {
		end = len(s)
	}
	return s[i:end]
}
