package theme

// TextStyle controls the font-family, color, and weight of a text layer.
// Empty fields fall back to the render defaults (monospace / #333).
// 
type TextStyle struct {
	Family string `json:"family,omitempty"`
	Color  string `json:"color,omitempty"`
	Weight string `json:"weight,omitempty"`
}

// TextPos expresses text placement. Exactly one mode is active: default
// (below image, centered), pixel (absolute X/Y), or ratio (RX/RY as
// 0..1 fractions of the image width/height). Migrated from
type TextPos struct {
	X  int     `json:"x,omitempty"`
	Y  int     `json:"y,omitempty"`
	RX float64 `json:"rx,omitempty"`
	RY float64 `json:"ry,omitempty"`
}

// FStyle is one named font-style theme, loaded from assets/f-theme
// JSON. Fields map directly to SVG <text> attributes. Migrated from
type FStyle struct {
	Name   string `json:"name"`
	Family string `json:"family,omitempty"`
	Color  string `json:"color,omitempty"`
	Weight string `json:"weight,omitempty"`
}

// Text render defaults. These are the single source of truth for
// text-layer render defaults

const (
	// DefaultFontSize is the counter text size in pixels when fsize is
	// not set. Font sizing is independent of image Scale.
	DefaultFontSize = 16

	// MonoCharWidthFactor approximates the advance width of one
	// monospace digit relative to font-size (~0.6em).
	MonoCharWidthFactor = 0.6

	// DefaultFontFamily is the CSS font-family used for counter text.
	DefaultFontFamily = "monospace"

	// DefaultFontColor is the fill color of the counter text.
	DefaultFontColor = "#333"

	// TextGapBelowImage is the extra pixels between the image bottom
	// and the counter text baseline, on top of the font size.
	TextGapBelowImage = 4
)
