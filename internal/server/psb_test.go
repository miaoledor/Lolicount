package server

import (
	"bytes"
	"compress/gzip"
	"io"
	"io/fs"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"
	"testing/fstest"

	"github.com/rs/zerolog"

	"github.com/miaoledor/lolicount/internal/config"
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

type trackingFS struct {
	fs.FS
	active atomic.Int32
}

func (f *trackingFS) Open(name string) (fs.File, error) {
	file, err := f.FS.Open(name)
	if err != nil {
		return nil, err
	}
	f.active.Add(1)
	return &trackingFile{File: file, active: &f.active}, nil
}

type trackingFile struct {
	fs.File
	active *atomic.Int32
	closed atomic.Bool
}

func (f *trackingFile) Close() error {
	if !f.closed.CompareAndSwap(false, true) {
		return nil
	}
	f.active.Add(-1)
	return f.File.Close()
}

// TestPsbDirLoadedFromDisk pins the on-disk loading contract: New wires
// psbFS from cfg.PSB_DIR, so model files are streamed from disk only when
// a request asks for them — the binary does not embed them.
func TestPsbDirLoadedFromDisk(t *testing.T) {
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, "disk-model"), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "disk-model", "model.psb"), []byte("from-disk"), 0o644); err != nil {
		t.Fatalf("write model: %v", err)
	}

	cfg := &config.Config{
		Host: "127.0.0.1", Port: 0, DBInterval: 10, PSBDir: dir,
		RateLimitIPPerSec: 10000, RateLimitIPPerMin: 100000, RateLimitNamePerSec: 10000,
	}
	s := New(cfg, zerolog.Nop(), nil, nil, nil, nil)
	t.Cleanup(func() {
		s.ipLimiter.Stop()
		s.nameLimiter.Stop()
	})
	if err := os.WriteFile(filepath.Join(dir, "disk-model", "model.psb"), []byte("read-on-request"), 0o644); err != nil {
		t.Fatalf("replace model after startup: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/api/psb/models", nil)
	resp, err := s.app.Test(req)
	if err != nil {
		t.Fatalf("app.Test models: %v", err)
	}
	body, _ := io.ReadAll(resp.Body)
	if !strings.Contains(string(body), `"disk-model"`) {
		t.Errorf("disk model missing from list: %s", body)
	}

	req = httptest.NewRequest(http.MethodGet, "/psb/disk-model", nil)
	resp, err = s.app.Test(req)
	if err != nil {
		t.Fatalf("app.Test fetch: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("fetch status: got %d want 200", resp.StatusCode)
	}
	body, _ = io.ReadAll(resp.Body)
	if string(body) != "read-on-request" {
		t.Errorf("streamed body: got %q want read-on-request", string(body))
	}
}

// The response owns the opened model stream. fasthttp must close it after
// the body is written so no file descriptor or server-side model resource
// remains live between requests.
func TestPsbModelStreamClosedAfterResponse(t *testing.T) {
	models := &trackingFS{FS: fstest.MapFS{
		"azuki/model.psb": &fstest.MapFile{Data: []byte("streamed")},
	}}
	s := newTestServer(t)
	s.psbFS = models

	req := httptest.NewRequest(http.MethodGet, "/psb/azuki", nil)
	resp, err := s.app.Test(req)
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}
	if _, err := io.Copy(io.Discard, resp.Body); err != nil {
		t.Fatalf("read body: %v", err)
	}
	if err := resp.Body.Close(); err != nil {
		t.Fatalf("close body: %v", err)
	}
	if active := models.active.Load(); active != 0 {
		t.Fatalf("open model streams after response: got %d want 0", active)
	}
}

// TestPsbDirMissingDisablesEndpoints verifies that a missing PSB_DIR
// degrades to an empty model list instead of failing startup.
func TestPsbDirMissingDisablesEndpoints(t *testing.T) {
	cfg := &config.Config{
		Host: "127.0.0.1", Port: 0, DBInterval: 10,
		PSBDir:            filepath.Join(t.TempDir(), "not-there"),
		RateLimitIPPerSec: 10000, RateLimitIPPerMin: 100000, RateLimitNamePerSec: 10000,
	}
	s := New(cfg, zerolog.Nop(), nil, nil, nil, nil)
	t.Cleanup(func() {
		s.ipLimiter.Stop()
		s.nameLimiter.Stop()
	})

	req := httptest.NewRequest(http.MethodGet, "/api/psb/models", nil)
	resp, err := s.app.Test(req)
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}
	body, _ := io.ReadAll(resp.Body)
	if got, want := string(body), `{"models":[]}`; got != want {
		t.Errorf("missing dir models: got %s want %s", got, want)
	}
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
	if cc := resp.Header.Get("Cache-Control"); cc != "private, max-age=31536000, immutable" {
		t.Errorf("Cache-Control: got %q want client-private immutable", cc)
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
	if cc := resp.Header.Get("Cache-Control"); cc != "private, max-age=31536000, immutable" {
		t.Errorf("Cache-Control: got %q want client-private immutable", cc)
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
