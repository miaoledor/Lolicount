package theme

import (
	"fmt"
	"strings"
)

// composeCharacterSVG renders an assembled character portrait as the
// background (layer 0) with the counter text below it (layer 1).
//
// The portrait layers are placed at their absolute left/top within the
// PSD canvas (CharacterCanvasW x CharacterCanvasH). The whole canvas is
// then scaled to the uniform display size (longest edge, M5.6) so
// character themes render at a consistent size alongside frame themes.
// Each layer is a data URI <image> (AGENTS.md Iron Rule 2: theme images
// use data URIs).
func composeCharacterSVG(p *ComposedPortrait, text string, params RenderParams) string {
	display := displaySize(params.Scale)
	imgW, imgH := scaledCanvasDims(CharacterCanvasW, CharacterCanvasH, display)
	fontSize := effectiveFontSize(params.FontSize)
	charW := int(float64(fontSize) * MonoCharWidthFactor)
	if charW < 1 {
		charW = 1
	}

	textW := textWidth(text, fontSize) + charW
	canvasWidth := imgW
	if textW > canvasWidth {
		canvasWidth = textW
	}
	canvasHeight := imgH
	if !params.UnshowFont {
		canvasHeight = imgH + fontSize + TextGapBelowImage
	}

	scaleX := float64(imgW) / float64(CharacterCanvasW)
	scaleY := float64(imgH) / float64(CharacterCanvasH)
	imgX := (canvasWidth - imgW) / 2
	if imgX < 0 {
		imgX = 0
	}

	var b strings.Builder
	fmt.Fprintf(&b, `<?xml version="1.0" encoding="UTF-8"?>`+"\n")
	fmt.Fprintf(&b, `<svg viewBox="0 0 %d %d" width="%d" height="%d" version="1.1" xmlns="http://www.w3.org/2000/svg" xmlns:xlink="http://www.w3.org/1999/xlink">`+"\n",
		canvasWidth, canvasHeight, canvasWidth, canvasHeight)
	b.WriteString("  <title>Lolicount</title>\n")
	// Layer 0: each portrait part, positioned by its scaled absolute
	// coordinates. z-order is the Parts slice order (body → eye → brow
	// → mouth → face), so later parts paint on top.
	for _, part := range p.Parts {
		x := imgX + int(float64(part.Left)*scaleX)
		y := int(float64(part.Top)*scaleY)
		w := int(float64(part.Width) * scaleX)
		h := int(float64(part.Height) * scaleY)
		if w < 1 {
			w = 1
		}
		if h < 1 {
			h = 1
		}
		fmt.Fprintf(&b, `  <image x="%d" y="%d" width="%d" height="%d" xlink:href="%s" />`+"\n",
			x, y, w, h, part.Data)
	}
	// Layer 1: counter text below the image, centered (M5.6).
	if !params.UnshowFont {
		textX := canvasWidth / 2
		textY := imgH + fontSize
		anchor := "middle"
		if params.Position.X != 0 || params.Position.Y != 0 {
			textX = params.Position.X
			textY = params.Position.Y + fontSize
			anchor = "start"
		} else if params.Position.RX != 0 || params.Position.RY != 0 {
			textX = int(float64(imgW) * params.Position.RX)
			textY = int(float64(imgH) * params.Position.RY) + fontSize
			anchor = "start"
		}
		family := params.FontStyle.Family
		if family == "" {
			family = DefaultFontFamily
		}
		color := params.FontStyle.Color
		if color == "" {
			color = DefaultFontColor
		}
		weightAttr := ""
		if params.FontStyle.Weight != "" {
			weightAttr = ` font-weight="` + escapeXML(params.FontStyle.Weight) + `"`
		}
		fmt.Fprintf(&b, `  <text x="%d" y="%d" text-anchor="%s" font-family="%s" font-size="%d" fill="%s"%s>%s</text>`+"\n",
			textX, textY, anchor, escapeXML(family), fontSize, escapeXML(color), weightAttr, escapeXML(text))
	}
	b.WriteString("</svg>\n")
	return b.String()
}

// scaledCanvasDims computes the displayed width/height of a canvas with
// the same longest-edge rule as scaledFrameDims, for the character
// canvas rather than a single frame.
func scaledCanvasDims(canvasW, canvasH, display int) (int, int) {
	if canvasW <= 0 || canvasH <= 0 || display <= 0 {
		return canvasW, canvasH
	}
	longest := canvasW
	if canvasH > longest {
		longest = canvasH
	}
	ratio := float64(display) / float64(longest)
	w := int(float64(canvasW) * ratio)
	h := int(float64(canvasH) * ratio)
	if w < 1 {
		w = 1
	}
	if h < 1 {
		h = 1
	}
	return w, h
}
