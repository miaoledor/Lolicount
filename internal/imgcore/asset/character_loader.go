package asset

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"path"
	"sort"

	"github.com/miaoledor/lolicount/internal/imgcore"
	"github.com/miaoledor/lolicount/internal/imgcore/render"
	"github.com/miaoledor/lolicount/internal/imgcore/theme"
)

// CharacterConfig describes a character theme's PSD layout: canvas
// dimensions and 1-based closed index ranges for each part category.
type CharacterConfig struct {
	CanvasW int                  `json:"canvasW"`
	CanvasH int                  `json:"canvasH"`
	Ranges  map[string]PartRange `json:"ranges"`
}

// PartRange is a 1-based closed interval [First, Last] of layer indices
// belonging to one part category.
type PartRange struct {
	First int `json:"first"`
	Last  int `json:"last"`
}

// CharacterManifest is one entry in ren.json: the absolute placement of
// a layer in the original PSD canvas.
type CharacterManifest struct {
	Name         string `json:"name"`
	Left         int    `json:"left"`
	Top          int    `json:"top"`
	Width        int    `json:"width"`
	Height       int    `json:"height"`
	Visible      int    `json:"visible"`
	LayerID      int    `json:"layer_id"`
	GroupLayerID int    `json:"group_layer_id"`
}

// CharacterTheme is a layered portrait theme loaded from a ren.json
// manifest inside assets/theme/<name>/. It holds the manifest,
// pre-decoded image data, and config. LoadThemes converts it to a
// *theme.Theme with GroupLayer parts on demand.
type CharacterTheme struct {
	Name     string
	Manifest []CharacterManifest
	Config   *CharacterConfig
	Display  *theme.DisplayConfig
	Parts    map[int]render.ImageLayer
}

// LoadCharacterTheme reads ren.json + config.json + display.json +
// the ren/ layer directory from fsys (rooted at the theme dir) and
// pre-decodes every referenced layer into a data URI. A theme directory
// is dispatched here when it contains a ren.json manifest.
func LoadCharacterTheme(fsys fs.FS, themeDir string) (*CharacterTheme, error) {
	manifestPath := path.Join(themeDir, ManifestName)
	raw, err := fs.ReadFile(fsys, manifestPath)
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", manifestPath, err)
	}
	var manifest []CharacterManifest
	if err := json.Unmarshal(raw, &manifest); err != nil {
		return nil, fmt.Errorf("parse %s: %w", manifestPath, err)
	}
	if len(manifest) == 0 {
		return nil, fmt.Errorf("%s: empty manifest", manifestPath)
	}

	renDir := path.Join(themeDir, "ren")
	parts := make(map[int]render.ImageLayer, len(manifest))
	for _, l := range manifest {
		if l.LayerID == 0 {
			continue
		}
		if _, dup := parts[l.LayerID]; dup {
			continue
		}
		imgPath, mime, err := FindImageFile(fsys, path.Join(renDir, fmt.Sprintf("%d", l.LayerID)))
		if err != nil {
			continue
		}
		decoded, err := DecodeImage(fsys, imgPath, mime)
		if err != nil {
			continue
		}
		parts[l.LayerID] = render.ImageLayer{
			Src:       decoded.Data,
			Width:     decoded.Width,
			Height:    decoded.Height,
			Transform: imgcore.Transform{
				X:        imgcore.FixedRange(float64(l.Left)),
				Y:        imgcore.FixedRange(float64(l.Top)),
				Scale:    imgcore.FixedRange(1),
				Rotation: imgcore.FixedRange(0),
			},
		}
	}
	if len(parts) == 0 {
		return nil, fmt.Errorf("%s: no layer images decoded", themeDir)
	}

	ch := &CharacterTheme{Name: themeDir, Manifest: manifest, Parts: parts}

	configPath := path.Join(themeDir, "config.json")
	if raw, err := fs.ReadFile(fsys, configPath); err == nil {
		var cfg CharacterConfig
		if err := json.Unmarshal(raw, &cfg); err != nil {
			return nil, fmt.Errorf("parse %s: %w", configPath, err)
		}
		if cfg.CanvasW > 0 && cfg.CanvasH > 0 && len(cfg.Ranges) > 0 {
			ch.Config = &cfg
		}
	}

	displayPath := path.Join(themeDir, "display.json")
	if raw, err := fs.ReadFile(fsys, displayPath); err == nil {
		var dp theme.DisplayConfig
		if err := json.Unmarshal(raw, &dp); err != nil {
			return nil, fmt.Errorf("parse %s: %w", displayPath, err)
		}
		if dp.Size > 0 {
			ch.Display = &dp
		}
	}
	return ch, nil
}

// CharacterThemeToTheme converts a CharacterTheme into a *theme.Theme
// with a GroupLayer (nested <svg viewBox>) for PSD coordinate mapping
// plus the display config. Runs at load time so the registry holds
// a ready-to-render *theme.Theme.
func CharacterThemeToTheme(ct *CharacterTheme) (*theme.Theme, error) {
	if ct == nil || ct.Config == nil {
		return nil, fmt.Errorf("character theme %s: missing config", ct.Name)
	}
	if len(ct.Parts) == 0 {
		return nil, fmt.Errorf("character theme %s: no decoded parts", ct.Name)
	}

	canvasW := ct.Config.CanvasW
	canvasH := ct.Config.CanvasH

	// Build GroupLayer parts from the manifest + decoded images.
	// Each range (category) becomes one GroupPart with all its
	// candidates, so the PRNG picks one candidate per request at render
	// time. Ranges are sorted by First index to preserve Z-order
	// (bottom-to-top) regardless of map iteration order. This supports
	// both the traditional fixed categories (lass/eye/brow/mouth/face)
	// and editor-exported themes where each layer name is its own
	// category.
	type rangeKey struct {
		name string
		rng  PartRange
	}
	var sortedRanges []rangeKey
	for name, rng := range ct.Config.Ranges {
		sortedRanges = append(sortedRanges, rangeKey{name: name, rng: rng})
	}
	sort.Slice(sortedRanges, func(i, j int) bool {
		return sortedRanges[i].rng.First < sortedRanges[j].rng.First
	})

	var groupParts []render.GroupPart
	for _, rk := range sortedRanges {
		candidates := collectCategoryCandidates(ct, rk.rng)
		if len(candidates) == 0 {
			continue
		}
		first := candidates[0]
		groupParts = append(groupParts, render.GroupPart{
			Src:        first.Src,
			X:          first.X,
			Y:          first.Y,
			Width:      first.Width,
			Height:     first.Height,
			Candidates: candidates,
		})
	}

	if len(groupParts) == 0 {
		return nil, fmt.Errorf("character theme %s: no parts assembled", ct.Name)
	}

	outW, outH := canvasW, canvasH
	vbX, vbY, vbW, vbH := 0, 0, canvasW, canvasH

	if ct.Display != nil && ct.Display.Size > 0 {
		vbW, vbH = canvasW, canvasH
		if ct.Display.Crop != nil && ct.Display.Crop.Width > 0 && ct.Display.Crop.Height > 0 {
			vbW = ct.Display.Crop.Width
			vbH = ct.Display.Crop.Height
			vbX = ct.Display.Crop.Left
			vbY = ct.Display.Crop.Top
		}
		outH = ct.Display.Size
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

	return &theme.Theme{
		Name:    ct.Name,
		Canvas:  theme.Canvas{Width: outW, Height: outH},
		BgW:     outW,
		BgH:     outH,
		Display: ct.Display,
		Layers:  []imgcore.Layer{groupLayer},
	}, nil
}

// collectCategoryCandidates gathers all decoded image candidates for one
// category range. Each candidate carries its placement from the manifest
// and its decoded image data URI. At render time the PRNG picks one.
func collectCategoryCandidates(ct *CharacterTheme, rng PartRange) []render.GroupCandidate {
	if rng.First < 0 || rng.Last >= len(ct.Manifest) || rng.First > rng.Last {
		return nil
	}
	var out []render.GroupCandidate
	for i := rng.First; i <= rng.Last; i++ {
		layer := ct.Manifest[i]
		img, ok := ct.Parts[layer.LayerID]
		if !ok {
			continue
		}
		out = append(out, render.GroupCandidate{
			Src:    img.Src,
			X:      layer.Left,
			Y:      layer.Top,
			Width:  img.Width,
			Height: img.Height,
		})
	}
	return out
}
