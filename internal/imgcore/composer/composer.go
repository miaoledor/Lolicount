// Package composer is the sole rendering entry point for imgcore: it
// iterates a theme's layer stack, calls each layer's Render method, and
// concatenates the SVG fragments into the final document.
package composer

import (
	"fmt"
	"sort"
	"strings"

	"github.com/miaoledor/lolicount/internal/imgcore"
	"github.com/miaoledor/lolicount/internal/imgcore/imgutils"
	"github.com/miaoledor/lolicount/internal/imgcore/theme"
)

// ComposeParams controls how the layer stack is composed into the final
// SVG. The caller resolves the counter text string (from count/number/
// demo) before calling Compose — the composer only renders, it does not
// decide what text to show.
type ComposeParams struct {
	Theme     *theme.Theme
	Seed      string // drives the PRNG (typically the counter name)
	CountText string
}

// Compose renders the final SVG by iterating the theme's layer stack in
// ZIndex order. Each layer's Render method produces an SVG fragment; the
// composer concatenates them inside a single <svg> wrapper. The viewBox
// is determined by the theme's canvas dimensions (or the union of all
// layer dimensions when the canvas is zero).
func Compose(p ComposeParams) (string, error) {
	if p.Theme == nil {
		return "", fmt.Errorf("composer: theme is nil")
	}

	prng := imgutils.NewPRNG(p.Seed)

	canvasW := p.Theme.Canvas.Width
	canvasH := p.Theme.Canvas.Height

	if canvasW == 0 || canvasH == 0 {
		for _, layer := range p.Theme.Layers {
			ctx := imgcore.RenderCtx{PRNG: prng, CanvasW: canvasW, CanvasH: canvasH, CountText: p.CountText}
			out := layer.Render(ctx)
			if out.Width > canvasW {
				canvasW = out.Width
			}
			if out.Height > canvasH {
				canvasH = out.Height
			}
		}
	}

	bgW := p.Theme.BgW
	if bgW == 0 {
		bgW = canvasW
	}
	bgH := p.Theme.BgH
	if bgH == 0 {
		bgH = canvasH
	}

	ctx := imgcore.RenderCtx{
		PRNG:      prng,
		CanvasW:   canvasW,
		CanvasH:   canvasH,
		BgW:       bgW,
		BgH:       bgH,
		CountText: p.CountText,
	}

	layers := make([]imgcore.Layer, len(p.Theme.Layers))
	copy(layers, p.Theme.Layers)
	sort.SliceStable(layers, func(i, j int) bool {
		return layers[i].ZIndex() < layers[j].ZIndex()
	})

	var b strings.Builder
	b.WriteString(`<?xml version="1.0" encoding="UTF-8"?>` + "\n")
	fmt.Fprintf(&b, `<svg viewBox="0 0 %d %d" width="%d" height="%d" version="1.1" xmlns="http://www.w3.org/2000/svg" xmlns:xlink="http://www.w3.org/1999/xlink">`+"\n",
		canvasW, canvasH, canvasW, canvasH)
	b.WriteString("  <title>Lolicount</title>\n")

	for _, layer := range layers {
		out := layer.Render(ctx)
		b.WriteString(out.Fragment)
	}

	b.WriteString("</svg>\n")
	return b.String(), nil
}
