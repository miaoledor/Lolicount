package renderer

import (
	"strings"
	"testing"

	"github.com/miaoledor/lolicount/internal/imgcore"
	"github.com/miaoledor/lolicount/internal/imgcore/cardthemedrawer"
)

// FrameIndexForCount must never return a negative index, even for
// negative counts (which can arise from int64 overflow).
func TestFrameIndexForCountNegative(t *testing.T) {
	cases := []struct {
		count int64
		size  int
	}{
		{-1, 3},
		{-5, 3},
		{-100, 7},
		{-1, 1},  // size=1 → always 0
		{-1, 0},  // size=0 → 0
		{-1, -1}, // negative size → 0
	}
	for _, tc := range cases {
		idx := FrameIndexForCount(tc.count, tc.size)
		if idx < 0 {
			t.Errorf("FrameIndexForCount(%d, %d) = %d, must be >= 0", tc.count, tc.size, idx)
		}
		if tc.size > 1 && idx >= tc.size {
			t.Errorf("FrameIndexForCount(%d, %d) = %d, must be < %d", tc.count, tc.size, idx, tc.size)
		}
	}
}

// FrameIndexForCount wraps correctly for large counts.
func TestFrameIndexForCountLarge(t *testing.T) {
	// count = int64 max - 1, size = 5 → should wrap to a valid index.
	idx := FrameIndexForCount(1<<62, 5)
	if idx < 0 || idx >= 5 {
		t.Errorf("large count index %d out of [0,5)", idx)
	}
}

// PickFrame with an out-of-range seq index returns false.
func TestPickFrameOutOfRange(t *testing.T) {
	th := &cardthemedrawer.Theme{Name: "fake", Frames: distinctFrames()}
	_, ok := PickFrame(th, imgcore.ModeSeq, 99, nil)
	if ok {
		t.Error("out-of-range index should return false")
	}
}

// PickFrame with a negative seq index returns false.
func TestPickFrameNegativeIndex(t *testing.T) {
	th := &cardthemedrawer.Theme{Name: "fake", Frames: distinctFrames()}
	_, ok := PickFrame(th, imgcore.ModeSeq, -1, nil)
	if ok {
		t.Error("negative index should return false")
	}
}

// PickFrame with an empty theme returns false for both modes.
func TestPickFrameEmptyTheme(t *testing.T) {
	th := &cardthemedrawer.Theme{Name: "empty", Frames: nil}
	_, okSeq := PickFrame(th, imgcore.ModeSeq, 0, nil)
	if okSeq {
		t.Error("empty theme seq should return false")
	}
	_, okRand := PickFrame(th, imgcore.ModeRandom, 0, nil)
	if okRand {
		t.Error("empty theme random should return false")
	}
}

// Render with empty text still produces valid SVG.
func TestRenderEmptyText(t *testing.T) {
	svg, err := fakeCardRender("")
	if err != nil {
		t.Fatalf("Render: %v", err)
	}
	if !strings.Contains(svg, "<svg") {
		t.Error("empty text should still produce SVG")
	}
}

// Render with scale=0 uses the uniform default (not 0).
func TestRenderScaleZero(t *testing.T) {
	svg, err := fakeCardRender("1", func(p *RenderParams) { p.Scale = 0 })
	if err != nil {
		t.Fatalf("Render: %v", err)
	}
	// 10x20 frame, scale=0 → default 400 longest edge → 200x400.
	if !strings.Contains(svg, `width="200" height="400"`) {
		t.Errorf("scale=0 should use default size: %s", sub(svg, "image"))
	}
}

// Render with negative scale also falls back to default (DisplaySize
// treats <=0 as default).
func TestRenderNegativeScale(t *testing.T) {
	svg, err := fakeCardRender("1", func(p *RenderParams) { p.Scale = -5 })
	if err != nil {
		t.Fatalf("Render: %v", err)
	}
	if !strings.Contains(svg, `width="200" height="400"`) {
		t.Errorf("negative scale should use default size: %s", sub(svg, "image"))
	}
}

// Render centers the background when text is wider than the image.
func TestRenderBackgroundCenteredWhenTextWider(t *testing.T) {
	svg, _ := Render(RenderParams{
		ThemeKind: imgcore.KindFrame,
		Frame:     cardthemedrawer.Frame{Width: 10, Height: 20, Data: "data:image/gif;base64,QQ"},
		Text:      "12345678901234567890",
		FontSize:  100,
	})
	// textW = 20*100*0.6 + 60 = 1260; canvasWidth = max(200,1260) = 1260.
	// imgX = (1260 - 200) / 2 = 530.
	if !strings.Contains(svg, `translate(530,0)`) {
		t.Errorf("background should be centered with translate(530,0): %s", sub(svg, "translate"))
	}
}

// Render does NOT wrap in <g> when text is not wider than the image.
func TestRenderNoTranslateWhenTextNarrower(t *testing.T) {
	svg, _ := fakeCardRender("1")
	if strings.Contains(svg, "translate(") {
		t.Errorf("narrow text should not need translate: %s", sub(svg, "translate"))
	}
}

// drawerRandomInt with nil rand uses global source (no panic).
func TestDrawerRandomIntNilRand(t *testing.T) {
	// Should not panic with nil rand.
	v := drawerRandomInt(nil, 5)
	if v < 0 || v >= 5 {
		t.Errorf("drawerRandomInt(nil, 5) = %d, want [0,5)", v)
	}
}

// drawerRandomInt with n<=0 returns 0.
func TestDrawerRandomIntZeroN(t *testing.T) {
	if v := drawerRandomInt(nil, 0); v != 0 {
		t.Errorf("drawerRandomInt(nil, 0) = %d, want 0", v)
	}
	if v := drawerRandomInt(nil, -1); v != 0 {
		t.Errorf("drawerRandomInt(nil, -1) = %d, want 0", v)
	}
}

// FrameIndexForCount: the M2.5 formula (count+1)%size.
func TestFrameIndexForCountFormula(t *testing.T) {
	// count=0, size=4 → (0+1)%4 = 1.
	if idx := FrameIndexForCount(0, 4); idx != 1 {
		t.Errorf("FrameIndexForCount(0,4) = %d, want 1", idx)
	}
	// count=3, size=4 → (3+1)%4 = 0.
	if idx := FrameIndexForCount(3, 4); idx != 0 {
		t.Errorf("FrameIndexForCount(3,4) = %d, want 0", idx)
	}
	// count=7, size=4 → (7+1)%4 = 0.
	if idx := FrameIndexForCount(7, 4); idx != 0 {
		t.Errorf("FrameIndexForCount(7,4) = %d, want 0", idx)
	}
}
