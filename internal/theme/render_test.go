package theme

import (
	"strings"
	"testing"
)

// fakeTheme builds a theme with N uniform 10x20 frames.
func fakeTheme(name string, n int) *Theme {
	th := &Theme{Name: name, Frames: make([]Frame, n)}
	for i := 0; i < n; i++ {
		th.Frames[i] = Frame{Width: 10, Height: 20, Data: "data:image/gif;base64,QQ"}
	}
	return th
}

func TestRenderFrameAndText(t *testing.T) {
	th := fakeTheme("fake", 3)
	svg, err := Render(th, RenderParams{FrameIndex: 1, Count: 5})
	if err != nil {
		t.Fatalf("Render: %v", err)
	}
	if !strings.HasPrefix(svg, "<?xml") {
		t.Errorf("not xml: %q", svg[:16])
	}
	// viewBox = frame 10 x (20 + textBand 24) = 10 x 44.
	if !strings.Contains(svg, `viewBox="0 0 20 44"`) {
		t.Errorf("viewBox wrong: %s", sub(svg, "viewBox"))
	}
	// count text 5 present and centered.
	if !strings.Contains(svg, `>5<`) {
		t.Errorf("count text missing:\n%s", svg)
	}
	if !strings.Contains(svg, `text-anchor="middle"`) {
		t.Errorf("text not centered")
	}
	// frame image present with data uri.
	if !strings.Contains(svg, `xlink:href="data:image/gif;base64,QQ"`) {
		t.Errorf("frame image missing")
	}
}

func TestRenderNumberOverridesText(t *testing.T) {
	th := fakeTheme("fake", 3)
	svg, err := Render(th, RenderParams{FrameIndex: 0, Count: 5, Number: 42})
	if err != nil {
		t.Fatalf("Render: %v", err)
	}
	if !strings.Contains(svg, `>42<`) {
		t.Errorf("number text 42 missing: %s", sub(svg, "text"))
	}
	if strings.Contains(svg, `>5<`) {
		t.Errorf("count should be overridden by number")
	}
}

func TestRenderFrameIndexOutOfRange(t *testing.T) {
	th := fakeTheme("fake", 2)
	if _, err := Render(th, RenderParams{FrameIndex: 5}); err == nil {
		t.Fatal("expected error for out-of-range frame")
	}
}

func TestRenderNilTheme(t *testing.T) {
	if _, err := Render(nil, RenderParams{}); err == nil {
		t.Fatal("expected error for nil theme")
	}
}

func TestRenderEmptyTheme(t *testing.T) {
	th := &Theme{Name: "empty", Frames: nil}
	if _, err := Render(th, RenderParams{}); err == nil {
		t.Fatal("expected error for empty theme")
	}
}

func TestEscapeXML(t *testing.T) {
	if got := escapeXML(`<a&b>`); got != "&lt;a&amp;b&gt;" {
		t.Errorf("escapeXML: %q", got)
	}
}

func TestThemeSizeAndFrame(t *testing.T) {
	th := fakeTheme("fake", 4)
	if th.Size() != 4 {
		t.Errorf("Size = %d want 4", th.Size())
	}
	if f, ok := th.Frame(2); !ok || f.Width != 10 {
		t.Errorf("Frame(2) = %+v ok=%v", f, ok)
	}
	if _, ok := th.Frame(4); ok {
		t.Error("Frame(4) should be out of range")
	}
}

func sub(s, marker string) string {
	i := strings.Index(s, marker)
	if i < 0 {
		return "(not found)"
	}
	end := i + 50
	if end > len(s) {
		end = len(s)
	}
	return s[i:end]
}

// A multi-digit count must widen the viewBox so the text never overflows.
func TestRenderWideTextWidensViewBox(t *testing.T) {
	th := fakeTheme("fake", 1) // frame 10x20
	svg, err := Render(th, RenderParams{FrameIndex: 0, Count: 123456})
	if err != nil {
		t.Fatalf("Render: %v", err)
	}
	// 6 digits * 10 + 10 padding = 70 > frame width 10.
	if !strings.Contains(svg, `viewBox="0 0 70 44"`) {
		t.Errorf("wide text viewBox wrong: %s", sub(svg, "viewBox"))
	}
	// frame image should be centered: x = (70-10)/2 = 30.
	if !strings.Contains(svg, `x="30" y="0"`) {
		t.Errorf("frame not centered: %s", sub(svg, "image"))
	}
}

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
	// Background uses external URL (Iron Rule 2), not data URI.
	if !strings.Contains(svg, `href="https://cdn.example.com/bg.png"`) {
		t.Errorf("background should reference external URL: %s", svg)
	}
	// Digits use data URI (Iron Rule 2).
	if !strings.Contains(svg, `href="data:image/gif;base64,QQ"`) {
		t.Errorf("digits should use data URI: %s", svg)
	}
	// Two digits (4, 2) -> two digit <image> tags plus the bg image = 3 total.
	if got := strings.Count(svg, "<image"); got != 3 {
		t.Errorf("expected 3 <image> tags (1 bg + 2 digits), got %d", got)
	}
}

func TestRenderWithBgInvalidBg(t *testing.T) {
	th := &Theme{Name: "loli", Frames: []Frame{{Width: 10, Height: 10, Data: "data:image/gif;base64,QQ"}}}
	// Empty URL must error.
	if _, err := RenderWithBg(th, BgParams{}, OverlayParams{}, RenderParams{FrameIndex: 0}); err == nil {
		t.Error("expected error for empty bg URL")
	}
}

func TestRenderWithBgScaleAffectsDigitSize(t *testing.T) {
	th := &Theme{Name: "loli", Frames: []Frame{{Width: 20, Height: 40, Data: "data:image/gif;base64,QQ"}}}
	bp := BgParams{URL: "https://cdn.example.com/bg.png", Width: 100, Height: 100}
	op := OverlayParams{X: 0, Y: 0, FSize: 0, Scale: 0.5}
	rp := RenderParams{Count: 5, Number: -1, FrameIndex: 0}

	svg, _ := RenderWithBg(th, bp, op, rp)
	// Native height 40 * scale 0.5 = 20.
	if !strings.Contains(svg, `height="20"`) {
		t.Errorf("digit height should be 20 (40*0.5): %s", svg)
	}
}
