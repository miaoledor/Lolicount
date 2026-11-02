package server

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

// TestRetiredMutationRoutesStayUnavailable protects the repository
// boundary after editor and theme submission moved out of this service.
func TestRetiredMutationRoutesStayUnavailable(t *testing.T) {
	s := newTestServer(t)
	tests := []struct {
		method string
		path   string
		want   int
	}{
		{method: http.MethodPost, path: "/api/editor/preview", want: http.StatusMethodNotAllowed},
		{method: http.MethodPost, path: "/api/editor/export", want: http.StatusMethodNotAllowed},
		{method: http.MethodPost, path: "/api/themes", want: http.StatusMethodNotAllowed},
		{method: http.MethodPost, path: "/api/themes/upload", want: http.StatusMethodNotAllowed},
		{method: http.MethodGet, path: "/api/admin/ping", want: http.StatusNotFound},
	}
	for _, tt := range tests {
		req := httptest.NewRequest(tt.method, tt.path, nil)
		resp, err := s.app.Test(req)
		if err != nil {
			t.Fatalf("%s %s: app.Test: %v", tt.method, tt.path, err)
		}
		if resp.StatusCode != tt.want {
			t.Errorf("%s %s: status %d, want %d", tt.method, tt.path, resp.StatusCode, tt.want)
		}
	}
}
