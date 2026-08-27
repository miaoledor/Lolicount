package server

import (
	"archive/zip"
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/gofiber/fiber/v3"

	"github.com/miaoledor/lolicount/internal/imgcore/asset"
	"github.com/miaoledor/lolicount/internal/imgcore/theme"
)

// reservedThemeNames cannot be used for exported themes.
var reservedThemeNames = map[string]bool{
	"demo":   true,
	"random": true,
}

// editorExportHandler handles POST /api/editor/export. It receives the
// editor layer stack as JSON, re-encodes all images server-side (Iron
// Rule 4: never trust client format), packages the theme as a zip, and
// returns it for download. Single-layer themes export as card themes
// (0..n-1.<ext>), multi-layer themes export as character themes
// (ren.json + config.json + ren/<id>.<ext>).
func (s *Server) editorExportHandler(c fiber.Ctx) error {
	var req EditorRequest
	if err := c.Bind().JSON(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid JSON: "+err.Error())
	}

	if err := validateExportName(req.Name); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}

	// Collect all non-empty layers (with at least one image).
	var activeLayers []EditorLayer
	for _, l := range req.Layers {
		if len(l.Images) > 0 {
			activeLayers = append(activeLayers, l)
		}
	}
	if len(activeLayers) == 0 {
		return fiber.NewError(fiber.StatusBadRequest, "at least one layer with images is required")
	}
	if req.Canvas.Width <= 0 || req.Canvas.Height <= 0 {
		return fiber.NewError(fiber.StatusBadRequest, "canvas dimensions must be positive")
	}

	// Decode all images from data URIs, re-encode server-side.
	decodedImgs, err := decodeAndReencodeImages(activeLayers)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}

	var zipBytes []byte
	if len(activeLayers) == 1 && len(activeLayers[0].Images) <= 1 {
		zipBytes, err = exportCardTheme(req.Name, activeLayers, decodedImgs)
	} else {
		zipBytes, err = exportCharacterTheme(req.Name, req.Canvas, req.Display, activeLayers, decodedImgs)
	}
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}

	c.Set("Content-Type", "application/zip")
	c.Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s.zip"`, req.Name))
	return c.Status(fiber.StatusOK).Send(zipBytes)
}

// validateExportName checks the theme name against reserved words and
// character rules (ASCII letters, digits, hyphens only).
func validateExportName(name string) error {
	if name == "" {
		return fmt.Errorf("theme name is required")
	}
	if reservedThemeNames[name] {
		return fmt.Errorf("theme name %q is reserved", name)
	}
	for _, r := range name {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '-' {
			continue
		}
		return fmt.Errorf("theme name must be ASCII letters, digits, or hyphens only")
	}
	return nil
}

// decodedImage holds the re-encoded image bytes and the file extension
// determined by the actual output MIME type (not the client's claim).
type decodedImage struct {
	bytes []byte
	ext   string // ".png", ".gif", or ".webp"
}

// manifestEntry is one row in ren.json: the absolute placement of a
// layer in the PSD canvas. Exported in the zip for the character theme
// loader to consume.
type manifestEntry struct {
	Name         string `json:"name"`
	Left         int    `json:"left"`
	Top          int    `json:"top"`
	Width        int    `json:"width"`
	Height       int    `json:"height"`
	Visible      int    `json:"visible"`
	LayerID      int    `json:"layer_id"`
	GroupLayerID int    `json:"group_layer_id"`
}

// decodeAndReencodeImages decodes all data-URI images from the editor
// layers, re-encodes them server-side via asset.ReEncodeImage (Iron
// Rule 4), and returns a map of layerID -> []decodedImage.
func decodeAndReencodeImages(layers []EditorLayer) (map[int][]decodedImage, error) {
	result := make(map[int][]decodedImage)
	for _, layer := range layers {
		var imgs []decodedImage
		for _, img := range layer.Images {
			raw, err := extractDataURI(img.Src)
			if err != nil {
				return nil, fmt.Errorf("layer %d image: %w", layer.ID, err)
			}
			encoded, mime, err := asset.ReEncodeImage(raw, asset.EncodeWebP)
			if err != nil {
				return nil, fmt.Errorf("layer %d re-encode: %w", layer.ID, err)
			}
			ext := mimeToExt(mime)
			imgs = append(imgs, decodedImage{bytes: encoded, ext: ext})
		}
		result[layer.ID] = imgs
	}
	return result, nil
}

// extractDataURI decodes a base64 data URI to raw bytes. Supports
// data:<mime>;base64,<data> format. Non-data-URI sources are rejected
// (the editor must send image data inline).
func extractDataURI(src string) ([]byte, error) {
	const prefix = "base64,"
	idx := strings.Index(src, prefix)
	if idx < 0 {
		return nil, fmt.Errorf("expected base64 data URI")
	}
	b64 := src[idx+len(prefix):]
	return base64.StdEncoding.DecodeString(b64)
}

// mimeToExt maps a MIME type to a file extension.
func mimeToExt(mime string) string {
	switch mime {
	case "image/png":
		return ".png"
	case "image/gif":
		return ".gif"
	case "image/webp":
		return ".webp"
	default:
		return ".png"
	}
}

// exportCardTheme packages a single-layer, single-image theme as a card
// theme zip (0.<ext> at the top level).
func exportCardTheme(name string, layers []EditorLayer, imgs map[int][]decodedImage) ([]byte, error) {
	var buf bytes.Buffer
	w := zip.NewWriter(&buf)

	layer := layers[0]
	di := imgs[layer.ID][0]
	fw, err := w.Create(fmt.Sprintf("%s/0%s", name, di.ext))
	if err != nil {
		return nil, fmt.Errorf("zip create: %w", err)
	}
	if _, err := fw.Write(di.bytes); err != nil {
		return nil, fmt.Errorf("zip write: %w", err)
	}

	if err := w.Close(); err != nil {
		return nil, fmt.Errorf("zip close: %w", err)
	}
	return buf.Bytes(), nil
}

// exportCharacterTheme packages a multi-layer theme as a character
// theme zip (ren.json + config.json + ren/<id>.<ext>).
func exportCharacterTheme(name string, canvas EditorCanvas, display *theme.DisplayConfig,
	layers []EditorLayer, imgs map[int][]decodedImage) ([]byte, error) {
	var buf bytes.Buffer
	w := zip.NewWriter(&buf)


	layerIDCounter := -1
	var manifest []manifestEntry

	// Build manifest and write images. layer_id is the 0-based array
	// index into ren.json, matching the theme loader's convention
	// (collectCategoryCandidates indexes ct.Manifest[i] directly).
	for _, layer := range layers {
		layerImgs := imgs[layer.ID]
		for i, di := range layerImgs {
			layerIDCounter++
			img := layer.Images[i]
			manifest = append(manifest, manifestEntry{
				Name:         layer.Category,
				Left:         img.Left,
				Top:          img.Top,
				Width:        img.Width,
				Height:       img.Height,
				Visible:      1,
				LayerID:      layerIDCounter,
				GroupLayerID: 0,
			})
			fw, err := w.Create(fmt.Sprintf("%s/ren/%d%s", name, layerIDCounter, di.ext))
			if err != nil {
				return nil, fmt.Errorf("zip create image: %w", err)
			}
			if _, err := fw.Write(di.bytes); err != nil {
				return nil, fmt.Errorf("zip write image: %w", err)
			}
		}
	}

	// Write ren.json
	manifestJSON, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("marshal ren.json: %w", err)
	}
	fw, err := w.Create(fmt.Sprintf("%s/ren.json", name))
	if err != nil {
		return nil, fmt.Errorf("zip create ren.json: %w", err)
	}
	if _, err := fw.Write(manifestJSON); err != nil {
		return nil, fmt.Errorf("zip write ren.json: %w", err)
	}

	// Build config.json ranges from category grouping
	ranges := buildRanges(manifest)
	cfg := struct {
		CanvasW int                        `json:"canvasW"`
		CanvasH int                        `json:"canvasH"`
		Ranges  map[string]asset.PartRange `json:"ranges"`
	}{
		CanvasW: canvas.Width,
		CanvasH: canvas.Height,
		Ranges:  ranges,
	}
	cfgJSON, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("marshal config.json: %w", err)
	}
	fw, err = w.Create(fmt.Sprintf("%s/config.json", name))
	if err != nil {
		return nil, fmt.Errorf("zip create config.json: %w", err)
	}
	if _, err := fw.Write(cfgJSON); err != nil {
		return nil, fmt.Errorf("zip write config.json: %w", err)
	}

	// Write optional display.json
	if display != nil && display.Size > 0 {
		dispJSON, err := json.MarshalIndent(display, "", "  ")
		if err != nil {
			return nil, fmt.Errorf("marshal display.json: %w", err)
		}
		fw, err := w.Create(fmt.Sprintf("%s/display.json", name))
		if err != nil {
			return nil, fmt.Errorf("zip create display.json: %w", err)
		}
		if _, err := fw.Write(dispJSON); err != nil {
			return nil, fmt.Errorf("zip write display.json: %w", err)
		}
	}

	if err := w.Close(); err != nil {
		return nil, fmt.Errorf("zip close: %w", err)
	}
	return buf.Bytes(), nil
}

// buildRanges computes 0-based closed array index ranges per category
// from the manifest. The theme loader (collectCategoryCandidates)
// indexes ct.Manifest[i] directly, so ranges must be 0-based to align
// with the manifest array.
func buildRanges(manifest []manifestEntry) map[string]asset.PartRange {
	first := map[string]int{}
	last := map[string]int{}
	for i, entry := range manifest {
		cat := entry.Name
		if cat == "" {
			cat = "misc"
		}
		if _, ok := first[cat]; !ok {
			first[cat] = i
		}
		last[cat] = i
	}
	ranges := make(map[string]asset.PartRange)
	for cat := range first {
		ranges[cat] = asset.PartRange{First: first[cat], Last: last[cat]}
	}
	return ranges
}
