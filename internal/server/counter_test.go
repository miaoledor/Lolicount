package server

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/rs/zerolog"

	"context"

	"github.com/miaoledor/lolicount/internal/config"
	"github.com/miaoledor/lolicount/internal/counter"
	"github.com/miaoledor/lolicount/internal/store"
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
// uniform 10x20 frames and a real counter buffer backed by an in-memory
// SQLite store.
func newCounterServer(t *testing.T) *Server {
	t.Helper()
	th := &theme.Theme{Name: "loli", Frames: make([]theme.Frame, 3)}
	for i := 0; i < 3; i++ {
		th.Frames[i] = theme.Frame{Width: 10, Height: 20, Data: "data:image/gif;base64,QQ"}
	}
	reg := &stubRegistry{themes: map[string]*theme.Theme{"loli": th}}

	repo, err := store.NewSQLite(context.Background(), ":memory:")
	if err != nil {
		t.Fatalf("NewSQLite: %v", err)
	}
	t.Cleanup(func() {
		if c, ok := repo.(interface{ Close() error }); ok {
			c.Close()
		}
	})
	buf := counter.New(repo, zerolog.Nop(), 3600) // long interval: no auto-flush in tests
	if err := buf.Start(context.Background()); err != nil {
		t.Fatalf("buffer start: %v", err)
	}
	t.Cleanup(buf.Stop)

	cfg := &config.Config{Host: "127.0.0.1", Port: 0, DBInterval: 10, RateLimitIPPerSec: 10000, RateLimitIPPerMin: 100000, RateLimitNamePerSec: 10000}
	s := New(cfg, zerolog.Nop(), reg, buf)
	t.Cleanup(func() {
		s.ipLimiter.Stop()
		s.nameLimiter.Stop()
	})
	return s
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
	// Iron Rule 1: demo is the ONLY path allowed long cache; real
	// counters stay no-store.
	if cc := resp.Header.Get("Cache-Control"); !strings.Contains(cc, "max-age=31536000") {
		t.Errorf("demo Cache-Control: %q want max-age=31536000", cc)
	}
	body := readBody(t, resp)
	if !strings.HasPrefix(body, "<?xml") || !strings.Contains(body, "<svg") {
		t.Errorf("body is not SVG: %q", trunc(body, 80))
	}
	// M5.5: viewBox height = frame height (20); width grows if text wider.
	if !strings.Contains(body, `viewBox="0 0 18 20"`) {
		t.Errorf("viewBox wrong: %s", sub(body, "viewBox"))
	}
}

func TestCounterNumberShowsValue(t *testing.T) {
	s := newCounterServer(t)
	// M5.5: number controls the displayed text value, not the frame
	// (theme frame is always 0 now).
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

// Multiple requests to the same name must increment the counter, and
// the rendered text must reflect the growing value.
func TestCounterIncrements(t *testing.T) {
	s := newCounterServer(t)
	for i := 1; i <= 5; i++ {
		req := httptest.NewRequest(http.MethodGet, "/@realcount?theme=loli", nil)
		resp, err := s.app.Test(req)
		if err != nil {
			t.Fatalf("app.Test: %v", err)
		}
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("iter %d status: %d", i, resp.StatusCode)
		}
		body := readBody(t, resp)
		want := strconv.Itoa(i)
		if !strings.Contains(body, ">"+want+"<") {
			t.Errorf("iter %d: text %s missing in %s", i, want, sub(body, "text"))
		}
	}
}

// /record/@:name returns the current count as JSON without incrementing.
func TestRecordHandlerJSON(t *testing.T) {
	s := newCounterServer(t)
	// Increment to 3.
	for i := 0; i < 3; i++ {
		req := httptest.NewRequest(http.MethodGet, "/@rec?theme=loli", nil)
		s.app.Test(req)
	}
	// Read via record (no increment).
	req := httptest.NewRequest(http.MethodGet, "/record/@rec", nil)
	resp, err := s.app.Test(req)
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status: %d", resp.StatusCode)
	}
	if ct := resp.Header.Get("Content-Type"); !strings.HasPrefix(ct, "application/json") {
		t.Errorf("Content-Type: %q", ct)
	}
	body := readBody(t, resp)
	var rec struct {
		Name string `json:"name"`
		Num  int64  `json:"num"`
	}
	if err := json.Unmarshal([]byte(body), &rec); err != nil {
		t.Fatalf("unmarshal: %v body=%s", err, body)
	}
	if rec.Name != "rec" || rec.Num != 3 {
		t.Errorf("record = %+v want rec/3", rec)
	}
	// record must not have incremented.
	req2 := httptest.NewRequest(http.MethodGet, "/record/@rec", nil)
	resp2, _ := s.app.Test(req2)
	b2 := readBody(t, resp2)
	var rec2 struct {
		Num int64 `json:"num"`
	}
	json.Unmarshal([]byte(b2), &rec2)
	if rec2.Num != 3 {
		t.Errorf("record after re-read = %d want 3 (no increment)", rec2.Num)
	}
}

// Debug: Incr via /@name then immediately /record/@name must agree.
func TestCounterRecordAgree(t *testing.T) {
	s := newCounterServer(t)
	// Seed: two increments.
	s.app.Test(httptest.NewRequest(http.MethodGet, "/@agree?theme=loli", nil))
	resp, _ := s.app.Test(httptest.NewRequest(http.MethodGet, "/@agree?theme=loli", nil))
	body := readBody(t, resp)
	t.Logf("second @agree SVG text: %s", sub(body, "text"))
	rec, _ := s.app.Test(httptest.NewRequest(http.MethodGet, "/record/@agree", nil))
	rbody := readBody(t, rec)
	t.Logf("record body: %s", rbody)
	if !strings.Contains(rbody, `"num":2`) {
		t.Errorf("record num should be 2: %s", rbody)
	}
}

// M5.5: the theme frame is a pure style background and must NOT change
// with the count. Two different counts must render the same background
// image; only the overlaid text differs.
func TestCounterFrameDoesNotChangeWithCount(t *testing.T) {
	s := newCounterServer(t)
	// Two increments -> counts 1 and 2.
	r1, _ := s.app.Test(httptest.NewRequest(http.MethodGet, "/@framefix?theme=loli", nil))
	b1 := readBody(t, r1)
	r2, _ := s.app.Test(httptest.NewRequest(http.MethodGet, "/@framefix?theme=loli", nil))
	b2 := readBody(t, r2)
	// Extract the <image href=...> (the background). It must be identical
	// across both counts — the theme must not reflect the count.
	img1 := sub(b1, "image href=")
	img2 := sub(b2, "image href=")
	if img1 != img2 {
		t.Errorf("theme background must not change with count:\n  count1: %s\n  count2: %s", img1, img2)
	}
	// But the text must differ (1 vs 2).
	if !strings.Contains(b1, ">1<") || !strings.Contains(b2, ">2<") {
		t.Errorf("text should differ: b1=%s b2=%s", sub(b1, "text"), sub(b2, "text"))
	}
}
