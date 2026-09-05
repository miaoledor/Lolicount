package server

import (
	"bytes"
	"compress/gzip"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
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
		"azuki/model.psb":    &fstest.MapFile{Data: []byte("azuki-bytes")},
		"chara/model.psb":    &fstest.MapFile{Data: []byte("chara-bytes")},
		"empty/notmodel.txt": &fstest.MapFile{Data: []byte("ignored")},
		"README.md":          &fstest.MapFile{Data: []byte("ignored")},
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

// Gzip-stored models are served as pre-compressed bytes with
// Content-Encoding: gzip — the browser decompresses transparently, so the
// body stays compressed on the wire while the driver receives pure PSB.
func TestPsbModelGzipServed(t *testing.T) {
	gz := gzipBytes(t, "PSB-GZ-TEST-BYTES")
	s := newPsbTestServer(t, fstest.MapFS{
		"chocola/model.psb.gz": &fstest.MapFile{Data: gz},
	})

	req := httptest.NewRequest(http.MethodGet, "/psb/chocola", nil)
	resp, err := s.app.Test(req)
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusOK)
	}
	if enc := resp.Header.Get("Content-Encoding"); enc != "gzip" {
		t.Errorf("Content-Encoding: got %q want gzip", enc)
	}
	body, _ := io.ReadAll(resp.Body)
	if !bytes.Equal(body, gz) {
		t.Errorf("body should be the raw gzipped bytes on the wire")
	}

	// The model must also be listed.
	req = httptest.NewRequest(http.MethodGet, "/api/psb/models", nil)
	resp, err = s.app.Test(req)
	if err != nil {
		t.Fatalf("app.Test models: %v", err)
	}
	got, _ := io.ReadAll(resp.Body)
	if !strings.Contains(string(got), `"chocola"`) {
		t.Errorf("gz-stored model missing from list: %s", got)
	}
}

// Plain (uncompressed) model.psb files still work and carry no
// Content-Encoding.
func TestPsbModelPlainHasNoEncoding(t *testing.T) {
	s := newPsbTestServer(t, fstest.MapFS{
		"azuki/model.psb": &fstest.MapFile{Data: []byte("plain")},
	})

	req := httptest.NewRequest(http.MethodGet, "/psb/azuki", nil)
	resp, err := s.app.Test(req)
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}
	if enc := resp.Header.Get("Content-Encoding"); enc != "" {
		t.Errorf("Content-Encoding: got %q want empty", enc)
	}
}

// The gz fixture is a real gzip stream so the response is well-formed.
func gzipBytes(t *testing.T, raw string) []byte {
	t.Helper()
	var buf bytes.Buffer
	w := gzip.NewWriter(&buf)
	if _, err := w.Write([]byte(raw)); err != nil {
		t.Fatal(err)
	}
	if err := w.Close(); err != nil {
		t.Fatal(err)
	}
	return buf.Bytes()
}

// GET /api/psb/:model/download delivers the stored file as an attachment
// — the gzip stays a real .psb.gz on disk (no Content-Encoding), with a
// proper filename.
func TestPsbModelDownload(t *testing.T) {
	gz := gzipBytes(t, "DOWNLOAD-TEST")
	s := newPsbTestServer(t, fstest.MapFS{
		"maple-dress-a/model.psb.gz": &fstest.MapFile{Data: gz},
	})

	req := httptest.NewRequest(http.MethodGet, "/api/psb/maple-dress-a/download", nil)
	resp, err := s.app.Test(req)
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status: got %d want 200", resp.StatusCode)
	}
	if cd := resp.Header.Get("Content-Disposition"); cd != `attachment; filename="maple-dress-a-model.psb.gz"` {
		t.Errorf("Content-Disposition: got %q", cd)
	}
	if ct := resp.Header.Get("Content-Type"); ct != "application/gzip" {
		t.Errorf("Content-Type: got %q want application/gzip", ct)
	}
	if enc := resp.Header.Get("Content-Encoding"); enc != "" {
		t.Errorf("Content-Encoding must be empty so the file stays gzipped, got %q", enc)
	}
	body, _ := io.ReadAll(resp.Body)
	if !bytes.Equal(body, gz) {
		t.Errorf("body should be the raw gzip file bytes")
	}
}

// Unknown model names 404 on the download endpoint.
func TestPsbModelDownloadNotFound(t *testing.T) {
	s := newPsbTestServer(t, fstest.MapFS{
		"azuki/model.psb": &fstest.MapFile{Data: []byte("x")},
	})
	req := httptest.NewRequest(http.MethodGet, "/api/psb/nope/download", nil)
	resp, err := s.app.Test(req)
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("status: got %d want 404", resp.StatusCode)
	}
}
