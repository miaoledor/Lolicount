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
