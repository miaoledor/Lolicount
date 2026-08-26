package render

import "github.com/miaoledor/lolicount/internal/imgcore"

// ImageOption is one candidate image in a RandomPickLayer. Each option
// is an ImageLayer with an associated Weight for weighted random
// selection. Migrated and generalized from characterthemedrawer's
// per-category layer selection.
type ImageOption struct {
	ImageLayer
	Weight float64 // selection weight; 0 = never picked; default 1
}

// RandomPickLayer randomly selects one ImageOption from its candidate
// list each render and delegates rendering to the selected option. This
// replaces characterthemedrawer's Assemble + pickLayer logic with a
// self-describing layer: the category name and candidates are data, not
// hardcoded ranges.
type RandomPickLayer struct {
	Category  string          // e.g. "brow", "eye", "mouth" (for debugging/metadata)
	Options   []ImageOption   // candidate images
	Transform imgcore.Transform // transform applied to the whole layer
	Z         int
	IsFixed   bool
}
