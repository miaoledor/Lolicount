package server

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"testing/fstest"
)

// GET /api/themes returns the registered theme names as JSON.
func TestAPIThemesList(t *testing.T) {
	s := newCounterServer(t) // stub has "lian"
	req := httptest.NewRequest(http.MethodGet, "/api/themes", nil)
	resp, err := s.app.Test(req)
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status: %d", resp.StatusCode)
	}
	body := readBody(t, resp)
	if !strings.Contains(body, `"lian"`) {
		t.Errorf("themes list missing lian: %s", body)
	}
}

// GET /api/fthemes returns the registered f-theme names.
func TestAPIFThemesList(t *testing.T) {
	s := newFThemeServer(t) // stub has "pink"
	req := httptest.NewRequest(http.MethodGet, "/api/fthemes", nil)
	resp, _ := s.app.Test(req)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status: %d", resp.StatusCode)
	}
	body := readBody(t, resp)
	if !strings.Contains(body, `"pink"`) {
		t.Errorf("fthemes list missing pink: %s", body)
	}
}

// GET /api/themes: static imgcore themes carry no "kind"; only the animated
// model entries (psb/spine/live2d/unity) do. A stub server with no animated
// models must therefore emit no "kind" at all.
func TestAPIThemesListNoKindForStatic(t *testing.T) {
	s := newCounterServer(t) // stub has "lian", no animated models
	// Ensure no animated model FS is present so the response is static-only.
	s.psbFS = nil
	s.spineFS = nil
	s.live2dFS = nil
	s.unityFS = nil
	req := httptest.NewRequest(http.MethodGet, "/api/themes", nil)
	resp, err := s.app.Test(req)
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status: %d", resp.StatusCode)
	}
	body := readBody(t, resp)
	if !strings.Contains(body, `"lian"`) {
		t.Errorf("themes list missing lian: %s", body)
	}
	if strings.Contains(body, "kind") {
		t.Errorf("static themes list should not contain kind: %s", body)
	}
}

// GET /api/config returns the configured baseUrl.
func TestAPIConfigBaseUrl(t *testing.T) {
	s := newCounterServer(t)
	s.cfg.BaseURL = "https://umi7.top"

	req := httptest.NewRequest(http.MethodGet, "/api/config", nil)
	resp, err := s.app.Test(req)
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status: %d", resp.StatusCode)
	}
	body := readBody(t, resp)
	if !strings.Contains(body, `"baseUrl":"https://umi7.top"`) {
		t.Errorf("api/config baseUrl missing: %s", body)
	}
}

func TestAPIConfigBaseUrlEmpty(t *testing.T) {
	s := newCounterServer(t)
	s.cfg.BaseURL = ""

	req := httptest.NewRequest(http.MethodGet, "/api/config", nil)
	resp, _ := s.app.Test(req)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status: %d", resp.StatusCode)
	}
	body := readBody(t, resp)
	if !strings.Contains(body, `"baseUrl":""`) {
		t.Errorf("api/config empty baseUrl missing: %s", body)
	}
}

// GET /api/themes appends emote models with an animated flag so the theme
// picker can mark them and switch to the widget embed flow.
func TestAPIThemesListAnimatedModels(t *testing.T) {
	s := newCounterServer(t) // stub has "lian"
	s.psbFS = fstest.MapFS{
		"azuki/model.psb": &fstest.MapFile{Data: []byte("bytes")},
	}

	req := httptest.NewRequest(http.MethodGet, "/api/themes", nil)
	resp, err := s.app.Test(req)
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}
	body := readBody(t, resp)
	// Animated entries carry the "kind" field (JSON map keys sort
	// alphabetically: animated, kind, name).
	if !strings.Contains(body, `{"animated":true,"kind":"psb","name":"azuki"}`) {
		t.Errorf("themes list missing animated azuki entry: %s", body)
	}
	if !strings.Contains(body, `"lian"`) {
		t.Errorf("imgcore themes should still be listed: %s", body)
	}
}

// GET /api/themes appends spine, live2d and unity models with their kinds, so
// the front-end can route each animated model to the matching player page.
func TestAPIThemesListSpineAndLive2DModels(t *testing.T) {
	s := newCounterServer(t)
	s.spineFS = fstest.MapFS{
		"kalts/dyn.skel":  &fstest.MapFile{Data: []byte("s")},
		"kalts/dyn.atlas": &fstest.MapFile{Data: []byte("a")},
	}
	s.live2dFS = fstest.MapFS{
		"archchan/model3.json":   &fstest.MapFile{Data: []byte("{}")},
		"archchan/archchan.moc3": &fstest.MapFile{Data: []byte("m")},
	}
	s.unityFS = fstest.MapFS{
		"maimai-navichara-01/model.json": &fstest.MapFile{Data: []byte("{}")},
	}

	req := httptest.NewRequest(http.MethodGet, "/api/themes", nil)
	resp, err := s.app.Test(req)
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}
	body := readBody(t, resp)
	if !strings.Contains(body, `{"animated":true,"kind":"spine","name":"kalts"}`) {
		t.Errorf("themes list missing spine entry: %s", body)
	}
	if !strings.Contains(body, `{"animated":true,"kind":"live2d","name":"archchan"}`) {
		t.Errorf("themes list missing live2d entry: %s", body)
	}
	if !strings.Contains(body, `{"animated":true,"kind":"unity","name":"maimai-navichara-01"}`) {
		t.Errorf("themes list missing unity entry: %s", body)
	}
}
