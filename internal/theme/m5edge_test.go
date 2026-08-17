package theme

import (
	"strings"
	"testing"
)

// M5.5: in bg overlay mode the count is rendered as <text>, NOT as an
// <image>. The only <image> is the background.
func TestRenderWithBgOverlay(t *testing.T) {
	th := &Theme{Name: "loli", Frames: []Frame{{Width: 20, Height: 30, Data: "data:image/gif;base64,QQ"}}}
	bp := BgParams{URL: "https://cdn.example.com/bg.png", Width: 400, Height: 300}
	op := OverlayParams{X: 20, Y: 180, FSize: 40, Scale: 1, Align: "top"}
	rp := RenderParams{Count: 42, Number: -1, FrameIndex: 0}

	svg, err := RenderWithBg(th, bp, op, rp)
	if err != nil {
		t.Fatalf("RenderWithBg: %v", err)
	}
	// viewBox fixed to background dimensions.
	if !strings.Contains(svg, `viewBox="0 0 400 300"`) {
		t.Errorf("viewBox should be background dims: %s", svg)
	}
	// Background uses external URL (Iron Rule 2).
	if !strings.Contains(svg, `href="https://cdn.example.com/bg.png"`) {
		t.Errorf("background should reference external URL: %s", svg)
	}
	// Count is <text>, not a data-URI <image> (M5.5: count is a font).
	if !strings.Contains(svg, `<text`) {
		t.Errorf("count should be rendered as <text>: %s", svg)
	}
	if !strings.Contains(svg, `>42<`) {
		t.Errorf("count text 42 missing: %s", sub(svg, "text"))
	}
	if strings.Contains(svg, `data:image/gif;base64,QQ`) {
		t.Errorf("theme frame data URI must NOT appear in bg mode (count is text now): %s", svg)
	}
	// fsize controls font-size.
	if !strings.Contains(svg, `font-size="40"`) {
		t.Errorf("fsize=40 should set font-size=40: %s", sub(svg, "font-size"))
	}
	// Only one <image> (the background); no digit images.
	if got := strings.Count(svg, "<image"); got != 1 {
		t.Errorf("expected 1 <image> (background only), got %d", got)
	}
}

func TestRenderWithBgInvalidBg(t *testing.T) {
	th := &Theme{Name: "loli", Frames: []Frame{{Width: 10, Height: 10, Data: "data:image/gif;base64,QQ"}}}
	if _, err := RenderWithBg(th, BgParams{}, OverlayParams{}, RenderParams{FrameIndex: 0}); err == nil {
		t.Error("expected error for empty bg URL")
	}
}

// fsize controls the text font-size in bg mode.
func TestRenderWithBgFSizeControlsText(t *testing.T) {
	th := &Theme{Name: "loli", Frames: []Frame{{Width: 20, Height: 40, Data: "data:image/gif;base64,QQ"}}}
	bp := BgParams{URL: "https://cdn.example.com/bg.png", Width: 100, Height: 100}
	op := OverlayParams{X: 0, Y: 0, FSize: 50, Scale: 1}
	svg, _ := RenderWithBg(th, bp, op, RenderParams{Count: 5, Number: -1, FrameIndex: 0})
	if !strings.Contains(svg, `font-size="50"`) {
		t.Errorf("fsize=50 should set font-size=50: %s", sub(svg, "font-size"))
	}
}

// Edge: negative X positions text off-canvas (valid, should still render).
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

// Edge: count=0 renders one "0" text.
func TestRenderWithBgCountZero(t *testing.T) {
	th := &Theme{Name: "loli", Frames: []Frame{{Width: 10, Height: 10, Data: "data:image/gif;base64,QQ"}}}
	bp := BgParams{URL: "https://x/y.png", Width: 100, Height: 100}
	svg, _ := RenderWithBg(th, bp, OverlayParams{Scale: 1}, RenderParams{Count: 0, Number: -1, FrameIndex: 0})
	if got := strings.Count(svg, "<image"); got != 1 {
		t.Errorf("count=0: expected 1 image (bg only), got %d", got)
	}
	if !strings.Contains(svg, `>0<`) {
		t.Errorf("should render text '0': %s", sub(svg, "text"))
	}
}

// Edge: multi-digit count renders as one <text> with all digits.
func TestRenderWithBgMultiDigit(t *testing.T) {
	th := &Theme{Name: "loli", Frames: []Frame{{Width: 10, Height: 10, Data: "data:image/gif;base64,QQ"}}}
	bp := BgParams{URL: "https://x/y.png", Width: 100, Height: 100}
	svg, _ := RenderWithBg(th, bp, OverlayParams{Scale: 1}, RenderParams{Count: 12345, Number: -1, FrameIndex: 0})
	// 1 bg image + 1 text element (not 5 digit images).
	if got := strings.Count(svg, "<image"); got != 1 {
		t.Errorf("expected 1 image (bg only), got %d", got)
	}
	if !strings.Contains(svg, `>12345<`) {
		t.Errorf("should render text '12345': %s", sub(svg, "text"))
	}
}
