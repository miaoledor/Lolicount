package server

import (
	"io/fs"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/miaoledor/lolicount/assets"
)

// Support must be served from its own prerendered document in the single binary.
func TestFrontendServesSupportPage(t *testing.T) {
	if _, err := fs.Stat(assets.DistFS, "dist/support/index.html"); err != nil {
		t.Skip("build the frontend and copy web/dist into assets/dist to test the support route")
	}
	s := newCounterServer(t)
	for _, route := range []string{"/support", "/support/"} {
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
			for _, marker := range []string{"support-page", "support-title", "jar-fallback", "https://github.com/sponsors/miaoledor"} {
				if !strings.Contains(body, marker) {
					t.Errorf("support HTML missing %q; route may have fallen back to the home page", marker)
				}
			}
		})
	}
}
