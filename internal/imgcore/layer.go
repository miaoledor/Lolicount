// Package imgcore defines the shared Layer contract and common types
// used by all render sub-packages and the composer. Each render layer
// produces a LayerOutput; the composer (internal/imgcore/composer)
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
// the PRNG (seeded by counter name), the canvas dimensions, the
// background (image) dimensions, and the counter text string (used by
// text layers bound to the counter).
//
// BgW/BgH are the layer-0 background dimensions (image only, excluding
// the text area). Text layers use these for ratio positioning so rx/ry
// are fractions of the image, not the full canvas (matching the original
// fdrawer.Draw semantics). When zero, text layers fall back to
// CanvasW/CanvasH.
type RenderCtx struct {
	PRNG      PRNGSource
	CanvasW   int
	CanvasH   int
	BgW       int
	BgH       int
	CountText string
}

// PRNGSource is the minimal PRNG interface used by layers to resolve
// Range values and make weighted random picks. Implemented by
// imgutils.PRNG.
type PRNGSource interface {
	FloatRange(r Range) float64
	WeightedPick(weights []float64) int
}

// Layer is the unified render-layer contract. Every layer type (image,
// text, random-pick) implements this interface. The composer calls
// Render once per layer in ZIndex order and concatenates the fragments.
type Layer interface {
	Kind() LayerKind
	ZIndex() int
	Fixed() bool
	Render(ctx RenderCtx) LayerOutput
}
