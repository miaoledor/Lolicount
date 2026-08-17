// Package theme renders digit-glyph counter images as SVG.
// A Theme is a set of named glyphs (0..9 plus optional _start/_end
// decorations) decoded from the embedded assets/theme tree.
package theme

import "image"

// CharName is a glyph slot key. Digits "0".."9" are required; "_start"
// and "_end" are optional decorations prepended/appended to the output.
type CharName string

// ThemeChar is a single decoded glyph: its pixel dimensions and a base64
// data URI carrying the original bytes. The data URI is embedded directly
// into the SVG so the counter renders offline (AGENTS.md Iron Rule 2).
type ThemeChar struct {
	Width  int
	Height int
	Data   string // data:<mime>;base64,...
}

// Theme is a named collection of glyphs keyed by CharName.
type Theme struct {
	Name  string
	Chars map[CharName]ThemeChar
}

// Dimensions returns the glyph's width/height, or zero for a missing char.
func (t *Theme) Dimensions(name CharName) (w, h int, ok bool) {
	c, exists := t.Chars[name]
	if !exists {
		return 0, 0, false
	}
	return c.Width, c.Height, true
}

// Lookup returns the glyph for name, or zero false if absent.
func (t *Theme) Lookup(name CharName) (ThemeChar, bool) {
	c, ok := t.Chars[name]
	return c, ok
}

// Names returns all glyph names in the theme (order is not guaranteed).
func (t *Theme) Names() []CharName {
	out := make([]CharName, 0, len(t.Chars))
	for k := range t.Chars {
		out = append(out, k)
	}
	return out
}

// Point is an integer pixel coordinate pair.
type Point image.Point

// Registry resolves a theme name to a Theme. Reserved names "demo" and
// "random" are handled by the caller (handler), not by the registry:
// the registry only owns concrete built-in (and later user-uploaded)
// themes.
type Registry interface {
	// Get returns the theme for name, or false if not registered.
	Get(name string) (*Theme, bool)
	// List returns the names of all registered themes.
	List() []string
}
