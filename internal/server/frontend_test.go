package server

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

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
