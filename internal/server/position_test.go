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
