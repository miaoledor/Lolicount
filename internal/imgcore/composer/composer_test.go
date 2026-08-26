package composer

import (
	"strings"
	"testing"

	"github.com/miaoledor/lolicount/internal/imgcore"
	"github.com/miaoledor/lolicount/internal/imgcore/render"
	"github.com/miaoledor/lolicount/internal/imgcore/theme"
)

// TestComposeSingleImage verifies composing a single image layer (card theme).
func TestComposeSingleImage(t *testing.T) {
	th := &theme.Theme{
		Name:   "test-card",
		Canvas: theme.Canvas{Width: 100, Height: 200},
		Layers: []imgcore.Layer{
			&render.ImageLayer{
				Src:       "data:image/png;base64,abc",
				Width:     100,
				Height:    200,
				Transform: imgcore.DefaultTransform(),
				Z:         0,
			},
		},
	}

	svg, err := Compose(ComposeParams{Theme: th, Seed: "test", CountText: "42"})
	if err != nil {
		t.Fatalf("Compose failed: %v", err)
	}
	if !strings.HasPrefix(svg, `<?xml`) {
		t.Fatal("expected XML declaration")
	}
	if !strings.Contains(svg, `<svg`) {
		t.Fatal("expected <svg> element")
	}
	if !strings.Contains(svg, "data:image/png;base64,abc") {
		t.Fatal("expected image data URI in SVG")
	}
}

// TestComposeWithText verifies composing image + text layers.
func TestComposeWithText(t *testing.T) {
	th := &theme.Theme{
		Name:   "test-with-text",
		Canvas: theme.Canvas{Width: 400, Height: 400},
		Layers: []imgcore.Layer{
			&render.ImageLayer{
				Src:       "data:image/png;base64,img",
				Width:     400,
				Height:    400,
				Transform: imgcore.DefaultTransform(),
				Z:         0,
			},
			&render.TextLayer{
				IsCounter: true,
				FontSize:  20,
				Transform: imgcore.DefaultTransform(),
				Z:         1,
			},
		},
	}

	svg, err := Compose(ComposeParams{Theme: th, Seed: "test", CountText: "123"})
	if err != nil {
		t.Fatalf("Compose failed: %v", err)
	}
	if !strings.Contains(svg, "123") {
		t.Fatal("expected counter text '123' in SVG")
	}
	if !strings.Contains(svg, "<image") {
		t.Fatal("expected <image> in SVG")
	}
	if !strings.Contains(svg, "<text") {
		t.Fatal("expected <text> in SVG")
	}
}

// TestComposeNilTheme verifies error on nil theme.
func TestComposeNilTheme(t *testing.T) {
	_, err := Compose(ComposeParams{Theme: nil})
	if err == nil {
		t.Fatal("expected error for nil theme")
	}
}

// TestComposeDeterminism verifies the same seed produces the same SVG.
func TestComposeDeterminism(t *testing.T) {
	th := &theme.Theme{
		Name:   "test-determinism",
		Canvas: theme.Canvas{Width: 100, Height: 100},
		Layers: []imgcore.Layer{
			&render.RandomPickLayer{
				Options: []render.ImageOption{
					{ImageLayer: render.ImageLayer{Src: "data:image/png;base64,a", Width: 10, Height: 10, Transform: imgcore.DefaultTransform()}, Weight: 1},
					{ImageLayer: render.ImageLayer{Src: "data:image/png;base64,b", Width: 10, Height: 10, Transform: imgcore.DefaultTransform()}, Weight: 1},
					{ImageLayer: render.ImageLayer{Src: "data:image/png;base64,c", Width: 10, Height: 10, Transform: imgcore.DefaultTransform()}, Weight: 1},
				},
				Transform: imgcore.DefaultTransform(),
				Z:         0,
			},
		},
	}

	svg1, _ := Compose(ComposeParams{Theme: th, Seed: "same-seed", CountText: "1"})
	svg2, _ := Compose(ComposeParams{Theme: th, Seed: "same-seed", CountText: "1"})
	if svg1 != svg2 {
		t.Fatal("same seed should produce identical SVG")
	}
}

// TestComposeZeroCanvas verifies canvas computed from layers when zero.
func TestComposeZeroCanvas(t *testing.T) {
	th := &theme.Theme{
		Name:   "test-zero-canvas",
		Canvas: theme.Canvas{Width: 0, Height: 0},
		Layers: []imgcore.Layer{
			&render.ImageLayer{
				Src:       "data:image/png;base64,abc",
				Width:     150,
				Height:    250,
				Transform: imgcore.DefaultTransform(),
				Z:         0,
			},
		},
	}

	svg, err := Compose(ComposeParams{Theme: th, Seed: "test", CountText: ""})
	if err != nil {
		t.Fatalf("Compose failed: %v", err)
	}
	if !strings.Contains(svg, `viewBox="0 0 150 250"`) {
		t.Fatalf("expected viewBox computed from layer dims, got: %s", svg)
	}
}
