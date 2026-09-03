package server

import (
	"io/fs"
	"net/http"
	"net/http/httptest"
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

// Prerendered Nuxt pages are directories in dist (emote/index.html etc.).
// Both the bare and trailing-slash forms must serve that page's HTML, not
// the home SPA fallback (which would navigate the client back to "/").
func TestFrontendServesPrerenderedSubPages(t *testing.T) {
	sub, err := fs.Sub(assets.DistFS, "dist")
	if err != nil {
		t.Skip("assets/dist unavailable")
	}
	if _, err := fs.Stat(sub, "emote/index.html"); err != nil {
		t.Skip("assets/dist has no emote/index.html; run `pnpm generate` to test frontend serving")
	}
	s := newCounterServer(t)
	for _, route := range []string{"/emote", "/emote/"} {
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
		if !strings.Contains(body, "Emote") {
			t.Errorf("%s: body does not contain the emote page marker", route)
		}
	}
}
