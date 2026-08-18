package theme

import (
	"fmt"
	"strconv"
	"strings"
)

// RenderParams controls how the background frame + counter text are
// composed. Per M5.5 the theme frame is the sole background (layer 0)
// and the count is rendered as <text> (layer 1). There is no separate
// "bg" concept: theme IS the background.
type RenderParams struct {
	// FrameIndex selects which theme frame to draw as the background.
	FrameIndex int
	// Count is the numeric value to draw as text on top of the frame.
	Count int64
	// Number, when >= 0, overrides the displayed counter text with this
	// fixed value (preview mode, like Moe-Counter's num).
	Number int64
	// FontSize controls the counter text size in pixels. 0 = default 16.
	FontSize int
	// Scale multiplies the font size. 0 = 1.
	Scale float64
	// X is the absolute left origin of the text block (0 = centered).
	X int
	// Y is the absolute top origin of the text block (0 = vertically
	// centered on the frame).
	Y int
}

// Render composes an SVG: the theme frame as the background image
// (layer 0, data URI per Iron Rule 2) with the counter value overlaid
// as <text> (layer 1). This is the single render path — theme and bg
// are the same thing.
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

	return composeSVG(frame, text, p), nil
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

// composeSVG builds the final SVG document. The theme frame is the
// background image (layer 0); the counter text is overlaid (layer 1).
// When X/Y are 0 the text is centered on the frame; otherwise the text
// block is positioned absolutely and the viewBox stays the frame size.
func composeSVG(frame Frame, text string, p RenderParams) string {
	fontSize := effectiveFontSize(p.FontSize, p.Scale)
	charW := int(float64(fontSize) * monoCharWidthFactor)
	if charW < 1 {
		charW = 1
	}

	absolute := p.X != 0 || p.Y != 0

	canvasWidth := frame.Width
	if tw := textWidth(text, fontSize) + charW; tw > canvasWidth {
		canvasWidth = tw
	}
	canvasHeight := frame.Height

	frameX := (canvasWidth - frame.Width) / 2

	var textX, textY int
	if absolute {
		// Absolute positioning: text origin is (X, Y+fontSize) so Y is
		// the top edge of the text block. viewBox stays frame size.
		textX = p.X
		textY = p.Y + fontSize
		canvasWidth = frame.Width
		canvasHeight = frame.Height
	} else {
		textX = canvasWidth / 2
		textY = frame.Height/2 + fontSize/3
	}

	var b strings.Builder
	fmt.Fprintf(&b, `<?xml version="1.0" encoding="UTF-8"?>`+"\n")
	fmt.Fprintf(&b, `<svg viewBox="0 0 %d %d" width="%d" height="%d" version="1.1" xmlns="http://www.w3.org/2000/svg" xmlns:xlink="http://www.w3.org/1999/xlink">`+"\n",
		canvasWidth, canvasHeight, canvasWidth, canvasHeight)
	b.WriteString("  <title>Lolicount</title>\n")
	// Layer 0: theme frame as the background image (M5.5).
	fmt.Fprintf(&b, `  <image x="%d" y="0" width="%d" height="%d" xlink:href="%s" />`+"\n",
		frameX, frame.Width, frame.Height, frame.Data)
	// Layer 1: counter text overlaid on top of the background (M5.5).
	anchor := "middle"
	if absolute {
		anchor = "start"
	}
	fmt.Fprintf(&b, `  <text x="%d" y="%d" text-anchor="%s" font-family="monospace" font-size="%d" fill="#333">%s</text>`+"\n",
		textX, textY, anchor, fontSize, escapeXML(text))
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
