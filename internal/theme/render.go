package theme

import (
	"fmt"
	"strconv"
	"strings"
)

// RenderParams controls how the background frame + counter text are
// composed. Per M5.5 the theme frame is a pure style background (layer 0)
// and the count is rendered as <text> (layer 1).
type RenderParams struct {
	// FrameIndex selects which theme frame to draw as the style background.
	// Per M5.5 the theme never reflects the count, so the handler always
	// passes 0. Kept for forward-compat with multi-background themes.
	FrameIndex int
	// Count is the numeric value to draw as text on top of the background.
	Count int64
	// Number, when >= 0, overrides the displayed counter text with this
	// fixed value (preview mode, like Moe-Counter's num).
	Number int64
	// FontSize controls the counter text size in pixels. 0 = default 16.
	FontSize int
	// Scale multiplies the font size. 0 = 1.
	Scale float64
}

// Render composes an SVG for the pure-digit mode (no bg param). Per M5.5
// the theme frame is the background (layer 0) and the count text is
// overlaid on top (layer 1). The frame image is embedded as a data URI
// (AGENTS.md Iron Rule 2: theme images use data URIs).
func Render(th *Theme, p RenderParams) (string, error) {
	if th == nil {
		return "", fmt.Errorf("theme: render called with nil theme")
	}
	if th.Size() == 0 {
		return "", fmt.Errorf("theme %s: no frames", th.Name)
	}

	frame, ok := th.Frame(p.FrameIndex)
	if !ok {
		return "", fmt.Errorf("theme %s: frame index %d out of range (size %d)", th.Name, p.FrameIndex, th.Size())
	}

	text := strconv.FormatInt(p.Count, 10)
	if p.Number > 0 {
		text = strconv.FormatInt(p.Number, 10)
	}

	return composeSVG(frame, text, p.FontSize, p.Scale), nil
}

const defaultFontSize = 16

// monoCharWidthFactor approximates the advance width of one monospace
// digit relative to font-size: monospace digits are ~0.6em wide.
const monoCharWidthFactor = 0.6

// effectiveFontSize resolves the font size from FontSize and Scale.
func effectiveFontSize(fsize int, scale float64) int {
	if scale == 0 {
		scale = 1
	}
	fs := fsize
	if fs <= 0 {
		fs = defaultFontSize
	}
	fs = int(float64(fs) * scale)
	if fs < 1 {
		fs = 1
	}
	return fs
}

// textWidth estimates the pixel width of the counter text at a given font
// size so the viewBox can grow to fit it when the text is wider than the
// frame.
func textWidth(text string, fontSize int) int {
	if len(text) == 0 {
		return 0
	}
	return int(float64(len(text)*fontSize) * monoCharWidthFactor)
}

// composeSVG builds the final SVG document. Per M5.5 the theme frame is
// the background image (layer 0, bottom) and the counter text is overlaid
// on top (layer 1). The viewBox is the frame dimensions; the text is
// centered on the frame. If the text is wider than the frame, the canvas
// widens (with the frame centered) so the count never overflows.
func composeSVG(frame Frame, text string, fsize int, scale float64) string {
	fontSize := effectiveFontSize(fsize, scale)
	charW := int(float64(fontSize) * monoCharWidthFactor)
	if charW < 1 {
		charW = 1
	}

	canvasWidth := frame.Width
	if tw := textWidth(text, fontSize) + charW; tw > canvasWidth {
		canvasWidth = tw
	}
	canvasHeight := frame.Height

	frameX := (canvasWidth - frame.Width) / 2
	cx := canvasWidth / 2
	// Vertically center the text baseline on the frame.
	textY := frame.Height/2 + fontSize/3

	var b strings.Builder
	fmt.Fprintf(&b, `<?xml version="1.0" encoding="UTF-8"?>`+"\n")
	fmt.Fprintf(&b, `<svg viewBox="0 0 %d %d" width="%d" height="%d" version="1.1" xmlns="http://www.w3.org/2000/svg" xmlns:xlink="http://www.w3.org/1999/xlink">`+"\n",
		canvasWidth, canvasHeight, canvasWidth, canvasHeight)
	b.WriteString("  <title>Lolicount</title>\n")
	// Layer 0: theme frame as the background image (M5.5).
	fmt.Fprintf(&b, `  <image x="%d" y="0" width="%d" height="%d" xlink:href="%s" />`+"\n",
		frameX, frame.Width, frame.Height, frame.Data)
	// Layer 1: counter text overlaid on top of the background (M5.5).
	fmt.Fprintf(&b, `  <text x="%d" y="%d" text-anchor="middle" font-family="monospace" font-size="%d" fill="#333">%s</text>`+"\n",
		cx, textY, fontSize, escapeXML(text))
	b.WriteString("</svg>\n")
	return b.String()
}

// escapeXML escapes the five special XML characters in a text node.
func escapeXML(s string) string {
	r := strings.NewReplacer(
		"&", "&amp;",
		"<", "&lt;",
		">", "&gt;",
		`"`, "&quot;",
		"'", "&apos;",
	)
	return r.Replace(s)
}

// BgParams describes the background layer for the overlay render mode.
// It mirrors bg.Background but lives in the theme package so theme does
// NOT import bg (AGENTS.md dependency direction: theme and bg are peer
// packages, both under server; theme must not depend on bg).
type BgParams struct {
	URL    string // external CDN URL, referenced as <image href> (Iron Rule 2)
	Width  int
	Height int
}

// OverlayParams controls how the counter text is placed on top of a
// background image. X/Y are the top-left origin of the text block
// (absolute, relative to the background's viewBox). Align controls the
// vertical alignment of the text baseline WITHIN the text block.
type OverlayParams struct {
	X     int
	Y     int
	Align string  // "top" | "center" | "bottom"
	FSize int     // absolute pixel height of the text; 0 = default 16
	Scale float64 // relative multiplier; defaults to 1
}

// RenderWithBg composes the overlay-mode SVG: the background image as
// the base layer (external URL, NOT embedded — Iron Rule 2) with the
// counter text drawn on top as <text> at (X, Y).
//
// Per M5.5 the count is ALWAYS rendered as text (layer 1), never as an
// image. The theme frame is NOT used here — the background comes from the
// bg param. The viewBox is fixed to the background dimensions; the text
// block is positioned absolutely via X/Y.
func RenderWithBg(th *Theme, bg BgParams, o OverlayParams, p RenderParams) (string, error) {
	if th == nil {
		return "", fmt.Errorf("theme: RenderWithBg called with nil theme")
	}
	// th is required for consistency but the frame is not drawn in bg mode
	// (the background image is the bg param, not a theme frame). We still
	// validate the theme has frames so a misconfigured theme errors early.
	if th.Size() == 0 {
		return "", fmt.Errorf("theme %s: no frames", th.Name)
	}
	if bg.URL == "" || bg.Width <= 0 || bg.Height <= 0 {
		return "", fmt.Errorf("theme: invalid background params")
	}

	text := strconv.FormatInt(p.Count, 10)
	if p.Number > 0 {
		text = strconv.FormatInt(p.Number, 10)
	}

	fontSize := effectiveFontSize(o.FSize, o.Scale)
	return composeOverlaySVG(bg, text, o, fontSize), nil
}

// composeOverlaySVG builds the overlay-mode SVG. The background is an
// external URL (Iron Rule 2); the counter is <text> overlaid at (X, Y).
func composeOverlaySVG(bg BgParams, text string, o OverlayParams, fontSize int) string {
	var b strings.Builder
	fmt.Fprintf(&b, `<?xml version="1.0" encoding="UTF-8"?>`+"\n")
	fmt.Fprintf(&b, `<svg viewBox="0 0 %d %d" width="%d" height="%d" version="1.1" xmlns="http://www.w3.org/2000/svg" xmlns:xlink="http://www.w3.org/1999/xlink">`+"\n",
		bg.Width, bg.Height, bg.Width, bg.Height)
	b.WriteString("  <title>Lolicount</title>\n")
	// Layer 0: background (external URL, never embedded — Iron Rule 2).
	fmt.Fprintf(&b, `  <image x="0" y="0" width="%d" height="%d" xlink:href="%s" />`+"\n",
		bg.Width, bg.Height, escapeXML(bg.URL))
	// Layer 1: counter text overlaid on top (M5.5: count is a font, not an image).
	// text-anchor=start so X is the left edge of the text block.
	textY := o.Y + fontSize
	fmt.Fprintf(&b, `  <text x="%d" y="%d" font-family="monospace" font-size="%d" fill="#333">%s</text>`+"\n",
		o.X, textY, fontSize, escapeXML(text))
	b.WriteString("</svg>\n")
	return b.String()
}
