package server

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

// TestExtractedEditorAPIRoutesStayUnregistered protects the repository
// boundary after the editor service moved to its standalone repository.
func TestExtractedEditorAPIRoutesStayUnregistered(t *testing.T) {
	s := newTestServer(t)
	for _, route := range []string{"/api/editor/preview", "/api/editor/export"} {
		req := httptest.NewRequest(http.MethodPost, route, nil)
		resp, err := s.app.Test(req)
		if err != nil {
			t.Fatalf("%s: app.Test: %v", route, err)
		}
		if resp.StatusCode != http.StatusNotFound {
			t.Errorf("%s: status %d, want %d", route, resp.StatusCode, http.StatusNotFound)
		}
	}
}
