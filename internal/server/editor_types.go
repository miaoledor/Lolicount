package server

import "github.com/miaoledor/lolicount/internal/imgcore/theme"

// EditorImage is one candidate image in an editor layer. A layer with
// multiple images becomes a GroupPart with Candidates (random pick at
// render time). Each image carries absolute PSD coordinates on the
// canvas.
type EditorImage struct {
	Src    string `json:"src"`    // data URI (base64) or URL
	Left   int    `json:"left"`   // canvas x position
	Top    int    `json:"top"`    // canvas y position
	Width  int    `json:"width"`  // pixel width
	Height int    `json:"height"` // pixel height
}

// EditorLayer is one layer in the editor's layer stack. Each layer maps
// to one GroupPart inside a single GroupLayer. The Name field is used
// at export time as the manifest entry name and for config.json range
// grouping.
type EditorLayer struct {
	ID       int           `json:"id"`
	Name     string        `json:"name"`
	ZIndex   int           `json:"zIndex"`
	Fixed    bool          `json:"fixed"`
	Images   []EditorImage `json:"images"`
}

// EditorCanvas is the PSD coordinate space for layer placement.
type EditorCanvas struct {
	Width  int `json:"width"`
	Height int `json:"height"`
}

// EditorRequest is the JSON body for POST /api/editor/preview and
// POST /api/editor/export. It carries the full editor state: canvas
// dimensions, optional display config, and the ordered layer stack.
// Text layer parameters are sent alongside (text layer is not persisted
// in the editor — it is per-user preview only).
type EditorRequest struct {
	Name    string             `json:"name"`
	Canvas  EditorCanvas       `json:"canvas"`
	Display *theme.DisplayConfig `json:"display,omitempty"`
	Layers  []EditorLayer      `json:"layers"`
	Text    string             `json:"text"`       // counter text to preview
	FSize   int                `json:"fsize"`      // font size (0 = default)
	Scale   float64            `json:"scale"`      // image scale (0 = 1)
	UnshowF bool               `json:"unshowf"`    // hide text
}
