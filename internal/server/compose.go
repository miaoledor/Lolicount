// Package server compose.go is the unified theme-to-SVG composition
// path. The registry returns a *theme.Theme with the correct layer
// stack; this file adds the text layer and computes the final canvas
// dimensions.
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

// BuildThemeLayers takes a loaded *theme.Theme (from the registry) and
// produces a renderable theme with the text layer appended. Scale is
// applied to image dimensions for single-image (ImageLayer) themes and
// multi-frame (RandomPickLayer) themes. Character (GroupLayer) themes
// handle their own sizing via display config, so scale only applies when
// no display config is present.
//
// This function does NOT mutate the registry's shared *theme.Theme — it
// builds fresh layers so concurrent requests and repeated calls see
// consistent dimensions.
//
// Exported so cmd/gen-theme-thumbs can pre-render identical gallery
// thumbnails without a running server.
func BuildThemeLayers(base *theme.Theme, scale float64, text string,
	fontSize int, unshowFont bool, style theme.TextStyle, pos theme.TextPos) (*theme.Theme, error) {

	if base == nil {
		return nil, fmt.Errorf("BuildThemeLayers: nil theme")
	}

	s := scaleOrOne(scale)
	display := imgutils.DisplaySize(s)

	layers := make([]imgcore.Layer, 0, len(base.Layers)+1)
	bgW, bgH := base.BgW, base.BgH

	for _, layer := range base.Layers {
		switch l := layer.(type) {
		case *render.ImageLayer:
			imgW, imgH := imgutils.ScaledDims(l.Width, l.Height, display)
			cp := *l
			cp.Width = imgW
			cp.Height = imgH
			cp.Transform = imgcore.DefaultTransform()
			layers = append(layers, &cp)
			bgW, bgH = imgW, imgH
		case *render.RandomPickLayer:
			// Scale every option to the display size and take the max.
			// Under the lazy one-image-per-slot strategy Options holds
			// only the resident frame, so FrameDims (every candidate's
			// dims) is folded in as well — the canvas must accommodate
			// the largest frame so no frame is clipped and the aspect
			// ratio stays stable as swaps rotate through frames.
			scaledOpts := make([]render.ImageOption, len(l.Options))
			var maxW, maxH int
			for i, opt := range l.Options {
				imgW, imgH := imgutils.ScaledDims(opt.Width, opt.Height, display)
				cp := opt.ImageLayer
				cp.Width = imgW
				cp.Height = imgH
				cp.Transform = imgcore.DefaultTransform()
				scaledOpts[i] = render.ImageOption{ImageLayer: cp, Weight: opt.Weight}
				if imgW > maxW {
					maxW = imgW
				}
				if imgH > maxH {
					maxH = imgH
				}
			}
			for _, d := range l.FrameDims {
				imgW, imgH := imgutils.ScaledDims(d.W, d.H, display)
				if imgW > maxW {
					maxW = imgW
				}
				if imgH > maxH {
					maxH = imgH
				}
			}
			layers = append(layers, &render.RandomPickLayer{
				Category:  l.Category,
				Options:   scaledOpts,
				FrameDims: l.FrameDims,
				Transform: imgcore.DefaultTransform(),
				Z:         l.Z,
				IsFixed:   l.IsFixed,
			})
			bgW, bgH = maxW, maxH
		default:
			layers = append(layers, l)
		}
	}

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
// appended, and the result is composed into SVG. When the registry is a
// lazy RenderPreparer (the one-image-per-slot strategy), each render
// first swaps in ONE freshly loaded slot candidate and works on a
// private snapshot — consecutive requests each evolve one part while
// concurrent renders stay isolated. The PRNG seed is salted with a
// per-request random number so the text-layer randomness differs per
// request.
func (s *Server) compose(entry composer.ThemeEntry, q *queryParams, text string,
	style theme.TextStyle) (string, error) {
	var base *theme.Theme
	if rp, ok := s.themes.(composer.RenderPreparer); ok {
		// PrepareRender already returns a private snapshot; fetching the
		// shared assembly first would assemble a cold theme twice.
		snapshot, err := rp.PrepareRender(entry.Name)
		if err != nil {
			return "", err
		}
		base = snapshot
	} else {
		var ok bool
		base, ok = s.themes.Get(entry.Name)
		if !ok {
			return "", fmt.Errorf("theme %q not found", entry.Name)
		}
	}

	pos := theme.TextPos{X: q.X, Y: q.Y, RX: q.RX, RY: q.RY}
	t, err := BuildThemeLayers(base, q.Scale, text, q.FSize, q.UnshowF, style, pos)
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
	if info, ok := s.themes.(composer.ThemeInfo); ok {
		return info.IsMultiFrame(name)
	}
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
