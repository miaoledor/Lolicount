// Package render implements the stateless rendering of each layer type
// into SVG fragments. Each file handles one layer kind; the composer
// (internal/imgcore/composer.go) concatenates fragments in ZIndex order.
package render

import (
	"fmt"
	"strings"

	"github.com/miaoledor/lolicount/internal/imgcore"
)

// ImageLayer renders a single fixed image as an <image> element with
// optional transform (position, scale, rotation). The image source is
// a data URI (base64) for offline rendering, per AGENTS.md Iron Rule 2.
type ImageLayer struct {
	Src       string          // data:<mime>;base64,... or CDN URL
	Width     int             // original pixel width
	Height    int             // original pixel height
	Transform imgcore.Transform
	Z         int
	IsFixed   bool
}

// Kind returns LayerImage.
func (l *ImageLayer) Kind() imgcore.LayerKind { return imgcore.LayerImage }

// ZIndex returns the stack order.
func (l *ImageLayer) ZIndex() int { return l.Z }

// Fixed reports whether the layer cannot be deleted.
func (l *ImageLayer) Fixed() bool { return l.IsFixed }

// Render produces the SVG <image> fragment. The PRNG resolves any
// random Range values in the Transform. Scale is applied to the
// original pixel dimensions; rotation is around the image center.
func (l *ImageLayer) Render(ctx imgcore.RenderCtx) imgcore.LayerOutput {
	x := ctx.PRNG.FloatRange(l.Transform.X)
	y := ctx.PRNG.FloatRange(l.Transform.Y)
	scale := ctx.PRNG.FloatRange(l.Transform.Scale)
	rotation := ctx.PRNG.FloatRange(l.Transform.Rotation)

	if scale <= 0 {
		scale = 1
	}
	imgW := int(float64(l.Width) * scale)
	imgH := int(float64(l.Height) * scale)
	if imgW < 1 {
		imgW = 1
	}
	if imgH < 1 {
		imgH = 1
	}

	transform := buildImageTransform(x, y, rotation, float64(imgW)/2, float64(imgH)/2)

	var b strings.Builder
	if transform != "" {
		fmt.Fprintf(&b, `  <image x="0" y="0" width="%d" height="%d" transform="%s" xlink:href="%s" />`+"\n",
			imgW, imgH, transform, l.Src)
	} else {
		fmt.Fprintf(&b, `  <image x="%d" y="%d" width="%d" height="%d" xlink:href="%s" />`+"\n",
			int(x), int(y), imgW, imgH, l.Src)
	}
	return imgcore.LayerOutput{Fragment: b.String(), Width: imgW, Height: imgH}
}

// buildImageTransform assembles an SVG transform string for translate +
// rotate. When rotation is 0 and position is (0,0), returns empty string
// so the caller can use a simpler <image> tag without transform.
func buildImageTransform(x, y, rotation, cx, cy float64) string {
	var parts []string
	if x != 0 || y != 0 {
		parts = append(parts, fmt.Sprintf("translate(%s %s)", formatFloat(x), formatFloat(y)))
	}
	if rotation != 0 {
		parts = append(parts, fmt.Sprintf("rotate(%s %s %s)", formatFloat(rotation), formatFloat(cx), formatFloat(cy)))
	}
	return strings.Join(parts, " ")
}

// formatFloat formats a float for SVG attribute output, trimming
// trailing zeros.
func formatFloat(v float64) string {
	return strings.TrimSuffix(strings.TrimRight(fmt.Sprintf("%.2f", v), "0"), ".")
}
