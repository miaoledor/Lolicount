package theme

import "github.com/miaoledor/lolicount/internal/imgcore"

// DefaultTheme is the theme used when the request omits ?theme=.
const DefaultTheme = "lian"

// Theme is a named, ordered layer stack plus its canvas dimensions.
// Each layer is an imgcore.Layer (image, text, or random-pick). The
// composer iterates Layers in order and concatenates their SVG fragments.
//
// A theme with a single image layer (excluding the text layer) is a
// card theme; a theme with multiple image/random-pick layers is a
// character theme. This distinction is inferred at save time, not
// stored as a field — see IsCardTheme.
type Theme struct {
	Name    string
	Canvas  Canvas
	Display *DisplayConfig
	Layers  []imgcore.Layer
}

// IsCardTheme reports whether the theme has at most one image layer
// (excluding text layers). A single-layer theme is a card theme;
// multiple layers make it a character theme. This matches the
// edit-design spec: "one layer (excluding text) => card theme;
// multiple layers => character theme".
func (t *Theme) IsCardTheme() bool {
	count := 0
	for _, l := range t.Layers {
		if l.Kind() == imgcore.LayerImage || l.Kind() == imgcore.LayerRandomPick {
			count++
		}
	}
	return count <= 1
}

// FThemeRegistry is the font-style registry interface. Migrated from
// fdrawer.Registry.
type FThemeRegistry interface {
	Get(name string) (FStyle, bool)
	List() []string
}
