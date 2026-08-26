// Package server bridge.go is a transitional adapter that converts old
// registry data (cardthemedrawer/characterthemedrawer types) into the
// new theme.Theme + []imgcore.Layer model expected by composer.Compose.
// This will be removed once the old drawer packages are deleted and the
// asset loaders directly produce theme.Theme.
package server

import (
	"fmt"
	"math/rand"

	"github.com/miaoledor/lolicount/internal/imgcore"
	"github.com/miaoledor/lolicount/internal/imgcore/asset"
	"github.com/miaoledor/lolicount/internal/imgcore/composer"
	"github.com/miaoledor/lolicount/internal/imgcore/imgutils"
	"github.com/miaoledor/lolicount/internal/imgcore/render"
	"github.com/miaoledor/lolicount/internal/imgcore/theme"
)

// buildCardThemeLayers converts a card theme frame into a theme.Theme
// with a single ImageLayer (the selected frame) plus a TextLayer.
//
// The canvas dimensions are computed from the scaled image size. The
// image layer's Scale is set to the ratio (displaySize / originalSize)
// so the rendered <image> width/height matches the canvas viewBox
// exactly — without this, the image renders at its original pixel
// dimensions and overflows the viewBox.
func buildCardThemeLayers(frame render.ImageLayer, scale float64, text string,
	fontSize int, unshowFont bool, style theme.TextStyle, pos theme.TextPos) *theme.Theme {

	s := scaleOrOne(scale)
	display := imgutils.DisplaySize(s)
	imgW, imgH := imgutils.ScaledDims(frame.Width, frame.Height, display)

	// Compute the per-axis scale so the image renders at the canvas size.
	scaleX := 1.0
	scaleY := 1.0
	if frame.Width > 0 {
		scaleX = float64(imgW) / float64(frame.Width)
	}
	if frame.Height > 0 {
		scaleY = float64(imgH) / float64(frame.Height)
	}

	frame.Transform = imgcore.Transform{
		X:        imgcore.FixedRange(0),
		Y:        imgcore.FixedRange(0),
		Scale:    imgcore.FixedRange(scaleX),
		Rotation: imgcore.FixedRange(0),
	}
	frame.Z = 0

	textLayer := &render.TextLayer{
		Text:       text,
		FontSize:   fontSize,
		UnshowFont: unshowFont,
		Style:      style,
		Position:   pos,
		Transform:  imgcore.DefaultTransform(),
		Z:          1,
	}

	textW, textH := render.MeasureText(text, fontSize, unshowFont)
	canvasW := imgW
	if textW > canvasW {
		canvasW = textW
	}
	canvasH := imgH + textH

	// scaleY is only used when the image has a different aspect ratio
	// than the display size; in practice ScaledDims preserves aspect
	// ratio so scaleX == scaleY.
	_ = scaleY

	return &theme.Theme{
		Canvas: theme.Canvas{Width: canvasW, Height: canvasH},
		Layers: []imgcore.Layer{&frame, textLayer},
	}
}

// buildCharacterThemeLayers converts an assembled character portrait
// into a theme.Theme with a GroupLayer (nested <svg> for PSD coordinate
// mapping) plus a TextLayer.
//
// The GroupLayer uses a nested <svg viewBox> to map the PSD canvas
// (or crop region) onto the output viewport, preserving sub-pixel
// alignment between layers. This mirrors the old
// characterthemedrawer.drawLayeredSVG approach.
func buildCharacterThemeLayers(parts []asset.CharacterPart, canvasW, canvasH int,
	display *theme.DisplayConfig, scale float64, text string,
	fontSize int, unshowFont bool, style theme.TextStyle, pos theme.TextPos) *theme.Theme {

	outW, outH := canvasW, canvasH
	vbX, vbY, vbW, vbH := 0, 0, canvasW, canvasH

	if display != nil && display.Size > 0 {
		vbW, vbH = canvasW, canvasH
		if display.Crop != nil && display.Crop.Width > 0 && display.Crop.Height > 0 {
			vbW = display.Crop.Width
			vbH = display.Crop.Height
			vbX = display.Crop.Left
			vbY = display.Crop.Top
		}
		outH = display.Size
		outW = int(float64(vbW) * float64(outH) / float64(vbH))
		if outW < 1 {
			outW = 1
		}
	} else {
		s := scaleOrOne(scale)
		dispSize := imgutils.DisplaySize(s)
		outW, outH = imgutils.ScaledCanvasDims(canvasW, canvasH, dispSize)
	}

	groupParts := make([]render.GroupPart, len(parts))
	for i, part := range parts {
		groupParts[i] = render.GroupPart{
			Src:    part.Data,
			X:      part.Left,
			Y:      part.Top,
			Width:  part.Width,
			Height: part.Height,
		}
	}

	groupLayer := &render.GroupLayer{
		Parts: groupParts,
		OutW:  outW,
		OutH:  outH,
		VbX:   vbX,
		VbY:   vbY,
		VbW:   vbW,
		VbH:   vbH,
		Z:     0,
	}

	textLayer := &render.TextLayer{
		Text:       text,
		FontSize:   fontSize,
		UnshowFont: unshowFont,
		Style:      style,
		Position:   pos,
		Transform:  imgcore.DefaultTransform(),
		Z:          1,
	}

	textW, textH := render.MeasureText(text, fontSize, unshowFont)
	canvasW2 := outW
	if textW > canvasW2 {
		canvasW2 = textW
	}
	canvasH2 := outH + textH

	return &theme.Theme{
		Canvas:  theme.Canvas{Width: canvasW2, Height: canvasH2},
		Display: display,
		Layers:  []imgcore.Layer{groupLayer, textLayer},
	}
}

// scaleOrOne returns scale when > 0, otherwise 1.
func scaleOrOne(scale float64) float64 {
	if scale <= 0 {
		return 1
	}
	return scale
}

// frameIndexForCount picks the background frame for a given count.
// Per M2.5: display frame[(count+1) % size].
func frameIndexForCount(count int64, size int) int {
	if size <= 1 {
		return 0
	}
	idx := int((count + 1) % int64(size))
	if idx < 0 {
		idx += size
	}
	return idx
}

// pickCardFrame selects a frame from a card theme.
func pickCardFrame(th *asset.CardTheme, mode string, frameIndex int) (render.ImageLayer, bool) {
	if mode == "random" {
		return th.FrameAt(rand.Intn(th.Size()))
	}
	return th.FrameAt(frameIndex)
}

// resolveMode determines the render mode for a theme.
func resolveMode(kind string, modeParam string) string {
	if kind == "character" {
		return "random"
	}
	if modeParam == "random" {
		return "random"
	}
	return "seq"
}

// composeCard renders a card theme via the new composer.
func (s *Server) composeCard(entry composer.ThemeEntry, q *queryParams, text string,
	frameIndex int, style theme.TextStyle) (string, error) {
	th, ok := s.themes.GetCard(entry.Name)
	if !ok {
		return "", fmt.Errorf("card theme not found")
	}
	mode := resolveMode(entry.Kind, q.Mode)
	frame, ok := pickCardFrame(th, mode, frameIndex)
	if !ok {
		return "", fmt.Errorf("frame index out of range")
	}
	pos := theme.TextPos{X: q.X, Y: q.Y, RX: q.RX, RY: q.RY}
	t := buildCardThemeLayers(frame, q.Scale, text, q.FSize, q.UnshowF, style, pos)
	return composer.Compose(composer.ComposeParams{Theme: t, Seed: entry.Name, CountText: text})
}

// composeCharacter renders a character theme via the new composer.
func (s *Server) composeCharacter(entry composer.ThemeEntry, q *queryParams, text string,
	style theme.TextStyle) (string, error) {
	ch, ok := s.themes.GetCharacter(entry.Name)
	if !ok {
		return "", fmt.Errorf("character theme not found")
	}
	parts, canvasW, canvasH, display, err := assembleCharacter(ch)
	if err != nil {
		return "", err
	}
	pos := theme.TextPos{X: q.X, Y: q.Y, RX: q.RX, RY: q.RY}
	t := buildCharacterThemeLayers(parts, canvasW, canvasH, display, q.Scale, text, q.FSize, q.UnshowF, style, pos)
	return composer.Compose(composer.ComposeParams{Theme: t, Seed: entry.Name, CountText: text})
}

// assembleCharacter picks one layer from each category using the
// manifest ranges. Transitional bridge until RandomPickLayers are
// constructed directly.
func assembleCharacter(ch *asset.CharacterTheme) ([]asset.CharacterPart, int, int, *theme.DisplayConfig, error) {
	if ch == nil || len(ch.Manifest) == 0 || ch.Config == nil {
		return nil, 0, 0, nil, fmt.Errorf("character theme has no config")
	}

	var parts []asset.CharacterPart
	for _, cat := range []string{"lass", "eye", "brow", "mouth", "face"} {
		rng, ok := ch.Config.Ranges[cat]
		if !ok {
			continue
		}
		layer, err := pickManifestLayer(ch, rng)
		if err != nil {
			return nil, 0, 0, nil, fmt.Errorf("character %s: %w", cat, err)
		}
		img, ok := ch.Parts[layer.LayerID]
		if !ok {
			return nil, 0, 0, nil, fmt.Errorf("character %s: layer_id %d not decoded", cat, layer.LayerID)
		}
		parts = append(parts, asset.CharacterPart{
			Left:   layer.Left,
			Top:    layer.Top,
			Width:  img.Width,
			Height: img.Height,
			Data:   img.Src,
		})
	}

	return parts, ch.Config.CanvasW, ch.Config.CanvasH, ch.Display, nil
}

// pickManifestLayer randomly selects a layer index in [First, Last]
// from the manifest.
func pickManifestLayer(ch *asset.CharacterTheme, rng asset.PartRange) (asset.CharacterManifest, error) {
	if rng.First < 0 || rng.Last >= len(ch.Manifest) || rng.First > rng.Last {
		return asset.CharacterManifest{}, fmt.Errorf("range [%d,%d] out of bounds (manifest=%d)", rng.First, rng.Last, len(ch.Manifest))
	}
	idx := rng.First
	if rng.Last > rng.First {
		idx = rng.First + rand.Intn(rng.Last-rng.First+1)
	}
	return ch.Manifest[idx], nil
}
