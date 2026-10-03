package server

import (
	"io/fs"
	"net/http"
	"net/http/httptest"
	"path"
	"regexp"
	"strings"
	"testing"

	"github.com/miaoledor/lolicount/assets"
)

// distHasIndex reports whether the embedded dist contains index.html.
// Tests that depend on serving the SSG frontend skip when it is absent
// (e.g. local runs without a prior `pnpm generate`; CI builds the dist
// before testing so this passes there).
func distHasIndex() bool {
	_, err := fs.Sub(assets.DistFS, "dist")
	if err != nil {
		return false
	}
	f, err := assets.DistFS.Open("dist/index.html")
	if err != nil {
		return false
	}
	f.Close()
	return true
}

// rewriteBaseUrl replaces the baked baseUrl payload value with the runtime
// BASE_URL so a single image can be re-pointed at any domain without a
// rebuild (build once, configure per env).
func TestRewriteBaseUrl(t *testing.T) {
	cases := []struct {
		name    string
		html    string
		baseURL string
		want    string
	}{
		{"empty baked", `config={public:{apiBase:"",baseUrl:""}}`, "https://lolicount.top", `config={public:{apiBase:"",baseUrl:"https://lolicount.top"}`},
		{"existing domain", `baseUrl:"https://old.example.com"},app`, "https://new.example.com", `baseUrl:"https://new.example.com"},app`},
		{"with path-like", `x baseUrl:"" y`, "http://x.io", `x baseUrl:"http://x.io" y`},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := rewriteBaseUrl([]byte(tc.html), tc.baseURL)
			if got == nil {
				t.Fatal("expected non-nil")
			}
			if !strings.Contains(string(got), tc.want) {
				t.Errorf("got %q, want to contain %q", string(got), tc.want)
			}
		})
	}
}

// When there is no baseUrl marker, rewriteBaseUrl returns nil so the
// caller can serve the original HTML untouched.
func TestRewriteBaseUrlNoMarker(t *testing.T) {
	out := rewriteBaseUrl([]byte("<html>no payload here</html>"), "https://x.io")
	if out != nil {
		t.Errorf("expected nil when no marker, got %q", string(out))
	}
}

// The served index.html reflects the runtime BASE_URL: a server built with
// an empty baked baseUrl serves a page whose payload carries the runtime
// domain, so embed links work without rebuilding the image.
func TestIndexHTMLRuntimeBaseUrlOverride(t *testing.T) {
	if !distHasIndex() {
		t.Skip("assets/dist has no index.html; run `pnpm generate` to test frontend serving")
	}
	s := newCounterServer(t)
	s.cfg.BaseURL = "https://runtime.example.com"

	req := httptest.NewRequest(http.MethodGet, "/some-spa-route", nil)
	resp, err := s.app.Test(req)
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status: %d", resp.StatusCode)
	}
	body := readBody(t, resp)
	if !strings.Contains(body, `baseUrl:"https://runtime.example.com"`) {
		t.Errorf("served index.html should carry runtime baseUrl, got: %s", body[:min(200, len(body))])
	}
}

// The home page IS the playground (landing-page swap): it must render
// the param panel, and the removed showcase button must not resurface.
// /themes stays alive as a prerendered alias for old shared ?theme=
// links. The editor quick panel keeps its border-box override — without
// it the mobile panel overflows by its horizontal padding (no global
// CSS reset is loaded).
func TestHomeIsPlayground(t *testing.T) {
	if !distHasIndex() {
		t.Skip("assets/dist has no index.html; run `pnpm generate` to test frontend serving")
	}

	s := newCounterServer(t)
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	resp, err := s.app.Test(req)
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status: %d", resp.StatusCode)
	}
	body := readBody(t, resp)
	if !strings.Contains(body, "loli-tool") {
		t.Error("home page does not render the playground param panel")
	}
	if strings.Contains(body, "showcase-playground-link") {
		t.Error("home page still renders the removed showcase browse button")
	}

	aliasReq := httptest.NewRequest(http.MethodGet, "/themes", nil)
	aliasResp, err := s.app.Test(aliasReq)
	if err != nil {
		t.Fatalf("alias app.Test: %v", err)
	}
	if !strings.Contains(readBody(t, aliasResp), "loli-tool") {
		t.Error("/themes alias does not render the playground")
	}

	dist, err := fs.Sub(assets.DistFS, "dist")
	if err != nil {
		t.Fatalf("open dist: %v", err)
	}
	if _, err := fs.Stat(dist, "_nuxt"); err != nil {
		t.Skip("assets/dist has no _nuxt directory; run `pnpm generate` to test built CSS")
	}

	rule := regexp.MustCompile(`\.quick-panel[^{}]*\{[^}]*box-sizing\s*:\s*border-box`)
	found := false
	err = fs.WalkDir(dist, "_nuxt", func(p string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if d.IsDir() || path.Ext(p) != ".css" {
			return nil
		}
		css, err := fs.ReadFile(dist, p)
		if err != nil {
			return err
		}
		if rule.Match(css) {
			found = true
		}
		return nil
	})
	if err != nil {
		t.Fatalf("walk built CSS: %v", err)
	}
	if !found {
		t.Error("built CSS is missing border-box sizing for .quick-panel")
	}
}

// Prerendered Nuxt pages are directories in dist (editor/index.html etc.).
// Both the bare and trailing-slash forms must serve that page's HTML, not
// the home SPA fallback (which would navigate the client back to "/").
func TestFrontendServesPrerenderedSubPages(t *testing.T) {
	sub, err := fs.Sub(assets.DistFS, "dist")
	if err != nil {
		t.Skip("assets/dist unavailable")
	}
	if _, err := fs.Stat(sub, "editor/index.html"); err != nil {
		t.Skip("assets/dist has no editor/index.html; run `pnpm generate` to test frontend serving")
	}
	s := newCounterServer(t)
	for _, route := range []string{"/editor", "/editor/"} {
		req := httptest.NewRequest(http.MethodGet, route, nil)
		resp, err := s.app.Test(req)
		if err != nil {
			t.Fatalf("%s: app.Test: %v", route, err)
		}
		if resp.StatusCode != http.StatusOK {
			t.Errorf("%s: status %d, want 200", route, resp.StatusCode)
			continue
		}
		body := readBody(t, resp)
		if strings.Contains(body, "Card Themes") {
			t.Errorf("%s: body is the home SPA fallback, not the prerendered page", route)
		}
	}
}

// Hashed _nuxt assets are immutable; HTML entry points must revalidate so
// a redeploy is picked up instead of serving a stale page that references
// deleted chunks; gallery images get a bounded window (they change only
// with a theme rebuild under a name-stable path).
func TestFrontendCachePolicy(t *testing.T) {
	s := newCounterServer(t)
	req := httptest.NewRequest(http.MethodGet, "/some-spa-route", nil)
	resp, err := s.app.Test(req)
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}
	if cc := resp.Header.Get("Cache-Control"); cc != "no-store" {
		t.Errorf("html Cache-Control: got %q want no-store", cc)
	}

	// A theme thumbnail must carry the bounded image window.
	thumb := "images/theme-thumbs/cafestella-yuna.webp"
	if _, err := fs.Stat(assets.DistFS, "dist/"+thumb); err != nil {
		t.Skipf("assets/dist has no %s; run `pnpm generate` to test frontend serving", thumb)
	}
	req = httptest.NewRequest(http.MethodGet, "/"+thumb, nil)
	resp, err = s.app.Test(req)
	if err != nil {
		t.Fatalf("thumb app.Test: %v", err)
	}
	if cc := resp.Header.Get("Cache-Control"); cc != "public, max-age=86400" {
		t.Errorf("thumb Cache-Control: got %q want public, max-age=86400", cc)
	}
}
