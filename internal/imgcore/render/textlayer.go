package render

import (
	"fmt"
	"strings"

	"github.com/miaoledor/lolicount/internal/imgcore"
	"github.com/miaoledor/lolicount/internal/imgcore/imgutils"
	"github.com/miaoledor/lolicount/internal/imgcore/theme"
)

// TextLayer renders counter text or a static label as a <text> element.
// When IsCounter is true, the rendered text comes from RenderCtx.CountText
// (the live counter value). The layer supports pixel/ratio positioning,
// rotation, and font-style theming. Migrated from fdrawer.Draw.
type TextLayer struct {
	Text       string // static text (ignored when IsCounter is true)
	IsCounter  bool   // bind to RenderCtx.CountText at render time
	FontSize   int    // 0 = default
	UnshowFont bool   // omit the text entirely
	Style      theme.TextStyle
	Position   theme.TextPos
	Transform  imgcore.Transform
	Z          int
	IsFixed    bool
}

// Kind returns LayerText.
func (l *TextLayer) Kind() imgcore.LayerKind { return imgcore.LayerText }

// ZIndex returns the stack order.
func (l *TextLayer) ZIndex() int { return l.Z }

// Fixed reports whether the layer cannot be deleted.
func (l *TextLayer) Fixed() bool { return l.IsFixed }

// Render produces the SVG <text> fragment. When UnshowFont is true,
// returns an empty LayerOutput. Position resolution: pixel (X/Y) >
// ratio (RX/RY) > default (below image, centered).
//
// The default Y placement assumes the canvas height includes space for
// the text (canvasH = imageH + textH). The text baseline is placed at
// imageH + fontSize, which equals canvasH - TextGapBelowImage. This
// keeps the text visible within the viewBox.
func (l *TextLayer) Render(ctx imgcore.RenderCtx) imgcore.LayerOutput {
	if l.UnshowFont {
		return imgcore.LayerOutput{}
	}

	fontSize := effectiveFontSize(l.FontSize)
	charW := int(float64(fontSize) * theme.MonoCharWidthFactor)
	if charW < 1 {
		charW = 1
	}

	displayText := l.Text
	if l.IsCounter {
		displayText = ctx.CountText
	}
	textW := textWidth(displayText, fontSize) + charW
	textH := fontSize + theme.TextGapBelowImage

	// Placement: pixel > ratio > default-below-center.
	// Default: text sits below the image, centered horizontally. The
	// image occupies canvasH - textH pixels; the text baseline is at
	// imageBottom + fontSize = (canvasH - textH) + fontSize.
	textX := ctx.CanvasW / 2
	textY := ctx.CanvasH - textH + fontSize
	anchor := "middle"
	if l.Position.X != 0 || l.Position.Y != 0 {
		textX = l.Position.X
		textY = l.Position.Y + fontSize
		anchor = "start"
	} else if l.Position.RX != 0 || l.Position.RY != 0 {
		// Ratio mode: fraction of the background (image) dims, not the
		// full canvas. This matches the original fdrawer.Draw which
		// received bgW/bgH separately from canvasWidth. Fall back to
		// canvas dims when BgW/BgH are unset (e.g. tests).
		bgW := ctx.BgW
		if bgW == 0 {
			bgW = ctx.CanvasW
		}
		bgH := ctx.BgH
		if bgH == 0 {
			bgH = ctx.CanvasH
		}
		textX = int(float64(bgW) * l.Position.RX)
		textY = int(float64(bgH) * l.Position.RY) + fontSize
		anchor = "start"
	}

	rotation := ctx.PRNG.FloatRange(l.Transform.Rotation)

	family := l.Style.Family
	if family == "" {
		family = theme.DefaultFontFamily
	}
	color := l.Style.Color
	if color == "" {
		color = theme.DefaultFontColor
	}
	weightAttr := ""
	if l.Style.Weight != "" {
		weightAttr = ` font-weight="` + imgutils.EscapeXML(l.Style.Weight) + `"`
	}

	transformAttr := ""
	if rotation != 0 {
		transformAttr = fmt.Sprintf(` transform="rotate(%s %d %d)"`, formatFloat(rotation), textX, textY)
	}

	var b strings.Builder
	fmt.Fprintf(&b, `  <text x="%d" y="%d" text-anchor="%s" font-family="%s" font-size="%d" fill="%s"%s%s>%s</text>`+"\n",
		textX, textY, anchor, imgutils.EscapeXML(family), fontSize, imgutils.EscapeXML(color), weightAttr, transformAttr, imgutils.EscapeXML(displayText))

	return imgcore.LayerOutput{
		Fragment: b.String(),
		Width:    textW,
		Height:   textH,
	}
}
