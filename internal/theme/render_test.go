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
	// M5.5: viewBox = frame dimensions (10 x 20); text overlays the frame.
	if !strings.Contains(svg, `viewBox="0 0 18 20"`) {
		t.Errorf("viewBox wrong: %s", sub(svg, "viewBox"))
	}
	// count text 5 present.
	if !strings.Contains(svg, `>5<`) {
		t.Errorf("count text missing:\n%s", svg)
	}
	if !strings.Contains(svg, `text-anchor="middle"`) {
		t.Errorf("text not centered")
	}
	// Layer order: image before text (background below text).
	if strings.Index(svg, "<image") > strings.Index(svg, "<text") {
		t.Errorf("image must precede text")
	}
}

// A multi-digit count must widen the viewBox so the text never overflows.
func TestRenderWideTextWidensViewBox(t *testing.T) {
	th := fakeTheme("fake", 1) // frame 10x20
	svg, err := Render(th, RenderParams{FrameIndex: 0, Count: 123456})
	if err != nil {
		t.Fatalf("Render: %v", err)
	}
	// 6 digits at font-size 16: width = 6*16*0.6 = 57.6 + charW(9) = 66.
	// canvasWidth must be >= 66 and > frame width 10.
	if !strings.Contains(svg, `viewBox="0 0 66 20"`) {
		t.Errorf("wide text viewBox wrong: %s", sub(svg, "viewBox"))
	}
}

// fsize controls the counter text font-size (M5.5: count is a font).
func TestRenderFontSizeFromParam(t *testing.T) {
	th := fakeTheme("fake", 1)
	svg, _ := Render(th, RenderParams{FrameIndex: 0, Count: 1, FontSize: 40})
	if !strings.Contains(svg, `font-size="40"`) {
		t.Errorf("fsize=40 should set font-size=40: %s", sub(svg, "font-size"))
	}
}

// scale multiplies the font size.
func TestRenderScaleMultipliesFontSize(t *testing.T) {
	th := fakeTheme("fake", 1)
	svg, _ := Render(th, RenderParams{FrameIndex: 0, Count: 1, FontSize: 20, Scale: 2})
	if !strings.Contains(svg, `font-size="40"`) {
		t.Errorf("fsize=20 scale=2 should set font-size=40: %s", sub(svg, "font-size"))
	}
}

// default font size is 16 when fsize/scale not set.
func TestRenderDefaultFontSize(t *testing.T) {
	th := fakeTheme("fake", 1)
	svg, _ := Render(th, RenderParams{FrameIndex: 0, Count: 1})
	if !strings.Contains(svg, `font-size="16"`) {
		t.Errorf("default font-size should be 16: %s", sub(svg, "font-size"))
	}
}

func TestRenderNilTheme(t *testing.T) {
	if _, err := Render(nil, RenderParams{}); err == nil {
		t.Error("expected error for nil theme")
	}
}

func TestRenderNoFrames(t *testing.T) {
	th := &Theme{Name: "empty"}
	if _, err := Render(th, RenderParams{}); err == nil {
		t.Error("expected error for theme with no frames")
	}
}

func TestRenderFrameIndexOutOfRange(t *testing.T) {
	th := fakeTheme("fake", 1)
	if _, err := Render(th, RenderParams{FrameIndex: 5}); err == nil {
		t.Error("expected error for out-of-range frame index")
	}
}

func TestRenderNumberOverride(t *testing.T) {
	th := fakeTheme("fake", 1)
	svg, _ := Render(th, RenderParams{FrameIndex: 0, Count: 1, Number: 99})
	if !strings.Contains(svg, `>99<`) {
		t.Errorf("number override should show 99: %s", sub(svg, "text"))
	}
}

func TestRenderLayerOrderFrameBelowText(t *testing.T) {
	th := fakeTheme("fake", 1)
	svg, _ := Render(th, RenderParams{FrameIndex: 0, Count: 7})
	imgIdx := strings.Index(svg, "<image")
	txtIdx := strings.Index(svg, "<text")
	if imgIdx < 0 || txtIdx < 0 {
		t.Fatalf("missing image or text in svg")
	}
	if imgIdx > txtIdx {
		t.Errorf("image must precede text (layer 0 below layer 1)")
	}
}

func sub(s, marker string) string {
	i := strings.Index(s, marker)
	if i < 0 {
		return "(not found)"
	}
	end := i + 60
	if end > len(s) {
		end = len(s)
	}
	return s[i:end]
}
