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

// M5.6: without Scale the frame is scaled to the uniform base size
// (longest edge = defaultDisplaySize = 400). A 10x20 frame has longest
// edge 20, so displayed dims = 200x400.
func TestRenderUniformDisplaySize(t *testing.T) {
	th := fakeTheme("fake", 1) // 10x20
	svg, err := Render(th, RenderParams{FrameIndex: 0, Count: 5})
	if err != nil {
		t.Fatalf("Render: %v", err)
	}
	if !strings.HasPrefix(svg, "<?xml") {
		t.Errorf("not xml: %q", svg[:16])
	}
	// Image scaled to 200x400 (longest edge 20 -> 400, aspect kept).
	if !strings.Contains(svg, `width="200" height="400"`) {
		t.Errorf("image not scaled to uniform size: %s", sub(svg, "image"))
	}
}

// M5.6: the count text sits BELOW the image, centered.
func TestRenderTextBelowImage(t *testing.T) {
	th := fakeTheme("fake", 1) // 10x20 -> displayed 200x400
	svg, _ := Render(th, RenderParams{FrameIndex: 0, Count: 5, FontSize: 16})
	// canvas height = imgH(400) + fontSize(16) + 4 = 420.
	if !strings.Contains(svg, `viewBox="0 0 200 420"`) {
		t.Errorf("viewBox should be image+text height: %s", sub(svg, "viewBox"))
	}
	// text y = imgH + fontSize = 400 + 16 = 416.
	if !strings.Contains(svg, `y="416"`) {
		t.Errorf("text should be below image (y=416): %s", sub(svg, "y="))
	}
	if !strings.Contains(svg, `text-anchor="middle"`) {
		t.Errorf("text should be centered")
	}
	// count text 5 present.
	if !strings.Contains(svg, `>5<`) {
		t.Errorf("count text missing:\n%s", svg)
	}
	// Layer order: image before text (background below text).
	if strings.Index(svg, "<image") > strings.Index(svg, "<text") {
		t.Errorf("image must precede text")
	}
}

// M5.6: Scale multiplies the uniform base display size, not the font.
func TestRenderScaleMultipliesImageSize(t *testing.T) {
	th := fakeTheme("fake", 1) // 10x20
	svg, _ := Render(th, RenderParams{FrameIndex: 0, Count: 1, Scale: 2})
	// base longest edge 400 * scale 2 = 800; 10x20 -> 400x800.
	if !strings.Contains(svg, `width="400" height="800"`) {
		t.Errorf("scale=2 should double image display size: %s", sub(svg, "image"))
	}
}

// M5.6: aspect ratio is preserved — a wide frame stays wide.
func TestRenderAspectRatioPreserved(t *testing.T) {
	th := &Theme{Name: "wide", Frames: []Frame{{Width: 2000, Height: 1000, Data: "data:image/gif;base64,QQ"}}}
	svg, _ := Render(th, RenderParams{FrameIndex: 0, Count: 1})
	// longest edge 2000 -> 400, ratio 0.2 -> 400x200.
	if !strings.Contains(svg, `width="400" height="200"`) {
		t.Errorf("wide frame aspect not preserved: %s", sub(svg, "image"))
	}
}

// M5.6: ?unshowf=true omits the counter <text> entirely.
func TestRenderUnshowFont(t *testing.T) {
	th := fakeTheme("fake", 1)
	svg, _ := Render(th, RenderParams{FrameIndex: 0, Count: 42, UnshowFont: true})
	if strings.Contains(svg, "<text") {
		t.Errorf("unshowf=true should omit <text>: %s", svg)
	}
	// canvas height = image height only (no text band).
	if !strings.Contains(svg, `viewBox="0 0 200 400"`) {
		t.Errorf("unshowf canvas should be image-only: %s", sub(svg, "viewBox"))
	}
}

// fsize controls the counter text font-size (independent of image scale).
func TestRenderFontSizeFromParam(t *testing.T) {
	th := fakeTheme("fake", 1)
	svg, _ := Render(th, RenderParams{FrameIndex: 0, Count: 1, FontSize: 40})
	if !strings.Contains(svg, `font-size="40"`) {
		t.Errorf("fsize=40 should set font-size=40: %s", sub(svg, "font-size"))
	}
}

// default font size is 16 when fsize not set.
func TestRenderDefaultFontSize(t *testing.T) {
	th := fakeTheme("fake", 1)
	svg, _ := Render(th, RenderParams{FrameIndex: 0, Count: 1})
	if !strings.Contains(svg, `font-size="16"`) {
		t.Errorf("default font-size should be 16: %s", sub(svg, "font-size"))
	}
}

// A multi-digit count widens the viewBox so the text never overflows.
func TestRenderWideTextWidensViewBox(t *testing.T) {
	th := fakeTheme("fake", 1) // displayed 200x400, font 16
	svg, _ := Render(th, RenderParams{FrameIndex: 0, Count: 123456})
	// text width = 6*16*0.6 + charW(9) = 66; img width 200 > 66, so canvas
	// width stays 200; height = 400+16+4 = 420.
	if !strings.Contains(svg, `viewBox="0 0 200 420"`) {
		t.Errorf("wide text viewBox wrong: %s", sub(svg, "viewBox"))
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

// M5.5: theme IS the background — exactly one <image> (the frame).
func TestRenderSingleImageSingleText(t *testing.T) {
	th := fakeTheme("fake", 1)
	svg, _ := Render(th, RenderParams{FrameIndex: 0, Count: 42})
	if got := strings.Count(svg, "<image"); got != 1 {
		t.Errorf("expected 1 <image> (theme frame only), got %d", got)
	}
	if got := strings.Count(svg, "<text"); got != 1 {
		t.Errorf("expected 1 <text>, got %d", got)
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
