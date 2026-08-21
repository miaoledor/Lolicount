package server

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/miaoledor/lolicount/internal/drawer/characterthemedrawer"
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

// M9: GET /api/themes returns each theme's kind so the front-end can
// render type-specific controls.
func TestAPIThemesListWithKind(t *testing.T) {
	s := newCounterServer(t)
	// stub already has "lian" (KindFrame). Add a character theme.
	ch := &characterthemedrawer.Character{
		Layers: make([]characterthemedrawer.CharacterLayer, 80),
		Parts:  make(map[int]characterthemedrawer.CharacterPart, 70),
	}
	if stub, ok := s.themes.(*stubRegistry); ok {
		if stub.characters == nil {
			stub.characters = map[string]*characterthemedrawer.Character{}
		}
		stub.characters["lian-ren"] = ch
	}

	req := httptest.NewRequest(http.MethodGet, "/api/themes", nil)
	resp, err := s.app.Test(req)
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status: %d", resp.StatusCode)
	}
	body := readBody(t, resp)
	if !strings.Contains(body, `"kind":"frame","name":"lian"`) {
		t.Errorf("frame theme kind missing/wrong: %s", body)
	}
	if !strings.Contains(body, `"kind":"character","name":"lian-ren"`) {
		t.Errorf("character theme kind missing/wrong: %s", body)
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
