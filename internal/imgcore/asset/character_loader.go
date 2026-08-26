package asset

import (
	"encoding/json"
	"math/rand"
	"fmt"
	"io/fs"
	"path"
	"sort"

	"github.com/miaoledor/lolicount/assets"
	"github.com/miaoledor/lolicount/internal/imgcore"
	"github.com/miaoledor/lolicount/internal/imgcore/render"
	"github.com/miaoledor/lolicount/internal/imgcore/theme"
)

// CharacterConfig describes a character theme's PSD layout: canvas
// dimensions and 1-based closed index ranges for each part category.
// Migrated from characterthemedrawer.CharacterConfig.
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
// a layer in the original PSD canvas. Migrated from
// characterthemedrawer.CharacterLayer.
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

// CharacterTheme is a layered portrait theme loaded from
// assets/character. It holds the manifest, pre-decoded image data, and
// config. The registry converts it to a theme.Theme with RandomPickLayer
// layers on demand.
type CharacterTheme struct {
	Name     string
	Manifest []CharacterManifest
	Config   *CharacterConfig
	Display  *theme.DisplayConfig
	Parts    map[int]render.ImageLayer // layer_id -> decoded image layer
}

// CharacterRegistry resolves a character theme name to its CharacterTheme.
type CharacterRegistry interface {
	Get(name string) (*CharacterTheme, bool)
	List() []string
}

// builtinCharacterRegistry loads character themes from assets/character.
type builtinCharacterRegistry struct {
	themes map[string]*CharacterTheme
}

// NewBuiltinCharacterRegistry scans the embedded assets/character
// directory and loads every valid portrait theme into memory.
func NewBuiltinCharacterRegistry() (CharacterRegistry, []error) {
	reg := &builtinCharacterRegistry{themes: make(map[string]*CharacterTheme)}
	var errs []error

	sub, err := fs.Sub(assets.FS, "character")
	if err != nil {
		return reg, []error{fmt.Errorf("character loader: open embedded character: %w", err)}
	}
	entries, err := fs.ReadDir(sub, ".")
	if err != nil {
		return reg, []error{fmt.Errorf("character loader: read character: %w", err)}
	}
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		name := e.Name()
		ch, err := LoadCharacterTheme(sub, name)
		if err != nil {
			errs = append(errs, fmt.Errorf("character %s: %w", name, err))
			continue
		}
		reg.themes[name] = ch
	}
	return reg, errs
}

// LoadCharacterTheme reads ren.json + config.json + display.json +
// the ren/ layer directory from fsys (rooted at the theme dir) and
// pre-decodes every referenced layer into a data URI.
func LoadCharacterTheme(fsys fs.FS, themeDir string) (*CharacterTheme, error) {
	manifestPath := path.Join(themeDir, "ren.json")
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
				X:     imgcore.FixedRange(float64(l.Left)),
				Y:     imgcore.FixedRange(float64(l.Top)),
				Scale: imgcore.FixedRange(1),
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

// Get returns the theme for name, or false if not registered.
func (r *builtinCharacterRegistry) Get(name string) (*CharacterTheme, bool) {
	ch, ok := r.themes[name]
	return ch, ok
}

// List returns registered theme names sorted for stable output.
func (r *builtinCharacterRegistry) List() []string {
	out := make([]string, 0, len(r.themes))
	for k := range r.themes {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

// CharacterPart is a decoded layer image ready to overlay, with its
// absolute placement coordinates from the PSD manifest. Used by the
// server bridge during the transitional period.
type CharacterPart struct {
	Left   int
	Top    int
	Width  int
	Height int
	Data   string
}

// CharacterThemeToTheme converts a CharacterTheme into a *theme.Theme
// with a GroupLayer (nested <svg viewBox>) for PSD coordinate mapping
// plus the display config. This mirrors the old server bridge's
// buildCharacterThemeLayers but runs at load time so the registry holds
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
	// Assemble picks one layer per category using the config ranges.
	var groupParts []render.GroupPart
	for _, cat := range []string{"lass", "eye", "brow", "mouth", "face"} {
		rng, ok := ct.Config.Ranges[cat]
		if !ok {
			continue
		}
		layer, err := pickManifestLayer(ct, rng)
		if err != nil {
			continue
		}
		img, ok := ct.Parts[layer.LayerID]
		if !ok {
			continue
		}
		groupParts = append(groupParts, render.GroupPart{
			Src:    img.Src,
			X:      layer.Left,
			Y:      layer.Top,
			Width:  img.Width,
			Height: img.Height,
		})
	}

	if len(groupParts) == 0 {
		return nil, fmt.Errorf("character theme %s: no parts assembled", ct.Name)
	}

	// Compute output dimensions from display config or fall back to
	// canvas dims.
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

// pickManifestLayer randomly selects a layer index in [First, Last]
// from the manifest. Exposed for CharacterThemeToTheme.
func pickManifestLayer(ct *CharacterTheme, rng PartRange) (CharacterManifest, error) {
	if rng.First < 0 || rng.Last >= len(ct.Manifest) || rng.First > rng.Last {
		return CharacterManifest{}, fmt.Errorf("range [%d,%d] out of bounds (manifest=%d)", rng.First, rng.Last, len(ct.Manifest))
	}
	idx := rng.First
	if rng.Last > rng.First {
		idx = rng.First + rand.Intn(rng.Last-rng.First+1)
	}
	return ct.Manifest[idx], nil
}
