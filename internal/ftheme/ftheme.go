// Package ftheme defines font-style themes for the counter text. A
// Style is a named set of CSS-ish text properties (font-family, color,
// weight) applied to the <text> overlay. f-theme is to the counter text
// what theme is to the background image: pure presentation, never
// affecting the count value.
package ftheme

// Style is one font-style theme. Fields map directly to SVG <text>
// attributes. Empty fields fall back to the render defaults in
// theme.DefaultFontFamily / theme.DefaultFontColor.
type Style struct {
	Name   string `json:"name"`
	Family string `json:"family,omitempty"` // CSS font-family, e.g. "monospace"
	Color  string `json:"color,omitempty"`  // fill color, e.g. "#e91e63"
	Weight string `json:"weight,omitempty"` // font-weight, e.g. "bold"
}

// Registry resolves a f-theme name to its Style. The reserved value
// "random" is handled by the caller (handler), not by the registry.
type Registry interface {
	// Get returns the style for name, or false if not registered.
	Get(name string) (Style, bool)
	// List returns the names of all registered styles.
	List() []string
}
