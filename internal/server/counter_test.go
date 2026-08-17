package server

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/rs/zerolog"

	"github.com/miaoledor/lolicount/internal/config"
	"github.com/miaoledor/lolicount/internal/theme"
)

// stubRegistry is an in-memory Registry for handler tests.
type stubRegistry struct {
	themes map[string]*theme.Theme
}

func (s *stubRegistry) Get(name string) (*theme.Theme, bool) {
	t, ok := s.themes[name]
	return t, ok
}
func (s *stubRegistry) List() []string {
	out := make([]string, 0, len(s.themes))
	for k := range s.themes {
		out = append(out, k)
	}
	return out
}

// newCounterServer builds a Server with a single fake "loli" theme of 3
// uniform 10x20 frames.
func newCounterServer(t *testing.T) *Server {
	t.Helper()
	th := &theme.Theme{Name: "loli", Frames: make([]theme.Frame, 3)}
	for i := 0; i < 3; i++ {
		th.Frames[i] = theme.Frame{Width: 10, Height: 20, Data: "data:image/gif;base64,QQ"}
	}
	reg := &stubRegistry{themes: map[string]*theme.Theme{"loli": th}}
	cfg := &config.Config{Host: "127.0.0.1", Port: 0, DBInterval: 10}
	return New(cfg, zerolog.Nop(), reg)
}

func TestCounterDemoSVG(t *testing.T) {
	s := newCounterServer(t)
	req := httptest.NewRequest(http.MethodGet, "/@demo?theme=loli", nil)
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
	if cc := resp.Header.Get("Cache-Control"); cc != "no-store" {
		t.Errorf("Cache-Control: %q want no-store", cc)
	}
	body := readBody(t, resp)
	if !strings.HasPrefix(body, "<?xml") || !strings.Contains(body, "<svg") {
		t.Errorf("body is not SVG: %q", trunc(body, 80))
	}
	// Frame 10 x (20 + 24) = 10 x 44.
	if !strings.Contains(body, `viewBox="0 0 20 44"`) {
		t.Errorf("viewBox wrong: %s", sub(body, "viewBox"))
	}
}

func TestCounterNumberSelectsFrame(t *testing.T) {
	s := newCounterServer(t)
	// number=2 selects frame 2; count text shows 2.
	req := httptest.NewRequest(http.MethodGet, "/@demo?theme=loli&number=2", nil)
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
	req := httptest.NewRequest(http.MethodGet, "/get/@demo?theme=loli", nil)
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
	// number negative -> validation error.
	req := httptest.NewRequest(http.MethodGet, "/@demo?theme=loli&number=-1", nil)
	resp, err := s.app.Test(req)
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("status: %d want 400", resp.StatusCode)
	}
}

func TestCounterRandomTheme(t *testing.T) {
	s := newCounterServer(t)
	req := httptest.NewRequest(http.MethodGet, "/@demo?theme=random", nil)
	resp, err := s.app.Test(req)
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status: %d", resp.StatusCode)
	}
}

func TestCounterDefaultTheme(t *testing.T) {
	s := newCounterServer(t)
	req := httptest.NewRequest(http.MethodGet, "/@demo", nil)
	resp, err := s.app.Test(req)
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status: %d", resp.StatusCode)
	}
}

// helpers
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

var _ = fiber.StatusOK

// A huge number must be rejected (4xx) rather than rendering an
// overlong text that overflows the frame.
func TestCounterHugeNumber400(t *testing.T) {
	s := newCounterServer(t)
	req := httptest.NewRequest(http.MethodGet, "/@demo?theme=loli&number=999999999999", nil)
	resp, err := s.app.Test(req)
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}
	if resp.StatusCode < 400 || resp.StatusCode >= 500 {
		t.Errorf("status: %d want 4xx", resp.StatusCode)
	}
}

// number=0 is the documented default and must render frame 0 with text "0".
func TestCounterNumberZeroDefault(t *testing.T) {
	s := newCounterServer(t)
	req := httptest.NewRequest(http.MethodGet, "/@demo?theme=loli&number=0", nil)
	resp, err := s.app.Test(req)
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status: %d", resp.StatusCode)
	}
	body := readBody(t, resp)
	if !strings.Contains(body, `>0<`) {
		t.Errorf("expected text 0: %s", sub(body, "text"))
	}
}

// number equal to frame count wraps via modulo (10 % 3 = 1) and still
// returns 200 with the number text shown.
func TestCounterNumberWrapsModulo(t *testing.T) {
	s := newCounterServer(t) // 3 frames
	req := httptest.NewRequest(http.MethodGet, "/@demo?theme=loli&number=3", nil)
	resp, err := s.app.Test(req)
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status: %d", resp.StatusCode)
	}
	body := readBody(t, resp)
	if !strings.Contains(body, `>3<`) {
		t.Errorf("expected text 3: %s", sub(body, "text"))
	}
}
