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

// newCounterServer builds a Server with a single fake "loli" theme whose
// digits are uniform 10x20 glyphs.
func newCounterServer(t *testing.T) *Server {
	t.Helper()
	th := &theme.Theme{Name: "loli", Chars: map[theme.CharName]theme.ThemeChar{}}
	for _, d := range []string{"0", "1", "2", "3", "4", "5", "6", "7", "8", "9"} {
		th.Chars[theme.CharName(d)] = theme.ThemeChar{Width: 10, Height: 20, Data: "data:image/gif;base64,QQ"}
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
	body, _ := readBody(resp)
	if !strings.HasPrefix(body, "<?xml") || !strings.Contains(body, "<svg") {
		t.Errorf("body is not SVG: %q", truncate(body, 80))
	}
	// demo = 0123456789, 10 digits * 10 wide = 100.
	if !strings.Contains(body, `viewBox="0 0 100 20"`) {
		t.Errorf("demo viewBox wrong: %s", substring(body, "viewBox"))
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

func TestCounterNumPreview(t *testing.T) {
	s := newCounterServer(t)
	// num=5, padding=0, prefix=-1 -> 1 digit, width 10.
	req := httptest.NewRequest(http.MethodGet, "/@x?num=5&theme=loli&padding=0&prefix=-1", nil)
	resp, err := s.app.Test(req)
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status: %d", resp.StatusCode)
	}
	body, _ := readBody(resp)
	if !strings.Contains(body, `viewBox="0 0 10 20"`) {
		t.Errorf("num preview viewBox wrong: %s", substring(body, "viewBox"))
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
	// scale out of range (>2).
	req := httptest.NewRequest(http.MethodGet, "/@demo?theme=loli&scale=5", nil)
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
	// No theme param -> defaults to "loli".
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
func readBody(resp *http.Response) (string, error) {
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
	return string(b), nil
}

func truncate(s string, n int) string {
	if len(s) > n {
		return s[:n]
	}
	return s
}

func substring(s, marker string) string {
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

// keep fiber import used (status constants via fiber are not needed here,
// but the package is referenced through app.Test).
var _ = fiber.StatusOK
