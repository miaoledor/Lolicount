package server

import (
	"io/fs"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/miaoledor/lolicount/assets"
)

// About stays a standalone usage guide with live-stat placeholders in SSG HTML.
func TestFrontendServesAboutPage(t *testing.T) {
	if _, err := fs.Stat(assets.DistFS, "dist/about/index.html"); err != nil {
		t.Skip("build the frontend and copy web/dist into assets/dist to test the about route")
	}
	s := newCounterServer(t)
	for _, route := range []string{"/about", "/about/"} {
		t.Run(route, func(t *testing.T) {
			resp, err := s.app.Test(httptest.NewRequest(http.MethodGet, route, nil))
			if err != nil {
				t.Fatalf("app.Test: %v", err)
			}
			if resp.StatusCode != http.StatusOK {
				t.Fatalf("status: %d, want 200", resp.StatusCode)
			}
			if got := resp.Header.Get("Cache-Control"); got != "no-store" {
				t.Errorf("HTML cache policy: %q, want no-store", got)
			}
			body := readBody(t, resp)
			for _, marker := range []string{`id="about-title"`, `data-stat="total"`, `data-stat="maximum"`, `id="howto"`, "your-name"} {
				if !strings.Contains(body, marker) {
					t.Errorf("about HTML missing %q", marker)
				}
			}
			if got := strings.Count(body, `class="howto-step"`); got != 3 {
				t.Errorf("usage steps: %d, want 3", got)
			}
			for _, removed := range []string{`class="about-hero"`, `class="socials"`, "about-plea", `class="home-footer"`} {
				if strings.Contains(body, removed) {
					t.Errorf("about page still contains removed section %q", removed)
				}
			}
		})
	}
}
