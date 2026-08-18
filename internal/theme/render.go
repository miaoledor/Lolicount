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
	// Count is the numeric value to draw as text below the frame.
	Count int64
	// Number, when >= 0, overrides the displayed counter text with this
	// fixed value (preview mode, like Moe-Counter's num).
	Number int64
	// FontSize controls the counter text size in pixels. 0 = default 16.
	FontSize int
	// Scale controls the image display size as a multiplier of the
	// uniform base size (defaultDisplaySize). 0 = use the base size
	// alone, so every theme renders at a consistent display size
	// regardless of its source resolution (M5.6). The aspect ratio is
	// always preserved — the image is scaled, never stretched.
	Scale float64
	// UnshowFont, when true, omits the counter <text> entirely (M5.6
	// ?unshowf=true).
	UnshowFont bool
}

// defaultDisplaySize is the longest-edge target (in px) every frame is
// scaled down to when no explicit Scale is given (M5.6: all images show
// at a consistent size). Frames smaller than this are scaled up to it
// as well so output is uniform across themes.
const defaultDisplaySize = 400

// Render composes an SVG: the theme frame as the background image
// (layer 0, data URI per Iron Rule 2) scaled to a uniform display size,
// with the counter value rendered as <text> below it (layer 1, M5.6).
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

// displaySize returns the target longest-edge display size for the
// frame. When Scale is 0 the uniform base size is used (M5.6); otherwise
// the base is multiplied by Scale.
func displaySize(scale float64) int {
	if scale <= 0 {
		return defaultDisplaySize
	}
	return int(float64(defaultDisplaySize) * scale)
}

// scaledFrameDims computes the displayed width/height of the frame,
// preserving aspect ratio so the longest edge equals the display size.
// Scaling (not stretching) keeps the image undistorted (M5.6).
func scaledFrameDims(frame Frame, display int) (int, int) {
	if frame.Width <= 0 || frame.Height <= 0 || display <= 0 {
		return frame.Width, frame.Height
	}
	longest := frame.Width
	if frame.Height > longest {
		longest = frame.Height
	}
	if longest == 0 {
		return frame.Width, frame.Height
	}
	ratio := float64(display) / float64(longest)
	w := int(float64(frame.Width) * ratio)
	h := int(float64(frame.Height) * ratio)
	if w < 1 {
		w = 1
	}
	if h < 1 {
		h = 1
	}
	return w, h
}

// effectiveFontSize resolves the font size from FontSize. Font sizing is
// independent of image Scale: fsize is absolute pixels (M5.6 keeps text
// readable at any image size).
func effectiveFontSize(fsize int) int {
	fs := fsize
	if fs <= 0 {
		fs = defaultFontSize
	}
	if fs < 1 {
		fs = 1
	}
	return fs
}

// textWidth estimates the pixel width of the counter text at a given font
// size so the viewBox can grow to fit it when the text is wider than the
// displayed image.
func textWidth(text string, fontSize int) int {
	if len(text) == 0 {
		return 0
	}
	return int(float64(len(text)*fontSize) * monoCharWidthFactor)
}

// composeSVG builds the final SVG document. The theme frame is the
// background image (layer 0), scaled to a uniform display size with its
// aspect ratio preserved. The counter text sits below the image,
// horizontally centered (M5.6). When UnshowFont is true the text is
// omitted entirely.
func composeSVG(frame Frame, text string, p RenderParams) string {
	display := displaySize(p.Scale)
	imgW, imgH := scaledFrameDims(frame, display)
	fontSize := effectiveFontSize(p.FontSize)
	charW := int(float64(fontSize) * monoCharWidthFactor)
	if charW < 1 {
		charW = 1
	}

	// Canvas: image width (or text width if wider) x image height + text.
	textW := textWidth(text, fontSize) + charW
	canvasWidth := imgW
	if textW > canvasWidth {
		canvasWidth = textW
	}
	canvasHeight := imgH
	if !p.UnshowFont {
		canvasHeight = imgH + fontSize + 4
	}

	imgX := (canvasWidth - imgW) / 2
	if imgX < 0 {
		imgX = 0
	}

	var b strings.Builder
	fmt.Fprintf(&b, `<?xml version="1.0" encoding="UTF-8"?>`+"\n")
	fmt.Fprintf(&b, `<svg viewBox="0 0 %d %d" width="%d" height="%d" version="1.1" xmlns="http://www.w3.org/2000/svg" xmlns:xlink="http://www.w3.org/1999/xlink">`+"\n",
		canvasWidth, canvasHeight, canvasWidth, canvasHeight)
	b.WriteString("  <title>Lolicount</title>\n")
	// Layer 0: theme frame as the background image (M5.5), scaled to a
	// uniform display size, aspect ratio preserved (M5.6).
	fmt.Fprintf(&b, `  <image x="%d" y="0" width="%d" height="%d" xlink:href="%s" />`+"\n",
		imgX, imgW, imgH, frame.Data)
	// Layer 1: counter text below the image, centered (M5.6). Omitted
	// when UnshowFont is set (?unshowf=true).
	if !p.UnshowFont {
		textX := canvasWidth / 2
		textY := imgH + fontSize
		fmt.Fprintf(&b, `  <text x="%d" y="%d" text-anchor="middle" font-family="monospace" font-size="%d" fill="#333">%s</text>`+"\n",
			textX, textY, fontSize, escapeXML(text))
	}
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
