package theme

import (
	"strings"
	"testing"
)

// Edge: negative X positions digits off-canvas (valid, should still render).
func TestRenderWithBgNegativeXY(t *testing.T) {
	th := &Theme{Name: "loli", Frames: []Frame{{Width: 10, Height: 10, Data: "data:image/gif;base64,QQ"}}}
	bp := BgParams{URL: "https://x/y.png", Width: 100, Height: 100}
	op := OverlayParams{X: -20, Y: -10, Scale: 1}
	svg, err := RenderWithBg(th, bp, op, RenderParams{Count: 5, Number: -1, FrameIndex: 0})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(svg, `x="-20"`) {
		t.Errorf("negative X should appear: %s", svg)
	}
}

// Edge: count=0 renders one digit "0".
func TestRenderWithBgCountZero(t *testing.T) {
	th := &Theme{Name: "loli", Frames: []Frame{{Width: 10, Height: 10, Data: "data:image/gif;base64,QQ"}}}
	bp := BgParams{URL: "https://x/y.png", Width: 100, Height: 100}
	svg, _ := RenderWithBg(th, bp, OverlayParams{Scale: 1}, RenderParams{Count: 0, Number: -1, FrameIndex: 0})
	if got := strings.Count(svg, "<image"); got != 2 {
		t.Errorf("count=0: expected 2 images (1 bg + 1 digit), got %d", got)
	}
}

// Edge: multi-digit count renders N digit images.
func TestRenderWithBgMultiDigit(t *testing.T) {
	th := &Theme{Name: "loli", Frames: []Frame{{Width: 10, Height: 10, Data: "data:image/gif;base64,QQ"}}}
	bp := BgParams{URL: "https://x/y.png", Width: 100, Height: 100}
	svg, _ := RenderWithBg(th, bp, OverlayParams{Scale: 1}, RenderParams{Count: 12345, Number: -1, FrameIndex: 0})
	if got := strings.Count(svg, "<image"); got != 6 {
		t.Errorf("count=12345: expected 6 images (1 bg + 5 digits), got %d", got)
	}
}

// Edge: tiny scale producing digitW<=0 falls back without panic.
func TestRenderWithBgTinyScaleNoPanic(t *testing.T) {
	th := &Theme{Name: "loli", Frames: []Frame{{Width: 1, Height: 100, Data: "data:image/gif;base64,QQ"}}}
	bp := BgParams{URL: "https://x/y.png", Width: 100, Height: 100}
	svg, err := RenderWithBg(th, bp, OverlayParams{Scale: 0.1}, RenderParams{Count: 1, Number: -1, FrameIndex: 0})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(svg, "<image") {
		t.Error("should still render an image")
	}
}
