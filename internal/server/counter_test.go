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
	"github.com/miaoledor/lolicount/internal/imgcore/composer"
	"github.com/miaoledor/lolicount/internal/imgcore/render"
	"github.com/miaoledor/lolicount/internal/imgcore/theme"
	"github.com/miaoledor/lolicount/internal/store"
)

// stubRegistry is an in-memory composer.ThemeRegistry for handler tests.
// It stores *theme.Theme directly, matching the unified registry
// interface. Themes are built via makeCardTheme for convenience.
type stubRegistry struct {
	themes map[string]*theme.Theme
}

func (s *stubRegistry) Get(name string) (*theme.Theme, bool) {
	t, ok := s.themes[name]
	return t, ok
}

func (s *stubRegistry) List() []composer.ThemeEntry {
	var out []composer.ThemeEntry
	for name := range s.themes {
		out = append(out, composer.ThemeEntry{Name: name})
	}
	return out
}

// makeCardFrames creates n uniform 10x20 frames as ImageLayers.
func makeCardFrames(n int) []render.ImageLayer {
	frames := make([]render.ImageLayer, n)
	for i := 0; i < n; i++ {
		frames[i] = render.ImageLayer{
			Src:       fmt.Sprintf("data:image/gif;base64,F%d", i),
			Width:     10,
			Height:    20,
			Transform: imgcore.DefaultTransform(),
		}
	}
	return frames
}

// makeCardTheme builds a *theme.Theme for a frame theme (single or multi-frame).
func makeCardTheme(name string, nFrames int) *theme.Theme {
	frames := makeCardFrames(nFrames)
	frame := frames[0]
	return &theme.Theme{
		Name:   name,
		Canvas: theme.Canvas{Width: frame.Width, Height: frame.Height},
		BgW:    frame.Width,
		BgH:    frame.Height,
		Layers: []imgcore.Layer{&frame},
	}
}

// newCounterServer builds a Server with a single fake "lian" theme of 3
// uniform 10x20 frames and a real counter buffer backed by an in-memory
// SQLite store.
func newCounterServer(t *testing.T) *Server {
	t.Helper()
	th := makeCardTheme("lian", 3)
	reg := &stubRegistry{themes: map[string]*theme.Theme{"lian": th}}

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



// TestCounterTextTemplate verifies that the ?text= parameter replaces
// {n} with the count number and preserves surrounding characters.
func TestCounterTextTemplate(t *testing.T) {
	s := newCounterServer(t)
	req := httptest.NewRequest(http.MethodGet, "/@demo?theme=lian&number=42&text=views:{n}!", nil)
	resp, err := s.app.Test(req)
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status: %d", resp.StatusCode)
	}
	body := readBody(t, resp)
	if !strings.Contains(body, `>views:42!<`) {
		t.Errorf("text template not rendered: %s", sub(body, "text"))
	}
}

// TestCounterTextTemplateOverridesUnshowf verifies that setting a text
// template forces the text layer to render even when unshowf=true.
func TestCounterTextTemplateOverridesUnshowf(t *testing.T) {
	s := newCounterServer(t)
	req := httptest.NewRequest(http.MethodGet, "/@demo?theme=lian&number=7&text=n={n}&unshowf=true", nil)
	resp, err := s.app.Test(req)
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status: %d", resp.StatusCode)
	}
	body := readBody(t, resp)
	if !strings.Contains(body, `>n=7<`) {
		t.Errorf("text template should override unshowf: %s", sub(body, "text"))
	}
}

// TestCounterTextTemplateDefaultN verifies that the bare template {n}
// renders the plain count number.
func TestCounterTextTemplateDefaultN(t *testing.T) {
	s := newCounterServer(t)
	req := httptest.NewRequest(http.MethodGet, "/@demo?theme=lian&number=99&text={n}", nil)
	resp, err := s.app.Test(req)
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status: %d", resp.StatusCode)
	}
	body := readBody(t, resp)
	if !strings.Contains(body, `>99<`) {
		t.Errorf("bare {n} template not rendered: %s", sub(body, "text"))
	}
}
