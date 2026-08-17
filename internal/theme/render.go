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
const textBandHeight = 24

// composeSVG builds the final SVG document.
func composeSVG(frame Frame, text string) string {
	totalHeight := frame.Height + textBandHeight
	var b strings.Builder
	fmt.Fprintf(&b, `<?xml version="1.0" encoding="UTF-8"?>`+"\n")
	fmt.Fprintf(&b, `<svg viewBox="0 0 %d %d" width="%d" height="%d" version="1.1" xmlns="http://www.w3.org/2000/svg" xmlns:xlink="http://www.w3.org/1999/xlink">`+"\n",
		frame.Width, totalHeight, frame.Width, totalHeight)
	b.WriteString("  <title>Lolicount</title>\n")
	// Frame image as the base layer.
	fmt.Fprintf(&b, `  <image x="0" y="0" width="%d" height="%d" xlink:href="%s" />`+"\n",
		frame.Width, frame.Height, frame.Data)
	// Counter text: centered horizontally, baseline in the text band.
	cx := frame.Width / 2
	textY := frame.Height + 16
	fmt.Fprintf(&b, `  <text x="%d" y="%d" text-anchor="middle" font-family="monospace" font-size="16" fill="#333">%s</text>`+"\n",
		cx, textY, escapeXML(text))
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
