package render

import (
	"strings"
	"testing"

	"github.com/miaoledor/lolicount/internal/imgcore"
	"github.com/miaoledor/lolicount/internal/imgcore/imgutils"
	"github.com/miaoledor/lolicount/internal/imgcore/theme"
)

// makeCtx creates a RenderCtx with a fixed-seed PRNG for deterministic
// test output.
func makeCtx(text string) imgcore.RenderCtx {
	return imgcore.RenderCtx{
		PRNG:      imgutils.NewPRNG("test"),
		CanvasW:   400,
		CanvasH:   400,
		CountText: text,
	}
}

// TestImageLayerRender verifies the basic <image> output.
func TestImageLayerRender(t *testing.T) {
	l := &ImageLayer{
		Src:    "data:image/png;base64,abc",
		Width:  100,
		Height: 200,
		Transform: imgcore.DefaultTransform(),
	}
	out := l.Render(makeCtx(""))
	if !strings.Contains(out.Fragment, "<image") {
		t.Fatal("expected <image> in fragment")
	}
	if !strings.Contains(out.Fragment, "data:image/png;base64,abc") {
		t.Fatal("expected data URI in fragment")
	}
	if out.Width != 100 || out.Height != 200 {
		t.Fatalf("expected 100x200, got %dx%d", out.Width, out.Height)
	}
}

// TestImageLayerScale verifies that Scale transforms the dimensions.
func TestImageLayerScale(t *testing.T) {
	l := &ImageLayer{
		Src:       "data:image/png;base64,abc",
		Width:     100,
		Height:    200,
		Transform: imgcore.Transform{X: imgcore.FixedRange(0), Y: imgcore.FixedRange(0), Scale: imgcore.FixedRange(2), Rotation: imgcore.FixedRange(0)},
	}
	out := l.Render(makeCtx(""))
	if out.Width != 200 || out.Height != 400 {
		t.Fatalf("expected 200x400 (2x scale), got %dx%d", out.Width, out.Height)
	}
}

// TestTextLayerRender verifies basic <text> output.
func TestTextLayerRender(t *testing.T) {
	l := &TextLayer{
		Text:      "hello",
		FontSize:  20,
		Style:     theme.TextStyle{Family: "sans", Color: "#f00"},
		Transform: imgcore.DefaultTransform(),
	}
	out := l.Render(makeCtx(""))
	if !strings.Contains(out.Fragment, "<text") {
		t.Fatal("expected <text> in fragment")
	}
	if !strings.Contains(out.Fragment, "hello") {
		t.Fatal("expected 'hello' in fragment")
	}
	if !strings.Contains(out.Fragment, "sans") {
		t.Fatal("expected font-family 'sans'")
	}
}

// TestTextLayerCounter verifies that IsCounter uses CountText.
func TestTextLayerCounter(t *testing.T) {
	l := &TextLayer{
		IsCounter: true,
		FontSize:  16,
		Transform: imgcore.DefaultTransform(),
	}
	out := l.Render(makeCtx("12345"))
	if !strings.Contains(out.Fragment, "12345") {
		t.Fatal("expected counter text '12345' in fragment")
	}
}

// TestTextLayerUnshowFont verifies empty output when hidden.
func TestTextLayerUnshowFont(t *testing.T) {
	l := &TextLayer{UnshowFont: true}
	out := l.Render(makeCtx(""))
	if out.Fragment != "" {
		t.Fatal("expected empty fragment when UnshowFont")
	}
}

// TestRandomPickLayerRender verifies that one option is selected and rendered.
func TestRandomPickLayerRender(t *testing.T) {
	l := &RandomPickLayer{
		Category: "eye",
		Options: []ImageOption{
			{ImageLayer: ImageLayer{Src: "data:image/png;base64,eye0", Width: 50, Height: 50, Transform: imgcore.DefaultTransform()}, Weight: 1},
			{ImageLayer: ImageLayer{Src: "data:image/png;base64,eye1", Width: 50, Height: 50, Transform: imgcore.DefaultTransform()}, Weight: 1},
		},
		Transform: imgcore.DefaultTransform(),
	}
	ctx := makeCtx("")
	out := l.Render(ctx)
	if !strings.Contains(out.Fragment, "<image") {
		t.Fatal("expected <image> from selected option")
	}
	if !strings.Contains(out.Fragment, "eye0") && !strings.Contains(out.Fragment, "eye1") {
		t.Fatal("expected one of the eye options in fragment")
	}
}

// TestRandomPickLayerDeterminism verifies the same seed picks the same option.
func TestRandomPickLayerDeterminism(t *testing.T) {
	l := &RandomPickLayer{
		Options: []ImageOption{
			{ImageLayer: ImageLayer{Src: "data:image/png;base64,a", Width: 10, Height: 10, Transform: imgcore.DefaultTransform()}, Weight: 1},
			{ImageLayer: ImageLayer{Src: "data:image/png;base64,b", Width: 10, Height: 10, Transform: imgcore.DefaultTransform()}, Weight: 1},
			{ImageLayer: ImageLayer{Src: "data:image/png;base64,c", Width: 10, Height: 10, Transform: imgcore.DefaultTransform()}, Weight: 1},
		},
		Transform: imgcore.DefaultTransform(),
	}
	out1 := l.Render(makeCtx(""))
	out2 := l.Render(makeCtx(""))
	if out1.Fragment != out2.Fragment {
		t.Fatal("same seed should produce identical output")
	}
}

// TestRandomPickLayerEmpty verifies empty output with no options.
func TestRandomPickLayerEmpty(t *testing.T) {
	l := &RandomPickLayer{Options: nil}
	out := l.Render(makeCtx(""))
	if out.Fragment != "" {
		t.Fatal("expected empty fragment with no options")
	}
}

// TestLayerKindValues verifies the Kind accessors.
func TestLayerKindValues(t *testing.T) {
	if (&ImageLayer{}).Kind() != imgcore.LayerImage {
		t.Fatal("ImageLayer kind mismatch")
	}
	if (&TextLayer{}).Kind() != imgcore.LayerText {
		t.Fatal("TextLayer kind mismatch")
	}
	if (&RandomPickLayer{}).Kind() != imgcore.LayerRandomPick {
		t.Fatal("RandomPickLayer kind mismatch")
	}
}

// TestArcPath verifies arc path generation.
func TestArcPath(t *testing.T) {
	p := ArcPath(100, 100, 50, 0, 90)
	if !strings.Contains(p, "M ") {
		t.Fatal("expected path starting with M")
	}
	if !strings.Contains(p, " A ") {
		t.Fatal("expected arc command A")
	}
}

// TestArcPathZeroRadius verifies empty path for zero radius.
func TestArcPathZeroRadius(t *testing.T) {
	if ArcPath(100, 100, 0, 0, 90) != "" {
		t.Fatal("expected empty path for zero radius")
	}
}
