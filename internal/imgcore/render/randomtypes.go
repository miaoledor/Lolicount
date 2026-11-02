package render

import "github.com/miaoledor/lolicount/internal/imgcore"

// ImageOption is one candidate image in a RandomPickLayer. Each option
// is an ImageLayer with an associated Weight for weighted random
// per-category layer selection.
type ImageOption struct {
	ImageLayer
	Weight float64 // selection weight; 0 = never picked; default 1
}

// RandomPickLayer randomly selects one ImageOption from its candidate
// list each render and delegates rendering to the selected option. This
// self-describing layer: the category name and candidates are data, not
// hardcoded ranges.
type RandomPickLayer struct {
	Category  string            // e.g. "brow", "eye", "mouth" (for debugging/metadata)
	Options   []ImageOption     // candidate images
	Transform imgcore.Transform // transform applied to the whole layer
	Z         int
	IsFixed   bool

	// FrameDims carries the pixel dimensions of EVERY candidate of the
	// slot, not just the ones currently in Options. Under the lazy
	// one-image-per-slot strategy Options holds only the resident frame,
	// while the canvas must accommodate all frames — BuildThemeLayers
	// scales these dims and folds them into the canvas max. When empty,
	// Options supplies the active image dimensions directly.
	FrameDims []FrameDim
}

// FrameDim is one candidate's pixel dimensions (metadata only).
type FrameDim struct {
	W, H int
}
