package fdrawer

import (
	"strings"
	"testing"
)

func TestDrawUnshowFontReturnsEmpty(t *testing.T) {
	layer := Draw(Params{Text: "5", UnshowFont: true}, 200, 400)
	if layer.Fragment != "" || layer.Height != 0 {
		t.Errorf("unshowf should return empty layer: %+v", layer)
	}
}

func TestDrawDefaultBelowImageCentered(t *testing.T) {
	layer := Draw(Params{Text: "5", FontSize: 16}, 200, 400)
	if !strings.Contains(layer.Fragment, `text-anchor="middle"`) {
		t.Errorf("default should be centered: %s", layer.Fragment)
	}
	// textY = bgH + fontSize = 400 + 16 = 416
	if !strings.Contains(layer.Fragment, `y="416"`) {
		t.Errorf("default text y should be below image: %s", layer.Fragment)
	}
}

func TestDrawPixelPosition(t *testing.T) {
	layer := Draw(Params{Text: "5", FontSize: 16, Position: TextPos{X: 50, Y: 100}}, 200, 400)
	if !strings.Contains(layer.Fragment, `text-anchor="start"`) {
		t.Errorf("pixel mode should use start anchor: %s", layer.Fragment)
	}
	if !strings.Contains(layer.Fragment, `x="50"`) {
		t.Errorf("pixel x should be 50: %s", layer.Fragment)
	}
	if !strings.Contains(layer.Fragment, `y="116"`) {
		t.Errorf("pixel y should be Y+fontSize=116: %s", layer.Fragment)
	}
}

func TestDrawRatioPosition(t *testing.T) {
	layer := Draw(Params{Text: "5", FontSize: 16, Position: TextPos{RX: 0.5, RY: 0.25}}, 200, 400)
	if !strings.Contains(layer.Fragment, `text-anchor="start"`) {
		t.Errorf("ratio mode should use start anchor: %s", layer.Fragment)
	}
	// textX = bgW(200) * 0.5 = 100; textY = bgH(400) * 0.25 + 16 = 116
	if !strings.Contains(layer.Fragment, `x="100"`) {
		t.Errorf("ratio x should be bgW*rx=100: %s", layer.Fragment)
	}
	if !strings.Contains(layer.Fragment, `y="116"`) {
		t.Errorf("ratio y should be bgH*ry+fs=116: %s", layer.Fragment)
	}
}

func TestDrawPixelOverRatio(t *testing.T) {
	layer := Draw(Params{Text: "5", FontSize: 16, Position: TextPos{X: 10, Y: 20, RX: 0.9, RY: 0.9}}, 200, 400)
	if !strings.Contains(layer.Fragment, `x="10"`) {
		t.Errorf("pixel should override ratio: %s", layer.Fragment)
	}
}

func TestDrawFontStyleApplied(t *testing.T) {
	layer := Draw(Params{Text: "5", FontSize: 16, FontStyle: FontStyle{Family: "serif", Color: "#e91e63", Weight: "bold"}}, 200, 400)
	if !strings.Contains(layer.Fragment, `font-family="serif"`) {
		t.Errorf("FontStyle.Family not applied: %s", layer.Fragment)
	}
	if !strings.Contains(layer.Fragment, `fill="#e91e63"`) {
		t.Errorf("FontStyle.Color not applied: %s", layer.Fragment)
	}
	if !strings.Contains(layer.Fragment, `font-weight="bold"`) {
		t.Errorf("FontStyle.Weight not applied: %s", layer.Fragment)
	}
}

func TestDrawFontStyleDefaults(t *testing.T) {
	layer := Draw(Params{Text: "5", FontSize: 16}, 200, 400)
	if !strings.Contains(layer.Fragment, `font-family="monospace"`) {
		t.Errorf("default family not applied: %s", layer.Fragment)
	}
	if !strings.Contains(layer.Fragment, `fill="#333"`) {
		t.Errorf("default color not applied: %s", layer.Fragment)
	}
	if strings.Contains(layer.Fragment, "font-weight") {
		t.Errorf("zero weight should omit font-weight: %s", layer.Fragment)
	}
}

func TestDrawDefaultFontSize(t *testing.T) {
	layer := Draw(Params{Text: "1"}, 200, 400)
	if !strings.Contains(layer.Fragment, `font-size="16"`) {
		t.Errorf("default font-size should be 16: %s", layer.Fragment)
	}
}

func TestDrawCustomFontSize(t *testing.T) {
	layer := Draw(Params{Text: "1", FontSize: 40}, 200, 400)
	if !strings.Contains(layer.Fragment, `font-size="40"`) {
		t.Errorf("fsize=40 should set font-size=40: %s", layer.Fragment)
	}
}

func TestDrawTextContent(t *testing.T) {
	layer := Draw(Params{Text: "0123456789", FontSize: 16}, 200, 400)
	if !strings.Contains(layer.Fragment, `>0123456789<`) {
		t.Errorf("text content missing: %s", layer.Fragment)
	}
}
