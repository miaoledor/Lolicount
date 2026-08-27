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

// TestGroupLayerRender verifies nested svg output with viewBox mapping.
func TestGroupLayerRender(t *testing.T) {
	l := &GroupLayer{
		Parts: []GroupPart{
			{Src: "data:image/png;base64,a", X: 10, Y: 20, Width: 100, Height: 200},
			{Src: "data:image/png;base64,b", X: 50, Y: 60, Width: 80, Height: 90},
		},
		OutW: 200, OutH: 300,
		VbX: 0, VbY: 0, VbW: 400, VbH: 600,
		Z: 0,
	}
	out := l.Render(makeCtx(""))
	if !strings.Contains(out.Fragment, "<svg") {
		t.Fatal("expected nested <svg> in fragment")
	}
	if !strings.Contains(out.Fragment, `viewBox="0 0 400 600"`) {
		t.Fatal("expected viewBox in nested svg")
	}
	if !strings.Contains(out.Fragment, `width="200" height="300"`) {
		t.Fatal("expected output dimensions on nested svg")
	}
	if !strings.Contains(out.Fragment, "data:image/png;base64,a") {
		t.Fatal("expected first part data URI")
	}
	if !strings.Contains(out.Fragment, "data:image/png;base64,b") {
		t.Fatal("expected second part data URI")
	}
	if out.Width != 200 || out.Height != 300 {
		t.Fatalf("expected 200x300, got %dx%d", out.Width, out.Height)
	}
}

// TestGroupLayerCrop verifies viewBox crop region is applied.
func TestGroupLayerCrop(t *testing.T) {
	l := &GroupLayer{
		Parts: []GroupPart{{Src: "data:image/png;base64,x", X: 137, Y: 323, Width: 50, Height: 50}},
		OutW: 100, OutH: 200,
		VbX: 137, VbY: 323, VbW: 367, VbH: 602,
		Z: 0,
	}
	out := l.Render(makeCtx(""))
	if !strings.Contains(out.Fragment, `viewBox="137 323 367 602"`) {
		t.Fatalf("expected crop viewBox, got: %s", out.Fragment)
	}
}

// TestGroupLayerKind verifies it reports LayerImage kind.
func TestGroupLayerKind(t *testing.T) {
	l := &GroupLayer{}
	if l.Kind() != imgcore.LayerImage {
		t.Fatal("GroupLayer should report LayerImage kind")
	}
}

// TestTextLayerDefaultYPosition verifies text is placed within the canvas
// bounds (not below the viewBox). This is the regression test for the
// text-below-canvas bug where textY was canvasH + fontSize instead of
// canvasH - textH + fontSize.
func TestTextLayerDefaultYPosition(t *testing.T) {
	l := &TextLayer{
		Text:      "123",
		FontSize:  16,
		Transform: imgcore.DefaultTransform(),
	}
	ctx := makeCtx("123")
	// CanvasH=400, textH=16+4=20, so textY should be 400-20+16=396.
	out := l.Render(ctx)
	if !strings.Contains(out.Fragment, `y="396"`) {
		t.Fatalf("expected y=396 (canvasH-textH+fontSize), got: %s", out.Fragment)
	}
}

// TestTextLayerDefaultYWithLargeFont verifies positioning with larger font.
func TestTextLayerDefaultYWithLargeFont(t *testing.T) {
	l := &TextLayer{
		Text:      "99",
		FontSize:  32,
		Transform: imgcore.DefaultTransform(),
	}
	ctx := makeCtx("99")
	// CanvasH=400, textH=32+4=36, so textY should be 400-36+32=396.
	out := l.Render(ctx)
	if !strings.Contains(out.Fragment, `y="396"`) {
		t.Fatalf("expected y=396, got: %s", out.Fragment)
	}
}

// TestTextLayerPixelPosition overrides default placement.
func TestTextLayerPixelPosition(t *testing.T) {
	l := &TextLayer{
		Text:      "X",
		FontSize:  16,
		Position:  theme.TextPos{X: 50, Y: 100},
		Transform: imgcore.DefaultTransform(),
	}
	out := l.Render(makeCtx("X"))
	if !strings.Contains(out.Fragment, `x="50"`) {
		t.Fatalf("expected x=50, got: %s", out.Fragment)
	}
	if !strings.Contains(out.Fragment, `y="116"`) {
		t.Fatalf("expected y=116 (100+16), got: %s", out.Fragment)
	}
	if !strings.Contains(out.Fragment, `text-anchor="start"`) {
		t.Fatal("expected text-anchor=start for pixel position")
	}
}

// TestTextLayerRatioPosition verifies ratio-based placement.
func TestTextLayerRatioPosition(t *testing.T) {
	l := &TextLayer{
		Text:      "X",
		FontSize:  16,
		Position:  theme.TextPos{RX: 0.5, RY: 0.5},
		Transform: imgcore.DefaultTransform(),
	}
	ctx := makeCtx("X")
	// CanvasW=400, CanvasH=400: x=200, y=200+16=216
	out := l.Render(ctx)
	if !strings.Contains(out.Fragment, `x="200"`) {
		t.Fatalf("expected x=200 (400*0.5), got: %s", out.Fragment)
	}
	if !strings.Contains(out.Fragment, `y="216"`) {
		t.Fatalf("expected y=216 (400*0.5+16), got: %s", out.Fragment)
	}
}

// TestTextLayerRotation verifies transform attribute is added.
func TestTextLayerRotation(t *testing.T) {
	l := &TextLayer{
		Text:      "rot",
		FontSize:  16,
		Transform: imgcore.Transform{
			X:        imgcore.FixedRange(0),
			Y:        imgcore.FixedRange(0),
			Scale:    imgcore.FixedRange(1),
			Rotation: imgcore.FixedRange(45),
		},
	}
	out := l.Render(makeCtx("rot"))
	if !strings.Contains(out.Fragment, `transform="rotate(`) {
		t.Fatalf("expected rotation transform, got: %s", out.Fragment)
	}
}

// TestTextLayerEmptyText verifies empty text still renders.
func TestTextLayerEmptyText(t *testing.T) {
	l := &TextLayer{
		Text:      "",
		FontSize:  16,
		Transform: imgcore.DefaultTransform(),
	}
	out := l.Render(makeCtx(""))
	if !strings.Contains(out.Fragment, "<text") {
		t.Fatal("expected <text> element even with empty text")
	}
}

// TestImageLayerRotation verifies rotation transform.
func TestImageLayerRotation(t *testing.T) {
	l := &ImageLayer{
		Src:       "data:image/png;base64,abc",
		Width:     100,
		Height:    100,
		Transform: imgcore.Transform{X: imgcore.FixedRange(0), Y: imgcore.FixedRange(0), Scale: imgcore.FixedRange(1), Rotation: imgcore.FixedRange(90)},
	}
	out := l.Render(makeCtx(""))
	if !strings.Contains(out.Fragment, "rotate(") {
		t.Fatalf("expected rotate transform, got: %s", out.Fragment)
	}
}

// TestImageLayerZeroScale verifies scale=0 defaults to 1.
func TestImageLayerZeroScale(t *testing.T) {
	l := &ImageLayer{
		Src:       "data:image/png;base64,abc",
		Width:     100,
		Height:    200,
		Transform: imgcore.Transform{X: imgcore.FixedRange(0), Y: imgcore.FixedRange(0), Scale: imgcore.FixedRange(0), Rotation: imgcore.FixedRange(0)},
	}
	out := l.Render(makeCtx(""))
	if out.Width != 100 || out.Height != 200 {
		t.Fatalf("expected 100x200 (scale=0 -> 1), got %dx%d", out.Width, out.Height)
	}
}

// TestRandomPickLayerWeighted verifies that higher-weight options are picked
// more frequently. Weight=0 defaults to 1 (JSON optional field convention).
func TestRandomPickLayerWeighted(t *testing.T) {
	l := &RandomPickLayer{
		Options: []ImageOption{
			{ImageLayer: ImageLayer{Src: "data:image/png;base64,light", Width: 10, Height: 10, Transform: imgcore.DefaultTransform()}, Weight: 0.001},
			{ImageLayer: ImageLayer{Src: "data:image/png;base64,heavy", Width: 10, Height: 10, Transform: imgcore.DefaultTransform()}, Weight: 100},
		},
		Transform: imgcore.DefaultTransform(),
	}
	heavyCount := 0
	for i := 0; i < 100; i++ {
		out := l.Render(makeCtx(""))
		if strings.Contains(out.Fragment, "base64,heavy") {
			heavyCount++
		}
	}
	if heavyCount < 90 {
		t.Fatalf("expected heavy option picked >=90/100 times, got %d", heavyCount)
	}
}

// TestRandomPickLayerSingleOption verifies single option always picked.
func TestRandomPickLayerSingleOption(t *testing.T) {
	l := &RandomPickLayer{
		Options: []ImageOption{
			{ImageLayer: ImageLayer{Src: "data:image/png;base64,only", Width: 5, Height: 5, Transform: imgcore.DefaultTransform()}, Weight: 1},
		},
		Transform: imgcore.DefaultTransform(),
	}
	out := l.Render(makeCtx(""))
	if !strings.Contains(out.Fragment, "base64,only") {
		t.Fatal("expected the single option to be picked")
	}
}

// TestMeasureTextVariousSizes verifies measurement across font sizes.
func TestMeasureTextVariousSizes(t *testing.T) {
	tests := []struct {
		text     string
		fontSize int
		unshow   bool
		wantW    int
		wantH    int
	}{
		{"", 16, false, 9, 20},
		{"1", 16, false, 18, 20},
		{"12345", 16, false, 57, 20},
		{"12345", 32, false, 115, 36},
		{"12345", 0, false, 57, 20},
		{"12345", 16, true, 0, 0},
	}
	for _, tt := range tests {
		w, h := MeasureText(tt.text, tt.fontSize, tt.unshow)
		if w != tt.wantW {
			t.Errorf("MeasureText(%q,%d,%v) width=%d, want %d", tt.text, tt.fontSize, tt.unshow, w, tt.wantW)
		}
		if h != tt.wantH {
			t.Errorf("MeasureText(%q,%d,%v) height=%d, want %d", tt.text, tt.fontSize, tt.unshow, h, tt.wantH)
		}
	}
}
