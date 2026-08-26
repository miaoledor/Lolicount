// Package theme defines the runtime data model for a renderable theme:
// an ordered layer stack plus canvas dimensions. Theme structs hold
// pre-decoded image data URIs and layer metadata; they do not render
// SVG (that is the job of internal/imgcore/render).
package theme

// Canvas defines the coordinate space for layer placement. All layer
// X/Y positions are relative to the canvas origin (top-left).
type Canvas struct {
	Width  int
	Height int
}

// DisplayConfig controls the final rendered output size. When set, the
// composer scales the portrait proportionally so the height equals Size.
// Crop optionally trims blank canvas margins so only the portrait area
// is shown. Migrated from characterthemedrawer.DisplaySize.
type DisplayConfig struct {
	Size int       `json:"size"`
	Crop *CropRect `json:"crop,omitempty"`
}

// CropRect defines a sub-rectangle of the canvas to display. When set,
// the composer maps only this region to the output viewport, trimming
// blank margins around the portrait. Migrated from
// characterthemedrawer.CropRect.
type CropRect struct {
	Left   int `json:"left"`
	Top    int `json:"top"`
	Width  int `json:"width"`
	Height int `json:"height"`
}
