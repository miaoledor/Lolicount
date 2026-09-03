package server

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"testing/fstest"
)

// newPsbTestServer returns a server whose embedded psb tree is replaced
// by an in-memory fixture.
func newPsbTestServer(t *testing.T, models fstest.MapFS) *Server {
	t.Helper()
	s := newTestServer(t)
	if models == nil {
		s.psbFS = nil
	} else {
		s.psbFS = models
	}
	return s
}

func TestListPsbModels(t *testing.T) {
	s := newPsbTestServer(t, fstest.MapFS{
		"azuki/model.psb":       &fstest.MapFile{Data: []byte("azuki-bytes")},
		"chara/model.psb":       &fstest.MapFile{Data: []byte("chara-bytes")},
		"empty/notmodel.txt":    &fstest.MapFile{Data: []byte("ignored")},
		"README.md":             &fstest.MapFile{Data: []byte("ignored")},
	})

	req := httptest.NewRequest(http.MethodGet, "/api/psb/models", nil)
	resp, err := s.app.Test(req)
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusOK)
	}
	body, _ := io.ReadAll(resp.Body)
	// Sorted, only directories containing model.psb.
	want := `{"models":["azuki","chara"]}`
	if got := string(body); got != want {
		t.Errorf("body: got %s want %s", got, want)
	}
	if cc := resp.Header.Get("Cache-Control"); cc != "public, max-age=60" {
		t.Errorf("Cache-Control: got %q want public, max-age=60", cc)
	}
}

func TestListPsbModelsEmpty(t *testing.T) {
	s := newPsbTestServer(t, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/psb/models", nil)
	resp, err := s.app.Test(req)
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}
	body, _ := io.ReadAll(resp.Body)
	if got, want := string(body), `{"models":[]}`; got != want {
		t.Errorf("body: got %s want %s", got, want)
	}
}

func TestPsbModelServed(t *testing.T) {
	s := newPsbTestServer(t, fstest.MapFS{
		"azuki/model.psb": &fstest.MapFile{Data: []byte("PSB-TEST-BYTES")},
	})

	req := httptest.NewRequest(http.MethodGet, "/psb/azuki", nil)
	resp, err := s.app.Test(req)
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusOK)
	}
	body, _ := io.ReadAll(resp.Body)
	if string(body) != "PSB-TEST-BYTES" {
		t.Errorf("body: got %q want PSB-TEST-BYTES", string(body))
	}
	if ct := resp.Header.Get("Content-Type"); ct != "application/octet-stream" {
		t.Errorf("Content-Type: got %q want application/octet-stream", ct)
	}
	if cc := resp.Header.Get("Cache-Control"); cc != "public, max-age=31536000, immutable" {
		t.Errorf("Cache-Control: got %q want immutable", cc)
	}
	// Open CORS so the dev Nuxt server and the widget can fetch the bytes
	// cross-origin.
	if allow := resp.Header.Get("Access-Control-Allow-Origin"); allow != "*" {
		t.Errorf("Access-Control-Allow-Origin: got %q want *", allow)
	}
}

func TestPsbModelRejectsTraversalAndUnknown(t *testing.T) {
	s := newPsbTestServer(t, fstest.MapFS{
		"azuki/model.psb": &fstest.MapFile{Data: []byte("bytes")},
	})

	for _, path := range []string{"/psb/..%2Ftheme", "/psb/azuki%2F..", "/psb/no-such", "/psb/不良名"} {
		req := httptest.NewRequest(http.MethodGet, path, nil)
		resp, err := s.app.Test(req)
		if err != nil {
			t.Fatalf("app.Test %s: %v", path, err)
		}
		if resp.StatusCode == http.StatusOK {
			t.Errorf("%s: got 200, want non-OK", path)
		}
	}
}
