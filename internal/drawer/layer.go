// Package drawer defines the shared Layer contract and common types
// used by all drawer sub-packages. Each drawer produces a Layer; the
// renderer merges layers into the final SVG. Drawers never import each
// other — they only share this root package for the Layer type.
package drawer

// Layer is the output of one drawer: an SVG fragment (element strings
// without the outer <svg> wrapper) plus the canvas width/height it
// occupies. The renderer uses Width/Height to compute the final viewBox
// and concatenates Fragment strings in layer order.
type Layer struct {
	Fragment string
	Width    int
	Height   int
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
