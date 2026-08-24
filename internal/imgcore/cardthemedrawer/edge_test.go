package cardthemedrawer

import (
	"testing"
)

// FrameAt returns false for out-of-range and negative indices.
func TestFrameAtOutOfRange(t *testing.T) {
	th := &Theme{Name: "t", Frames: []Frame{{Width: 1, Height: 1, Data: "a"}}}
	if _, ok := th.FrameAt(1); ok {
		t.Error("index 1 should be out of range for 1-frame theme")
	}
	if _, ok := th.FrameAt(-1); ok {
		t.Error("negative index should return false")
	}
}

// FrameAt returns the correct frame.
func TestFrameAtValid(t *testing.T) {
	th := &Theme{Name: "t", Frames: []Frame{
		{Width: 1, Height: 1, Data: "a"},
		{Width: 2, Height: 2, Data: "b"},
	}}
	f, ok := th.FrameAt(1)
	if !ok {
		t.Fatal("FrameAt(1) should succeed")
	}
	if f.Data != "b" {
		t.Errorf("got %q, want b", f.Data)
	}
}

// Size returns the frame count.
func TestThemeSize(t *testing.T) {
	th := &Theme{Name: "t", Frames: []Frame{{}, {}, {}}}
	if th.Size() != 3 {
		t.Errorf("Size = %d, want 3", th.Size())
	}
}

// Draw with negative scale falls back to default.
func TestDrawNegativeScale(t *testing.T) {
	frame := Frame{Width: 10, Height: 20, Data: "data:image/gif;base64,QQ"}
	layer := Draw(frame, -1)
	if layer.Width != 200 || layer.Height != 400 {
		t.Errorf("negative scale should use default: got %dx%d, want 200x400", layer.Width, layer.Height)
	}
}

// Draw with very small scale still produces at least 1x1.
func TestDrawTinyScale(t *testing.T) {
	frame := Frame{Width: 10, Height: 20, Data: "data:image/gif;base64,QQ"}
	layer := Draw(frame, 0.001)
	if layer.Width < 1 || layer.Height < 1 {
		t.Errorf("tiny scale should still be >= 1x1: got %dx%d", layer.Width, layer.Height)
	}
}

// frameIndexFromName rejects non-numeric and negative names.
func TestFrameIndexFromNameInvalid(t *testing.T) {
	cases := []string{"abc.gif", "-1.png", "readme.txt", "1.txt"}
	for _, name := range cases {
		if idx := frameIndexFromName(name); idx >= 0 {
			t.Errorf("frameIndexFromName(%q) = %d, want -1", name, idx)
		}
	}
}

// frameIndexFromName accepts valid numeric names.
func TestFrameIndexFromNameValid(t *testing.T) {
	cases := map[string]int{
		"0.gif":  0,
		"1.png":  1,
		"99.webp": 99,
	}
	for name, want := range cases {
		if idx := frameIndexFromName(name); idx != want {
			t.Errorf("frameIndexFromName(%q) = %d, want %d", name, idx, want)
		}
	}
}

// pathExt extracts file extensions correctly.
func TestPathExt(t *testing.T) {
	cases := map[string]string{
		"0.gif":    ".gif",
		"1.png":    ".png",
		"readme":   "",
		"a/b.webp": ".webp",
	}
	for name, want := range cases {
		if got := pathExt(name); got != want {
			t.Errorf("pathExt(%q) = %q, want %q", name, got, want)
		}
	}
}
