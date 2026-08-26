package server

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestCounterPositionPixel(t *testing.T) {
	s := newCounterServer(t) // stub 10x20 -> 200x400
	req := httptest.NewRequest(http.MethodGet, "/@demo?theme=lian&x=50&y=100", nil)
	resp, _ := s.app.Test(req)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status: %d", resp.StatusCode)
	}
	body := readBody(t, resp)
	if !strings.Contains(body, `text-anchor="start"`) {
		t.Errorf("pixel x/y should use start anchor: %s", sub(body, "text-anchor"))
	}
	if !strings.Contains(body, `x="50"`) {
		t.Errorf("pixel x=50 missing: %s", sub(body, "x="))
	}
}

// M6: ratio rx/ry positions the text by fraction of image dims.
// The stub image is 10x20 -> scaled to 200x400 (longest edge = 400).
// rx=0.5 => x = 200*0.5 = 100; ry=0.25 => y = 400*0.25 + fontSize(16) = 116.
// This verifies ratio uses the IMAGE dims, not the full canvas (which
// would be 400 + textH, shifting y).
func TestCounterPositionRatio(t *testing.T) {
	s := newCounterServer(t) // stub 10x20 -> 200x400
	req := httptest.NewRequest(http.MethodGet, "/@demo?theme=lian&rx=0.5&ry=0.25", nil)
	resp, _ := s.app.Test(req)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status: %d", resp.StatusCode)
	}
	body := readBody(t, resp)
	if !strings.Contains(body, `x="100"`) {
		t.Errorf("ratio rx=0.5 should give x=100: %s", sub(body, "x="))
	}
	if !strings.Contains(body, `y="116"`) {
		t.Errorf("ratio ry=0.25 should give y=116 (400*0.25+16), got: %s", sub(body, "y="))
	}
}

// TestCounterPositionRatioY verifies ry uses image height, not canvas
// height. The canvas height is imageH + textH (400 + 20 = 420). If ry
// used canvasH, y would be 420*0.5+16 = 226 instead of 400*0.5+16 = 216.
func TestCounterPositionRatioY(t *testing.T) {
	s := newCounterServer(t) // 10x20 -> 200x400, canvasH = 420
	req := httptest.NewRequest(http.MethodGet, "/@demo?theme=lian&rx=0&ry=0.5", nil)
	resp, _ := s.app.Test(req)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status: %d", resp.StatusCode)
	}
	body := readBody(t, resp)
	// 400*0.5 + 16 = 216 (image-based); 420*0.5 + 16 = 226 (canvas-based, wrong)
	if strings.Contains(body, `y="226"`) {
		t.Errorf("ry used canvas height (226) instead of image height; want y=216: %s", sub(body, "y="))
	}
	if !strings.Contains(body, `y="216"`) {
		t.Errorf("ry=0.5 should give y=216 (400*0.5+16), got: %s", sub(body, "y="))
	}
}

// TestCounterPositionPixelY verifies pixel y places the text baseline at
// y + fontSize (the top edge is y, baseline is y + fontSize).
func TestCounterPositionPixelY(t *testing.T) {
	s := newCounterServer(t)
	req := httptest.NewRequest(http.MethodGet, "/@demo?theme=lian&x=10&y=50", nil)
	resp, _ := s.app.Test(req)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status: %d", resp.StatusCode)
	}
	body := readBody(t, resp)
	// y=50, fontSize=16 => baseline y = 50 + 16 = 66
	if !strings.Contains(body, `y="66"`) {
		t.Errorf("pixel y=50 should give baseline y=66 (50+16), got: %s", sub(body, "y="))
	}
}

// M6: invalid rx (>1) should 400.
func TestCounterPositionInvalidRatio400(t *testing.T) {
	s := newCounterServer(t)
	req := httptest.NewRequest(http.MethodGet, "/@demo?theme=lian&rx=2", nil)
	resp, _ := s.app.Test(req)
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("rx=2 should 400, got %d", resp.StatusCode)
	}
}

// GET /api/themes returns the registered theme names as JSON.
