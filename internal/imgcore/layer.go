// Package imgcore defines the shared Layer contract and common types
// used by all render sub-packages and the composer. Each render layer
// produces a LayerOutput; the composer (internal/imgcore/composer.go)
// merges layer outputs into the final SVG. Render layers never import
// each other — they only share this root package for the Layer type.
package imgcore

// Range is a closed numeric interval. Min == Max represents a fixed
// value; Min < Max represents a random range resolved by the PRNG at
// render time (seeded by the counter name so the same name yields the
// same random result across requests).
type Range struct {
	Min float64
	Max float64
}

// FixedRange creates a Range with a single value.
func FixedRange(v float64) Range {
	return Range{Min: v, Max: v}
}

// Transform holds the spatial transform applied to a layer. Each field
// is a Range so positions, scales, and rotations can be either fixed or
// random. The PRNG resolves each Range to a concrete value at render
// time.
type Transform struct {
	X        Range // canvas x position (top-left corner)
	Y        Range // canvas y position (top-left corner)
	Scale    Range // uniform scale factor, default 1
	Rotation Range // rotation in degrees, default 0
}

// DefaultTransform returns a Transform with no offset, unit scale, and
// zero rotation.
func DefaultTransform() Transform {
	return Transform{
		X:        FixedRange(0),
		Y:        FixedRange(0),
		Scale:    FixedRange(1),
		Rotation: FixedRange(0),
	}
}

// LayerKind classifies a render layer.
type LayerKind int

const (
	// LayerImage is a single fixed image layer.
	LayerImage LayerKind = iota
	// LayerText is a text layer (counter text or static label).
	LayerText
	// LayerRandomPick is a layer that randomly selects one image from a
	// set of candidates each render.
	LayerRandomPick
)

// LayerOutput is the rendered fragment of one layer: an SVG fragment
// (element strings without the outer <svg> wrapper) plus the canvas
// width/height it occupies.
type LayerOutput struct {
	Fragment string
	Width    int
	Height   int
}

// RenderCtx carries the per-render context shared across all layers:
// the PRNG (seeded by counter name), the canvas dimensions, and the
// counter text string (used by text layers bound to the counter).
type RenderCtx struct {
	PRNG      PRNGSource
	CanvasW   int
	CanvasH   int
	CountText string
}

// PRNGSource is the minimal PRNG interface used by layers to resolve
// Range values and make weighted random picks. Implemented by
// imgutils.PRNG.
type PRNGSource interface {
	// FloatRange resolves a Range to a concrete float64.
	FloatRange(r Range) float64
	// WeightedPick returns the index of a randomly selected item from a
	// slice of weights. Returns 0 when weights are empty or all zero.
	WeightedPick(weights []float64) int
}

// Layer is the unified render-layer contract. Every layer type (image,
// text, random-pick) implements this interface. The composer calls
// Render once per layer in ZIndex order and concatenates the fragments.
type Layer interface {
	// Kind classifies the layer for introspection (image/text/randomPick).
	Kind() LayerKind
	// ZIndex returns the stack order (0 = bottom, higher = on top).
	ZIndex() int
	// Fixed reports whether the layer cannot be deleted (e.g. the base
	// image layer or the top text layer in the editor).
	Fixed() bool
	// Render produces the SVG fragment for this layer given the render
	// context.
	Render(ctx RenderCtx) LayerOutput
}

// LegacyLayer is the old struct-based layer output kept for backward
// compatibility during the incremental migration to the Layer interface.
// New code should use LayerOutput and implement the Layer interface.
// This will be removed once all old drawer packages are migrated.
type LegacyLayer = LayerOutput

// LegacyKind classifies how a theme's background image is produced.
// Kept for backward compatibility during migration; new code uses
// LayerKind and theme.IsCardTheme() instead.
type LegacyKind int

const (
	// LegacyKindFrame is the ordinary ordered-frame theme.
	LegacyKindFrame LegacyKind = iota
	// LegacyKindCharacter is a layered portrait theme.
	LegacyKindCharacter
)

// LegacyMode selects how the background frame is chosen for a request.
// Kept for backward compatibility during migration.
type LegacyMode int

const (
	// LegacyModeSeq picks frame[(count+1)%size].
	LegacyModeSeq LegacyMode = iota
	// LegacyModeRandom picks a random frame each request.
	LegacyModeRandom
)
