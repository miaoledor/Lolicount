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
func buildCardThemeLayers(frame render.ImageLayer, scale float64, text string,
	fontSize int, unshowFont bool, style theme.TextStyle, pos theme.TextPos) *theme.Theme {

	frame.Transform = imgcore.Transform{
		X:        imgcore.FixedRange(0),
		Y:        imgcore.FixedRange(0),
		Scale:    imgcore.FixedRange(scaleOrOne(scale)),
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

	display := imgutils.DisplaySize(scaleOrOne(scale))
	imgW, imgH := imgutils.ScaledDims(frame.Width, frame.Height, display)

	return &theme.Theme{
		Canvas: theme.Canvas{Width: imgW, Height: imgH},
		Layers: []imgcore.Layer{&frame, textLayer},
	}
}

// buildCharacterThemeLayers converts an assembled character portrait
// into a theme.Theme with ImageLayers (one per part) plus a TextLayer.
func buildCharacterThemeLayers(parts []asset.CharacterPart, canvasW, canvasH int,
	display *theme.DisplayConfig, scale float64, text string,
	fontSize int, unshowFont bool, style theme.TextStyle, pos theme.TextPos) *theme.Theme {

	layers := make([]imgcore.Layer, 0, len(parts)+1)
	for i, part := range parts {
		img := render.ImageLayer{
			Src:       part.Data,
			Width:     part.Width,
			Height:    part.Height,
			Transform: imgcore.Transform{
				X:        imgcore.FixedRange(float64(part.Left)),
				Y:        imgcore.FixedRange(float64(part.Top)),
				Scale:    imgcore.FixedRange(1),
				Rotation: imgcore.FixedRange(0),
			},
			Z: i,
		}
		layers = append(layers, &img)
	}

	textLayer := &render.TextLayer{
		Text:       text,
		FontSize:   fontSize,
		UnshowFont: unshowFont,
		Style:      style,
		Position:   pos,
		Transform:  imgcore.DefaultTransform(),
		Z:          len(parts),
	}
	layers = append(layers, textLayer)

	return &theme.Theme{
		Canvas:  theme.Canvas{Width: canvasW, Height: canvasH},
		Display: display,
		Layers:  layers,
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
// manifest ranges. Transitional bridge until the old characterthemedrawer
// is removed and RandomPickLayers are constructed directly.
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
