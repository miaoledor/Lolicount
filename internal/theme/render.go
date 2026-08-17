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

// Render composes an SVG: the selected frame image as the base layer,
// with the count text drawn centered horizontally and directly below the
// image. The viewBox is frame width x (frame height + text band).
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

// textBandHeight is the vertical space reserved below the image for the
// counter text. Tuned for a readable default font size.
const (
	textBandHeight = 24
	fontSize       = 16
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

// composeSVG builds the final SVG document. The viewBox width is the
// larger of the frame width and the estimated text width (plus padding),
// so a long counter value never overflows the canvas. The frame image is
// centered horizontally; the text is centered under it.
func composeSVG(frame Frame, text string) string {
	totalHeight := frame.Height + textBandHeight
	// Pad the text so it is not flush against the viewBox edge.
	canvasWidth := frame.Width
	if tw := textWidth(text) + monoCharWidth; tw > canvasWidth {
		canvasWidth = tw
	}

	frameX := (canvasWidth - frame.Width) / 2
	cx := canvasWidth / 2
	textY := frame.Height + 16

	var b strings.Builder
	fmt.Fprintf(&b, `<?xml version="1.0" encoding="UTF-8"?>`+"\n")
	fmt.Fprintf(&b, `<svg viewBox="0 0 %d %d" width="%d" height="%d" version="1.1" xmlns="http://www.w3.org/2000/svg" xmlns:xlink="http://www.w3.org/1999/xlink">`+"\n",
		canvasWidth, totalHeight, canvasWidth, totalHeight)
	b.WriteString("  <title>Lolicount</title>\n")
	// Frame image, horizontally centered.
	fmt.Fprintf(&b, `  <image x="%d" y="0" width="%d" height="%d" xlink:href="%s" />`+"\n",
		frameX, frame.Width, frame.Height, frame.Data)
	// Counter text: centered horizontally, baseline in the text band.
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
