// Package theme renders a single themed image frame with the counter
// value overlaid as text. A Theme is an ordered set of frames
// (0.png .. size-1.png); the caller picks a frame index and the package
// composes an SVG = frame image + count text.
//
// This is the M2.5 model: one image per frame, counter value drawn as
// text, not the M2 per-digit-glyph composition.
package theme

// Frame is one decoded image of a theme: pixel dimensions and a base64
// data URI of the original bytes. The data URI is embedded directly into
// the SVG so the counter renders offline (AGENTS.md Iron Rule 2: digit
// images use data URIs).
type Frame struct {
	Width  int
	Height int
	Data   string // data:<mime>;base64,...
}

// Theme is a named, ordered set of frames. Size = len(Frames). The
// counter displays frame[(count+1) % Size] by default.
type Theme struct {
	Name   string
	Frames []Frame
}

// Size returns the number of frames in the theme.
func (t *Theme) Size() int { return len(t.Frames) }

// Frame returns the frame at index, or false if out of range.
func (t *Theme) Frame(index int) (Frame, bool) {
	if index < 0 || index >= len(t.Frames) {
		return Frame{}, false
	}
	return t.Frames[index], true
}

// Registry resolves a theme name to a Theme. Reserved names "demo" and
// "random" are handled by the caller (handler), not by the registry.
type Registry interface {
	// Get returns the theme for name, or false if not registered.
	Get(name string) (*Theme, bool)
	// List returns the names of all registered themes.
	List() []string
}
