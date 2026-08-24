// Package fdrawer owns the font-style theme assets and the layer-1 text
// drawer. It loads font-style themes from the embedded assets/f-theme
// tree, provides a Registry for name lookup, and draws the counter text
// as a <text> Layer.
//
// This package merged the former internal/ftheme package (Style,
// Registry, builtin loader) and the text-drawing half of composeSVG /
// composeCharacterSVG — which were previously duplicated across two
// render functions.
package fdrawer

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"sort"
	"strings"

	"github.com/miaoledor/lolicount/assets"
	"github.com/miaoledor/lolicount/internal/imgcore"
	"github.com/miaoledor/lolicount/internal/imgcore/imgutils"
)

// FontStyle is the font-style overlay applied to the counter <text>.
// Empty fields fall back to DefaultFontFamily / DefaultFontColor.
type FontStyle struct {
	Family string
	Color  string
	Weight string
}

// TextPos expresses the counter text placement. Exactly one mode is
// active: default (below image, centered), pixel (absolute X/Y), or
// ratio (RX/RY as 0..1 fractions of the image width/height).
type TextPos struct {
	X  int     // absolute pixel x (left edge of text block)
	Y  int     // absolute pixel y (top edge of text block)
	RX float64 // ratio x, 0..1 of image width
	RY float64 // ratio y, 0..1 of image height
}

// Params controls how the counter text is drawn.
type Params struct {
	Text       string
	FontSize   int
	UnshowFont bool
	FontStyle  FontStyle
	Position   TextPos
}

// Style is one named font-style theme. Fields map directly to SVG <text>
// attributes. Empty fields fall back to the render defaults.
type Style struct {
	Name   string `json:"name"`
	Family string `json:"family,omitempty"`
	Color  string `json:"color,omitempty"`
	Weight string `json:"weight,omitempty"`
}

// Registry resolves a f-theme name to its Style. The reserved value
// "random" is handled by the caller (renderer), not by the registry.
type Registry interface {
	Get(name string) (Style, bool)
	List() []string
}

// Measure returns the width and height the text layer would occupy,
// without generating the SVG fragment. The renderer calls this before
// Draw so it can compute the final canvas width and pass it to Draw for
// default centering (the default text x is canvasWidth/2, not bgW/2).
func Measure(p Params) (width, height int) {
	if p.UnshowFont {
		return 0, 0
	}
	fontSize := effectiveFontSize(p.FontSize)
	charW := int(float64(fontSize) * MonoCharWidthFactor)
	if charW < 1 {
		charW = 1
	}
	return textWidth(p.Text, fontSize) + charW, fontSize + TextGapBelowImage
}

// Draw renders the counter text as layer-1. It receives the background
// dimensions (bgW, bgH) for ratio positioning and the default
// below-image Y placement, plus canvasWidth for default horizontal
// centering (the default text x is canvasWidth/2 so the text centers on
// the full merged canvas, not just the image). When UnshowFont is true
// it returns an empty Layer.
func Draw(p Params, bgW, bgH, canvasWidth int) imgcore.Layer {
	if p.UnshowFont {
		return imgcore.Layer{}
	}

	fontSize := effectiveFontSize(p.FontSize)
	charW := int(float64(fontSize) * MonoCharWidthFactor)
	if charW < 1 {
		charW = 1
	}
	textW := textWidth(p.Text, fontSize) + charW

	// Placement: pixel > ratio > default-below-center (M6).
	textX := canvasWidth / 2
	textY := bgH + fontSize
	anchor := "middle"
	if p.Position.X != 0 || p.Position.Y != 0 {
		// Pixel mode: X is the left edge, Y is the top edge.
		textX = p.Position.X
		textY = p.Position.Y + fontSize
		anchor = "start"
	} else if p.Position.RX != 0 || p.Position.RY != 0 {
		// Ratio mode: fraction of the displayed image dims.
		textX = int(float64(bgW) * p.Position.RX)
		textY = int(float64(bgH) * p.Position.RY) + fontSize
		anchor = "start"
	}

	family := p.FontStyle.Family
	if family == "" {
		family = DefaultFontFamily
	}
	color := p.FontStyle.Color
	if color == "" {
		color = DefaultFontColor
	}
	weightAttr := ""
	if p.FontStyle.Weight != "" {
		weightAttr = ` font-weight="` + imgutils.EscapeXML(p.FontStyle.Weight) + `"`
	}

	var b strings.Builder
	fmt.Fprintf(&b, `  <text x="%d" y="%d" text-anchor="%s" font-family="%s" font-size="%d" fill="%s"%s>%s</text>`+"\n",
		textX, textY, anchor, imgutils.EscapeXML(family), fontSize, imgutils.EscapeXML(color), weightAttr, imgutils.EscapeXML(p.Text))

	return imgcore.Layer{Fragment: b.String(), Width: textW, Height: fontSize + TextGapBelowImage}
}

// effectiveFontSize resolves the font size from fsize. Font sizing is
// independent of image Scale (M5.6 keeps text readable at any image size).
func effectiveFontSize(fsize int) int {
	fs := fsize
	if fs <= 0 {
		fs = DefaultFontSize
	}
	if fs < 1 {
		fs = 1
	}
	return fs
}

// textWidth estimates the pixel width of the counter text at a given font
// size so the viewBox can grow to fit it.
func textWidth(text string, fontSize int) int {
	if len(text) == 0 {
		return 0
	}
	return int(float64(len(text)*fontSize) * MonoCharWidthFactor)
}

// builtinRegistry loads font-style themes from the embedded
// assets/f-theme tree at construction time.
type builtinRegistry struct {
	styles map[string]Style
}

// NewBuiltinRegistry scans the embedded assets/f-theme directory and
// loads every *.json file as a Style.
func NewBuiltinRegistry() (Registry, []error) {
	reg := &builtinRegistry{styles: make(map[string]Style)}
	var errs []error

	root := "f-theme"
	sub, err := fs.Sub(assets.FS, root)
	if err != nil {
		return reg, []error{fmt.Errorf("fdrawer: open embedded %s: %w", root, err)}
	}
	entries, err := fs.ReadDir(sub, ".")
	if err != nil {
		return reg, []error{fmt.Errorf("fdrawer: read %s: %w", root, err)}
	}

	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".json") {
			continue
		}
		st, err := loadStyle(sub, e.Name())
		if err != nil {
			errs = append(errs, err)
			continue
		}
		reg.styles[st.Name] = st
	}
	return reg, errs
}

// loadStyle decodes one *.json file into a Style.
func loadStyle(fsys fs.FS, filename string) (Style, error) {
	raw, err := fs.ReadFile(fsys, filename)
	if err != nil {
		return Style{}, fmt.Errorf("fdrawer: read %s: %w", filename, err)
	}
	var st Style
	if err := json.Unmarshal(raw, &st); err != nil {
		return Style{}, fmt.Errorf("fdrawer: parse %s: %w", filename, err)
	}
	if st.Name == "" {
		return Style{}, fmt.Errorf("fdrawer: %s missing name", filename)
	}
	return st, nil
}

// Get returns the style for name, or false if not registered.
func (r *builtinRegistry) Get(name string) (Style, bool) {
	st, ok := r.styles[name]
	return st, ok
}

// List returns registered style names sorted for stable output.
func (r *builtinRegistry) List() []string {
	out := make([]string, 0, len(r.styles))
	for name := range r.styles {
		out = append(out, name)
	}
	sort.Strings(out)
	return out
}
