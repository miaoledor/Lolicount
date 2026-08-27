package server

import (
	"encoding/base64"
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
	// When layers exist but none have images yet (e.g. user just
	// clicked "add layer" but hasn't uploaded), render an empty canvas
	// with just the text layer instead of erroring. This lets the user
	// see the canvas + grid before adding images.

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

// placeholderImageSize is the default dimension for empty-layer
// placeholders shown in the editor preview.
const placeholderImageSize = 100

// buildGroupParts converts editor layers into GroupParts for the
// GroupLayer. Each layer with images becomes one GroupPart; a layer
// with multiple images stores them as Candidates for random selection
// at render time. Layers with no images get a placeholder part (dashed
// border + layer name label) so the user can see where each layer sits
// on the canvas before uploading images.
func buildGroupParts(layers []EditorLayer) ([]render.GroupPart, error) {
	var parts []render.GroupPart
	for _, layer := range layers {
		if len(layer.Images) == 0 {
			placeholder := makePlaceholderSVG(layer.Name, placeholderImageSize, placeholderImageSize)
			parts = append(parts, render.GroupPart{
				Src:    placeholder,
				X:      0,
				Y:      0,
				Width:  placeholderImageSize,
				Height: placeholderImageSize,
			})
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

// makePlaceholderSVG generates a data-URI SVG image that renders as a
// dashed-border rectangle with the layer name label centered
// inside. Used for empty layers in the editor preview so the user can
// see each layer's position and default size before uploading images.
func makePlaceholderSVG(name string, w, h int) string {
	label := name
	if label == "" {
		label = "layer"
	}
	svg := fmt.Sprintf(
		`<svg xmlns="http://www.w3.org/2000/svg" width="%d" height="%d" viewBox="0 0 %d %d">`+
			`<rect x="1" y="1" width="%d" height="%d" fill="none" stroke="#999" stroke-width="2" stroke-dasharray="6 4"/>`+
			`<text x="%d" y="%d" fill="#999" font-size="14" font-family="monospace" text-anchor="middle" dominant-baseline="middle">%s</text>`+
			`</svg>`,
		w, h, w, h, w-2, h-2, w/2, h/2, label,
	)
	return "data:image/svg+xml;base64," + base64.StdEncoding.EncodeToString([]byte(svg))
}
