package server

import (
	"testing"

	"github.com/miaoledor/lolicount/internal/imgcore"
	"github.com/miaoledor/lolicount/internal/imgcore/imgutils"
	"github.com/miaoledor/lolicount/internal/imgcore/render"
	"github.com/miaoledor/lolicount/internal/imgcore/theme"
)

// TestScaleOrOne verifies the scale fallback.
func TestScaleOrOne(t *testing.T) {
	if v := scaleOrOne(0); v != 1 {
		t.Errorf("scaleOrOne(0) = %v, want 1", v)
	}
	if v := scaleOrOne(-1); v != 1 {
		t.Errorf("scaleOrOne(-1) = %v, want 1", v)
	}
	if v := scaleOrOne(1.5); v != 1.5 {
		t.Errorf("scaleOrOne(1.5) = %v, want 1.5", v)
	}
}

// makeMultiFrameTheme builds a *theme.Theme whose single layer is a
// RandomPickLayer with the given source frame dimensions. This mirrors
// how FrameThemeToTheme produces a multi-frame card theme.
func makeMultiFrameTheme(name string, frames []struct{ W, H int }) *theme.Theme {
	opts := make([]render.ImageOption, len(frames))
	for i, f := range frames {
		opts[i] = render.ImageOption{
			ImageLayer: render.ImageLayer{
				Src:       "data:image/png;base64,",
				Width:     f.W,
				Height:    f.H,
				Transform: imgcore.DefaultTransform(),
			},
			Weight: 1,
		}
	}
	return &theme.Theme{
		Name:   name,
		Canvas: theme.Canvas{Width: frames[0].W, Height: frames[0].H},
		BgW:    frames[0].W,
		BgH:    frames[0].H,
		Layers: []imgcore.Layer{
			&render.RandomPickLayer{
				Category:  name,
				Options:   opts,
				Transform: imgcore.DefaultTransform(),
				Z:         0,
			},
		},
	}
}

// TestBuildThemeLayersRandomPickScale verifies that a multi-frame card
// theme (RandomPickLayer) gets its frame dimensions scaled to the
// default display size, matching the behavior of single-frame
// ImageLayer themes. Before the fix, RandomPickLayer options were left
// at their original pixel dimensions, producing oversized output.
func TestBuildThemeLayersRandomPickScale(t *testing.T) {
	srcW, srcH := 800, 1200
	base := makeMultiFrameTheme("test-multi", []struct{ W, H int }{
		{srcW, srcH},
		{srcW, srcH},
	})

	got, err := buildThemeLayers(base, 0, "123", 50, false, theme.TextStyle{}, theme.TextPos{})
	if err != nil {
		t.Fatalf("buildThemeLayers: %v", err)
	}

	wantW, wantH := imgutils.ScaledDims(srcW, srcH, imgutils.DefaultDisplaySize)
	if got.BgW != wantW || got.BgH != wantH {
		t.Errorf("BgW/BgH = %dx%d, want %dx%d", got.BgW, got.BgH, wantW, wantH)
	}

	rpl, ok := got.Layers[0].(*render.RandomPickLayer)
	if !ok {
		t.Fatalf("layer 0 is %T, want *RandomPickLayer", got.Layers[0])
	}
	for i, opt := range rpl.Options {
		if opt.Width != wantW || opt.Height != wantH {
			t.Errorf("option %d dims = %dx%d, want %dx%d", i, opt.Width, opt.Height, wantW, wantH)
		}
	}
}

// TestBuildThemeLayersDoesNotMutateRegistry verifies that calling
// buildThemeLayers does not mutate the shared registry theme's layer
// dimensions across repeated calls. This guards against a regression
// where ImageLayer Width/Height were modified in place.
func TestBuildThemeLayersDoesNotMutateRegistry(t *testing.T) {
	srcW, srcH := 800, 1200
	base := makeMultiFrameTheme("test-nomut", []struct{ W, H int }{
		{srcW, srcH},
	})

	for i := 0; i < 3; i++ {
		_, _ = buildThemeLayers(base, 0, "1", 50, false, theme.TextStyle{}, theme.TextPos{})
	}

	rpl, ok := base.Layers[0].(*render.RandomPickLayer)
	if !ok {
		t.Fatalf("registry layer is %T, want *RandomPickLayer", base.Layers[0])
	}
	for i, opt := range rpl.Options {
		if opt.Width != srcW || opt.Height != srcH {
			t.Errorf("registry option %d mutated to %dx%d, want original %dx%d",
				i, opt.Width, opt.Height, srcW, srcH)
		}
	}
	if base.BgW != srcW || base.BgH != srcH {
		t.Errorf("registry BgW/BgH mutated to %dx%d, want %dx%d", base.BgW, base.BgH, srcW, srcH)
	}
}

// TestBuildThemeLayersRandomPickCanvasFitsAllFrames verifies that the
// canvas (BgW/BgH) for a multi-frame RandomPickLayer theme is at least
// as large as every frame's scaled dimensions. Frames in a multi-frame
// theme can have different aspect ratios, so the canvas must use the
// max scaled dims — otherwise wider/taller frames get clipped by the
// viewBox. This is a regression test for a bug introduced during the
// theme unification refactor where only the first frame's dims were
// used for the canvas.
func TestBuildThemeLayersRandomPickCanvasFitsAllFrames(t *testing.T) {
	frames := []struct{ W, H int }{
		{508, 512}, // frame 0: tall
		{512, 497}, // frame 1: wide
		{512, 468}, // frame 2: shorter
		{499, 512}, // frame 3: narrow + tall
	}
	base := makeMultiFrameTheme("test-fit", frames)

	got, err := buildThemeLayers(base, 0, "1", 50, false, theme.TextStyle{}, theme.TextPos{})
	if err != nil {
		t.Fatalf("buildThemeLayers: %v", err)
	}

	rpl, ok := got.Layers[0].(*render.RandomPickLayer)
	if !ok {
		t.Fatalf("layer 0 is %T, want *RandomPickLayer", got.Layers[0])
	}

	for i, opt := range rpl.Options {
		if opt.Width > got.BgW {
			t.Errorf("frame %d: width %d > BgW %d (will be clipped)", i, opt.Width, got.BgW)
		}
		if opt.Height > got.BgH {
			t.Errorf("frame %d: height %d > BgH %d (will be clipped)", i, opt.Height, got.BgH)
		}
	}
}
