package fdrawer

import (
	"testing"
)

func TestMeasureUnshowFont(t *testing.T) {
	w, h := Measure(Params{Text: "12345", UnshowFont: true})
	if w != 0 || h != 0 {
		t.Errorf("UnshowFont should measure 0x0, got %dx%d", w, h)
	}
}

func TestMeasureDefaultFontSize(t *testing.T) {
	w, h := Measure(Params{Text: "1"})
	// fontSize=16, charW=16*0.6=9 (floor 9), textW=1*16*0.6=9, +charW=18
	// height = 16 + 4 = 20
	if w != 18 {
		t.Errorf("width: got %d, want 18", w)
	}
	if h != 20 {
		t.Errorf("height: got %d, want 20", h)
	}
}

func TestMeasureCustomFontSize(t *testing.T) {
	w, h := Measure(Params{Text: "12", FontSize: 100})
	// fontSize=100, charW=100*0.6=60, textW=2*100*0.6=120, +charW=180
	// height = 100 + 4 = 104
	if w != 180 {
		t.Errorf("width: got %d, want 180", w)
	}
	if h != 104 {
		t.Errorf("height: got %d, want 104", h)
	}
}

func TestMeasureEmptyText(t *testing.T) {
	w, h := Measure(Params{Text: ""})
	// textWidth("")=0, +charW=9 (fontSize 16), height=20
	if w != 9 {
		t.Errorf("empty text width: got %d, want 9 (charW only)", w)
	}
	if h != 20 {
		t.Errorf("empty text height: got %d, want 20", h)
	}
}

func TestMeasureMatchesDrawWidth(t *testing.T) {
	// Measure and Draw must agree on the Layer width so the renderer's
	// canvas-width calculation is consistent.
	params := Params{Text: "12345", FontSize: 32}
	mw, _ := Measure(params)
	layer := Draw(params, 400, 400, 400)
	if mw != layer.Width {
		t.Errorf("Measure width %d != Draw width %d", mw, layer.Width)
	}
}

func TestMeasureMatchesDrawHeight(t *testing.T) {
	params := Params{Text: "99", FontSize: 50}
	_, mh := Measure(params)
	layer := Draw(params, 100, 200, 100)
	if mh != layer.Height {
		t.Errorf("Measure height %d != Draw height %d", mh, layer.Height)
	}
}
