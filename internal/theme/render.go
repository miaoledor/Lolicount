package theme

import (
	"fmt"
	"math"
	"strconv"
	"strings"
)

// RenderParams controls how a number is rendered into an SVG.
// Pure-number mode only (no background); background overlay is M5.
type RenderParams struct {
	Count     int64   // value to render
	Padding   int     // zero-pad digit count, e.g. 7 -> 0000123
	Prefix    int64   // >=0 prepends these digits before the count
	Offset    float64 // extra pixels between glyphs (negative overlaps)
	Align     string  // top|center|bottom vertical alignment within the run
	Scale     float64 // relative multiplier applied after fsize normalization
	FontSize  int     // fsize: target pixel height; 0 = use glyph native height
	Pixelated string  // "1" pixelated, else smooth
	DarkMode  string  // "0"|"1"|"auto"
}

// Render produces the counter SVG for the given theme and params.
// It is the Go port of Moe-Counter's themify.js getCountImage, with the
// added fsize normalization step: final height = (fsize>0 ? fsize :
// native) * scale.
func Render(th *Theme, p RenderParams) (string, error) {
	if th == nil {
		return "", fmt.Errorf("theme: render called with nil theme")
	}
	if err := p.normalize(); err != nil {
		return "", err
	}

	run, err := buildRun(th, p)
	if err != nil {
		return "", err
	}

	maxH := 0.0
	for _, g := range run {
		if g.h > maxH {
			maxH = g.h
		}
	}

	var defs strings.Builder
	var parts strings.Builder
	x := 0.0

	// Defs: one <image> per unique glyph actually present.
	seen := map[CharName]bool{}
	for _, g := range run {
		if seen[g.slot] {
			continue
		}
		seen[g.slot] = true
		fmt.Fprintf(&defs, "\n    <image id=%q width=%q height=%q xlink:href=%q />",
			g.slot, fnum(g.h), fnum(g.w), g.data)
	}

	// Body: a <use> per glyph, positioned left-to-right.
	for _, g := range run {
		yOffset := 0.0
		switch p.Align {
		case "center":
			yOffset = (maxH - g.h) / 2
		case "bottom":
			yOffset = maxH - g.h
		}
		if yOffset != 0 {
			fmt.Fprintf(&parts, "\n    <use x=%q y=%q xlink:href=%q />", fnum(x), fnum(yOffset), g.slot)
		} else {
			fmt.Fprintf(&parts, "\n    <use x=%q xlink:href=%q />", fnum(x), g.slot)
		}
		x += g.w + p.Offset
	}
	x -= p.Offset // trailing offset was over-added

	width := x
	height := maxH
	if width < 0 {
		width = 0
	}
	if height < 0 {
		height = 0
	}

	return svgDocument(width, height, p, defs.String(), parts.String()), nil
}

// renderedGlyph is one glyph slot resolved to its scaled dimensions+data.
type renderedGlyph struct {
	slot CharName
	w    float64
	h    float64
	data string
}

// buildRun resolves the full glyph sequence (prefix + padded count +
// decorations) into scaled glyphs, skipping any decoration absent from
// the theme. A missing digit is a hard error (themes guarantee 0..9).
func buildRun(th *Theme, p RenderParams) ([]renderedGlyph, error) {
	str := strconv.FormatInt(p.Count, 10)
	if p.Padding > 0 && len(str) < p.Padding {
		str = strings.Repeat("0", p.Padding-len(str)) + str
	}
	if p.Prefix >= 0 {
		str = strconv.FormatInt(p.Prefix, 10) + str
	}

	slots := make([]CharName, 0, len(str)+2)
	if c, ok := th.Lookup("_start"); ok {
		slots = append(slots, "_start")
		_ = c
	}
	for _, r := range str {
		slots = append(slots, CharName(r))
	}
	if _, ok := th.Lookup("_end"); ok {
		slots = append(slots, "_end")
	}

	out := make([]renderedGlyph, 0, len(slots))
	for _, s := range slots {
		c, ok := th.Lookup(s)
		if !ok {
			return nil, fmt.Errorf("theme %s: missing glyph %q", th.Name, s)
		}
		w, h := scaledSize(c, p)
		out = append(out, renderedGlyph{slot: s, w: w, h: h, data: c.Data})
	}
	return out, nil
}

// scaledSize computes the final w/h for a glyph.
// final height = (FontSize>0 ? FontSize : nativeHeight) * Scale
// width scales proportionally so aspect ratio is preserved.
func scaledSize(c ThemeChar, p RenderParams) (w, h float64) {
	nativeW := float64(c.Width)
	nativeH := float64(c.Height)
	targetH := nativeH
	if p.FontSize > 0 {
		targetH = float64(p.FontSize)
	}
	targetH *= p.Scale
	if nativeH == 0 {
		return nativeW * p.Scale, targetH
	}
	w = nativeW * (targetH / nativeH)
	return w, targetH
}

// normalize clamps/defaults mutable params.
func (p *RenderParams) normalize() error {
	if p.Scale <= 0 {
		p.Scale = 1
	}
	if p.Padding < 0 {
		p.Padding = 0
	}
	if p.Offset == 0 {
		p.Offset = 0
	}
	switch p.Align {
	case "", "top":
		p.Align = "top"
	case "center", "bottom":
	default:
		return fmt.Errorf("theme: invalid align %q", p.Align)
	}
	switch p.DarkMode {
	case "", "0", "1", "auto":
	default:
		return fmt.Errorf("theme: invalid darkmode %q", p.DarkMode)
	}
	if p.Pixelated == "" {
		p.Pixelated = "1"
	}
	return nil
}

// svgDocument assembles the final SVG string with style for pixelation
// and dark-mode brightness.
func svgDocument(w, h float64, p RenderParams, defs, parts string) string {
	var style strings.Builder
	style.WriteString("\n  svg {\n    ")
	if p.Pixelated == "1" {
		style.WriteString("image-rendering: pixelated;\n    ")
	}
	if p.DarkMode == "1" {
		style.WriteString("filter: brightness(.6);\n    ")
	}
	style.WriteString("\n  }")
	if p.DarkMode == "auto" {
		style.WriteString("\n  @media (prefers-color-scheme: dark) { svg { filter: brightness(.6); } }")
	}

	return fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
<svg viewBox="0 0 %s %s" width="%s" height="%s" version="1.1" xmlns="http://www.w3.org/2000/svg" xmlns:xlink="http://www.w3.org/1999/xlink">
  <title>Lolicount</title>
  <style>%s
  </style>
  <defs>%s
  </defs>
  <g>%s
  </g>
</svg>
`, fnum(w), fnum(h), fnum(w), fnum(h), style.String(), defs, parts)
}

// fnum formats a float for SVG attributes: up to 5 decimals, trailing
// zeros trimmed, matching Moe-Counter's toFixed(x, 5) output shape.
func fnum(v float64) string {
	if math.IsNaN(v) || math.IsInf(v, 0) {
		return "0"
	}
	s := strconv.FormatFloat(v, 'f', 5, 64)
	s = strings.TrimRight(s, "0")
	s = strings.TrimRight(s, ".")
	if s == "" || s == "-" {
		return "0"
	}
	return s
}
