package theme

import "github.com/miaoledor/lolicount/internal/imgcore"

// DefaultTheme is the theme used when the request omits ?theme=.
const DefaultTheme = "lian"

// Theme is a named, ordered layer stack plus its canvas dimensions.
// Each layer is an imgcore.Layer (image, text, or random-pick). The
// composer iterates Layers in order and concatenates their SVG fragments.
//
// The render path does not distinguish card vs character themes —
// all themes go through the same compose path. IsCardTheme() is used
// only by the export standard to decide packaging.
type Theme struct {
	Name    string
	Canvas  Canvas
	Display *DisplayConfig
	// BgW/BgH are the layer-0 background (image) dimensions, excluding
	// the text area. Text layers use these for ratio positioning so
	// rx/ry are fractions of the image. When zero, the composer falls
	// back to Canvas dims.
	BgW     int
	BgH     int
	Layers  []imgcore.Layer
}

// IsCardTheme reports whether the theme has at most one image layer
// (excluding text layers). Used only by the export standard to decide
// whether a theme packages as a card (single-layer) or character
// (multi-layer) bundle. Not used in the render path.
func (t *Theme) IsCardTheme() bool {
	count := 0
	for _, l := range t.Layers {
		if l.Kind() == imgcore.LayerImage || l.Kind() == imgcore.LayerRandomPick {
			count++
		}
	}
	return count <= 1
}

// FThemeRegistry is the font-style registry interface.
type FThemeRegistry interface {
	Get(name string) (FStyle, bool)
	List() []string
}
