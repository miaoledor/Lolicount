package theme

import (
	"fmt"
	"strconv"
	"strings"
)

// RenderParams controls how a frame + counter text are composed.
// M2.5 model: one frame image with the count drawn as text below it.
type RenderParams struct {
	// FrameIndex selects which frame of the theme to draw. The handler
	// computes (count+1) % Size by default; number param overrides.
	FrameIndex int
	// Count is the numeric value to draw as text below the frame.
	Count int64
	// Number, when >= 0, overrides the displayed counter text with this
	// fixed value (preview mode, like Moe-Counter's num).
	Number int64
}

// Render composes an SVG for the pure-digit mode (no bg param). Per M5.5
// the theme frame is the background (layer 0) and the count text is
// overlaid on top (layer 1); the viewBox is the frame dimensions.
//
// The frame image is embedded as a data URI (AGENTS.md Iron Rule 2:
// digit/counter images use data URIs, not external URLs).
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

	return composeSVG(frame, text), nil
}

const (
	fontSize = 16
	// monoCharWidth approximates the advance width of one monospace glyph
	// at fontSize 16. monospace digits are ~0.6em wide, so ~9.6px; 10 is a
	// safe whole-number upper bound to avoid clipping the last digit.
	monoCharWidth = 10
)

// textWidth estimates the pixel width of the counter text so the viewBox
// can grow to fit it when the text is wider than the frame.
func textWidth(text string) int {
	if len(text) == 0 {
		return 0
	}
	return len(text) * monoCharWidth
}

// composeSVG builds the final SVG document. Per M5.5 the theme frame is
// the background image (layer 0, bottom) and the counter text is overlaid
// on top of it (layer 1). The viewBox is the frame dimensions; the text is
// centered on the frame. If the text is wider than the frame, the canvas
// widens (with the frame centered) so the count never overflows.
func composeSVG(frame Frame, text string) string {
	canvasWidth := frame.Width
	if tw := textWidth(text) + monoCharWidth; tw > canvasWidth {
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

// OverlayParams controls how the counter digits are placed on top of a
// background image. X/Y are the top-left origin of the digit block
// (absolute, relative to the background's viewBox). Align controls the
// vertical alignment of each digit glyph WITHIN the digit block, not
// the block's position on the background (AGENTS.md Rendering).
type OverlayParams struct {
	X     int
	Y     int
	Align string  // "top" | "center" | "bottom"
	FSize int     // absolute pixel height of each digit; 0 = glyph native height
	Scale float64 // relative multiplier; defaults to 1
}

// RenderWithBg composes the overlay-mode SVG: the background image as
// the base layer (external URL, NOT embedded — Iron Rule 2) with the
// counter digits drawn on top as data-URI images at (X, Y).
//
// The viewBox is fixed to the background dimensions; the digit block is
// positioned absolutely via X/Y. Each digit glyph from the theme is
// rendered at height = (FSize>0 ? FSize : native) * Scale, stacked
// horizontally from X.
func RenderWithBg(th *Theme, bg BgParams, o OverlayParams, p RenderParams) (string, error) {
	if th == nil {
		return "", fmt.Errorf("theme: RenderWithBg called with nil theme")
	}
	if th.Size() == 0 {
		return "", fmt.Errorf("theme %s: no frames", th.Name)
	}
	if bg.URL == "" || bg.Width <= 0 || bg.Height <= 0 {
		return "", fmt.Errorf("theme: invalid background params")
	}

	frame, ok := th.Frame(p.FrameIndex)
	if !ok {
		return "", fmt.Errorf("theme %s: frame index %d out of range", th.Name, p.FrameIndex)
	}

	text := strconv.FormatInt(p.Count, 10)
	if p.Number > 0 {
		text = strconv.FormatInt(p.Number, 10)
	}

	if o.Scale == 0 {
		o.Scale = 1
	}
	digitH := frame.Height
	if o.FSize > 0 {
		digitH = o.FSize
	}
	digitH = int(float64(digitH) * o.Scale)
	if digitH <= 0 {
		digitH = 1
	}
	// Derive digitW from the scaled height, preserving the frame's
	// aspect ratio. Clamp to 1 so a very thin glyph never vanishes.
	digitW := int(float64(frame.Width) * float64(digitH) / float64(frame.Height))
	if digitW <= 0 {
		digitW = 1
	}

	return composeOverlaySVG(bg, frame, text, o, digitW, digitH), nil
}

// composeOverlaySVG builds the overlay-mode SVG. The background is an
// external URL (Iron Rule 2); digits are data-URI images stacked from X.
// Align adjusts each glyph's Y within the digit block so mixed-height
// digits line up (AGENTS.md: align is intra-block, not block-vs-bg).
func composeOverlaySVG(bg BgParams, frame Frame, text string, o OverlayParams, digitW, digitH int) string {
	var b strings.Builder
	fmt.Fprintf(&b, `<?xml version="1.0" encoding="UTF-8"?>`+"\n")
	fmt.Fprintf(&b, `<svg viewBox="0 0 %d %d" width="%d" height="%d" version="1.1" xmlns="http://www.w3.org/2000/svg" xmlns:xlink="http://www.w3.org/1999/xlink">`+"\n",
		bg.Width, bg.Height, bg.Width, bg.Height)
	b.WriteString("  <title>Lolicount</title>\n")
	// Background: external URL, never embedded (Iron Rule 2).
	fmt.Fprintf(&b, `  <image x="0" y="0" width="%d" height="%d" xlink:href="%s" />`+"\n",
		bg.Width, bg.Height, escapeXML(bg.URL))
	// Digit glyphs: data-URI images, horizontally stacked from X.
	// In the M2.5 single-frame model every digit glyph shares the same
	// height, so align (intra-block vertical alignment) has no visible
	// effect; the digit block origin is fully determined by X/Y. align is
	// accepted for API compatibility and forward-compat with per-digit
	// themes (AGENTS.md Rendering).
	for i := range text {
		dx := o.X + i*digitW
		fmt.Fprintf(&b, `  <image x="%d" y="%d" width="%d" height="%d" xlink:href="%s" />`+"\n",
			dx, o.Y, digitW, digitH, frame.Data)
	}
	b.WriteString("</svg>\n")
	return b.String()
}
