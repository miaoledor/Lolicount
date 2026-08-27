package server

import (
	"fmt"
	"math/rand"
	"strconv"

	"github.com/gofiber/fiber/v3"

	"github.com/miaoledor/lolicount/internal/imgcore"
	"github.com/miaoledor/lolicount/internal/imgcore/composer"
	"github.com/miaoledor/lolicount/internal/imgcore/render"
	"github.com/miaoledor/lolicount/internal/imgcore/theme"
)

// editorPreviewHandler handles POST /api/editor/preview. It receives
// the editor layer stack as JSON, builds a *theme.Theme using a single
// GroupLayer (unified PSD coordinate model), appends the text layer,
// and renders the final SVG via composer.Compose. No data is persisted.
func (s *Server) editorPreviewHandler(c fiber.Ctx) error {
	var req EditorRequest
	if err := c.Bind().JSON(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid JSON: "+err.Error())
	}

	t, err := buildEditorTheme(&req)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}

	seed := "editor:" + strconv.FormatInt(rand.Int63(), 16)
	svg, err := composer.Compose(composer.ComposeParams{
		Theme: t, Seed: seed, CountText: req.Text,
	})
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "compose: "+err.Error())
	}

	c.Set("Content-Type", "image/svg+xml")
	c.Set("Cache-Control", "no-store")
	return c.Status(fiber.StatusOK).SendString(svg)
}

// buildEditorTheme converts an EditorRequest into a renderable
// *theme.Theme. All editor layers are assembled into a single
// GroupLayer with PSD absolute coordinates — this is the unified model
// for both card (single-layer) and character (multi-layer) themes in
// the editor. The text layer is appended via buildThemeLayers so text
// positioning and canvas sizing reuse the same logic as the counter
// render path.
func buildEditorTheme(req *EditorRequest) (*theme.Theme, error) {
	if len(req.Layers) == 0 {
		return nil, fmt.Errorf("at least one layer is required")
	}
	if req.Canvas.Width <= 0 || req.Canvas.Height <= 0 {
		return nil, fmt.Errorf("canvas dimensions must be positive")
	}

	groupParts, err := buildGroupParts(req.Layers)
	if err != nil {
		return nil, err
	}
	if len(groupParts) == 0 {
		return nil, fmt.Errorf("no valid image parts in layers")
	}

	canvasW := req.Canvas.Width
	canvasH := req.Canvas.Height
	outW, outH := canvasW, canvasH
	vbX, vbY, vbW, vbH := 0, 0, canvasW, canvasH

	if req.Display != nil && req.Display.Size > 0 {
		vbW, vbH = canvasW, canvasH
		if req.Display.Crop != nil && req.Display.Crop.Width > 0 && req.Display.Crop.Height > 0 {
			vbW = req.Display.Crop.Width
			vbH = req.Display.Crop.Height
			vbX = req.Display.Crop.Left
			vbY = req.Display.Crop.Top
		}
		outH = req.Display.Size
		outW = int(float64(vbW) * float64(outH) / float64(vbH))
		if outW < 1 {
			outW = 1
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

	base := &theme.Theme{
		Name:    req.Name,
		Canvas:  theme.Canvas{Width: outW, Height: outH},
		BgW:     outW,
		BgH:     outH,
		Display: req.Display,
		Layers:  []imgcore.Layer{groupLayer},
	}

	text := req.Text
	if text == "" {
		text = "0"
	}
	return buildThemeLayers(base, req.Scale, text, req.FSize, req.UnshowF,
		theme.TextStyle{}, theme.TextPos{})
}

// buildGroupParts converts editor layers into GroupParts for the
// GroupLayer. Each layer with images becomes one GroupPart; a layer
// with multiple images stores them as Candidates for random selection
// at render time. Layers with no images are skipped.
func buildGroupParts(layers []EditorLayer) ([]render.GroupPart, error) {
	var parts []render.GroupPart
	for _, layer := range layers {
		if len(layer.Images) == 0 {
			continue
		}
		first := layer.Images[0]
		part := render.GroupPart{
			Src:    first.Src,
			X:      first.Left,
			Y:      first.Top,
			Width:  first.Width,
			Height: first.Height,
		}
		if len(layer.Images) > 1 {
			cands := make([]render.GroupCandidate, len(layer.Images))
			for i, img := range layer.Images {
				cands[i] = render.GroupCandidate{
					Src:    img.Src,
					X:      img.Left,
					Y:      img.Top,
					Width:  img.Width,
					Height: img.Height,
				}
			}
			part.Candidates = cands
		}
		parts = append(parts, part)
	}
	return parts, nil
}
