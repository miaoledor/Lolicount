// Package theme renders a single themed image frame with the counter
// value overlaid as text. A Theme is an ordered set of frames
// (0.png .. size-1.png); the caller picks a frame index and the package
// composes an SVG = frame image + count text.
//
// This is the M2.5 model: one image per frame, counter value drawn as
// text, not the M2 per-digit-glyph composition.
//
// M9 introduces two theme kinds:
//   - KindFrame: ordinary frame themes. Support sequential mode
//     (frame[(count+1)%size]) and random mode (a random frame each
//     request).
//   - KindCharacter: character-portrait themes (e.g. 莲/Ren). Built from
//     layered PSD assets (ren.json + ren/*.webp); each request randomly
//     assembles one portrait (clothing + expression). Character themes
//     only support random mode.
//
// Both kinds act as the background (layer 0); the count is <text>
// (layer 1). Both work with all three embed formats since they are just
// /@:name URLs.
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

// Kind classifies how a theme's background image is produced.
type Kind int

const (
	// KindFrame is the ordinary ordered-frame theme (0.png..size-1.png).
	KindFrame Kind = iota
	// KindCharacter is a layered portrait theme assembled randomly per
	// request from ren.json + ren/*.webp (M9).
	KindCharacter
)

// Mode selects how the background frame is chosen for a request.
type Mode int

const (
	// ModeSeq picks frame[(count+1)%size] — the background cycles as the
	// counter grows (M2.5 default).
	ModeSeq Mode = iota
	// ModeRandom picks a random frame each request. For character themes
	// this means a freshly assembled portrait each request.
	ModeRandom
)

// Theme is a named theme. For KindFrame, Frames is the ordered set and
// Size = len(Frames); the counter displays frame[(count+1) % Size] by
// default. For KindCharacter, Frames holds the layered portrait parts
// and a portrait is assembled per request.
type Theme struct {
	Name   string
	Kind   Kind
	Frames []Frame
	// Character holds the layered-portrait data for KindCharacter themes.
	// nil for KindFrame themes.
	Character *Character
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
