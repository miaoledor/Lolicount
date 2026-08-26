// Package server compose.go is the unified theme-to-SVG composition
// path. It replaces the old bridge.go dual-path (buildCardThemeLayers +
// buildCharacterThemeLayers) with a single buildThemeLayers function
// that works for all theme types. The registry already returns a
// *theme.Theme with the correct layer stack; this file adds the text
// layer and computes the final canvas dimensions.
//
// All themes use random frame/layer selection — there is no sequential
// mode. Each request salts the PRNG seed with a per-request random
// number so multi-frame themes show a different frame each time.
package server

import (
	"fmt"
	"math/rand"
	"strconv"

	"github.com/miaoledor/lolicount/internal/imgcore"
	"github.com/miaoledor/lolicount/internal/imgcore/composer"
	"github.com/miaoledor/lolicount/internal/imgcore/imgutils"
	"github.com/miaoledor/lolicount/internal/imgcore/render"
	"github.com/miaoledor/lolicount/internal/imgcore/theme"
)

// buildThemeLayers takes a loaded *theme.Theme (from the registry) and
// produces a renderable theme with the text layer appended. For card
// (single-image) themes, scale is applied to the image dimensions. For
// character (multi-layer) themes, the display config already handles
// sizing, so scale is only used when no display config is present.
func buildThemeLayers(base *theme.Theme, scale float64, text string,
	fontSize int, unshowFont bool, style theme.TextStyle, pos theme.TextPos) (*theme.Theme, error) {

	if base == nil {
		return nil, fmt.Errorf("buildThemeLayers: nil theme")
	}

	layers := make([]imgcore.Layer, 0, len(base.Layers)+1)

	// Process existing layers (image/group layers from the registry).
	// Scale is applied to plain ImageLayers (single-image themes).
	// GroupLayers and RandomPickLayers handle their own sizing.
	for _, layer := range base.Layers {
		if il, ok := layer.(*render.ImageLayer); ok {
			s := scaleOrOne(scale)
			display := imgutils.DisplaySize(s)
			imgW, imgH := imgutils.ScaledDims(il.Width, il.Height, display)
			il.Width = imgW
			il.Height = imgH
			il.Transform = imgcore.DefaultTransform()
		}
		layers = append(layers, layer)
	}

	// Append the text layer.
	textLayer := &render.TextLayer{
		Text:       text,
		FontSize:   fontSize,
		UnshowFont: unshowFont,
		Style:      style,
		Position:   pos,
		Transform:  imgcore.DefaultTransform(),
		Z:          len(layers),
	}
	layers = append(layers, textLayer)

	// Compute canvas dimensions from the first ImageLayer if present
	// (single-image themes). GroupLayers carry their own output dims.
	bgW := base.BgW
	bgH := base.BgH
	if len(base.Layers) > 0 {
		if il, ok := base.Layers[0].(*render.ImageLayer); ok {
			bgW = il.Width
			bgH = il.Height
		}
	}

	textW, textH := render.MeasureText(text, fontSize, unshowFont)
	canvasW := bgW
	if textW > canvasW {
		canvasW = textW
	}
	canvasH := bgH + textH

	return &theme.Theme{
		Name:    base.Name,
		Canvas:  theme.Canvas{Width: canvasW, Height: canvasH},
		BgW:     bgW,
		BgH:     bgH,
		Display: base.Display,
		Layers:  layers,
	}, nil
}

// scaleOrOne returns scale when > 0, otherwise 1.
func scaleOrOne(scale float64) float64 {
	if scale <= 0 {
		return 1
	}
	return scale
}

// compose renders any theme (card or character) via the unified
// composer. The theme is fetched from the registry, the text layer is
// appended, and the result is composed into SVG. The PRNG seed is
// salted with a per-request random number so multi-frame themes show a
// different frame/combination on each request.
func (s *Server) compose(entry composer.ThemeEntry, q *queryParams, text string,
	style theme.TextStyle) (string, error) {
	base, ok := s.themes.Get(entry.Name)
	if !ok {
		return "", fmt.Errorf("theme %q not found", entry.Name)
	}

	pos := theme.TextPos{X: q.X, Y: q.Y, RX: q.RX, RY: q.RY}
	t, err := buildThemeLayers(base, q.Scale, text, q.FSize, q.UnshowF, style, pos)
	if err != nil {
		return "", err
	}

	seed := entry.Name + ":" + strconv.FormatInt(rand.Int63(), 16)
	return composer.Compose(composer.ComposeParams{Theme: t, Seed: seed, CountText: text})
}

// themeIsMultiFrame reports whether the named theme has multiple frames
// (i.e. random selection produces different output per request). Used by
// the counter handler to decide cache policy: multi-frame demo requests
// must be no-store because the background changes each request.
func (s *Server) themeIsMultiFrame(name string) bool {
	base, ok := s.themes.Get(name)
	if !ok {
		return false
	}
	for _, layer := range base.Layers {
		if _, ok := layer.(*render.RandomPickLayer); ok {
			return true
		}
		if _, ok := layer.(*render.GroupLayer); ok {
			return true
		}
	}
	return false
}
