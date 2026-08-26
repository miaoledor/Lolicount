package theme

import "github.com/miaoledor/lolicount/internal/imgcore"

// DefaultTheme is the theme used when the request omits ?theme=.
const DefaultTheme = "lian"

// Theme is a named, ordered layer stack plus its canvas dimensions.
// Each layer is an imgcore.Layer (image, text, or random-pick). The
// composer iterates Layers in order and concatenates their SVG fragments.
//
// The render path does not distinguish card vs character themes —
// all themes go through the same compose path.
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

// FThemeRegistry is the font-style registry interface.
type FThemeRegistry interface {
	Get(name string) (FStyle, bool)
	List() []string
}
