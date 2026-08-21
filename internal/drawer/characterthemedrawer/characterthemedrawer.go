// Package characterthemedrawer owns the character (layered portrait)
// theme assets and the layer-0 background drawer for character themes.
// It loads portrait themes from the embedded assets/character tree
// (ren.json + ren/*.webp), assembles a random portrait per request, and
// draws it as a nested-<svg> Layer.
//
// This package merged the former internal/theme package's character data
// types (Character, CharacterLayer, CharacterPart, ComposedPortrait),
// the Assemble/LoadCharacter logic, and the background-drawing half of
// composeCharacterSVG.
package characterthemedrawer

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"image"
	"io/fs"
	"math/rand"
	"path"
	"strings"

	"github.com/miaoledor/lolicount/assets"
	"github.com/miaoledor/lolicount/internal/drawer"
	"github.com/miaoledor/lolicount/internal/utils"
)

// supportedExts mirrors cardthemedrawer: accepted layer extensions.
var supportedExts = map[string]string{
	".gif":  "image/gif",
	".png":  "image/png",
	".webp": "image/webp",
}

// CharacterCanvasW/H are the original PSD canvas dimensions used to place
// character layers. Layers carry absolute left/top within this canvas.
// These match the reference 莲 PSD (504 x 925).
const (
	CharacterCanvasW = 504
	CharacterCanvasH = 925
)

// Character is a layered portrait theme (M9): a PSD split into
// transparent webp layers described by a JSON manifest. Each request
// randomly picks one layer from each of five part categories and
// overlays them at absolute coordinates to compose a full portrait.
type Character struct {
	// Layers is the full manifest, 1-based by convention (index 0 is the
	// "汗"/sweat layer, skipped). Indices 71-79 are PSD group labels,
	// unused by assembly.
	Layers []CharacterLayer
	// Parts is the pre-decoded layer data URI for each layer_id, keyed by
	// layer_id. Loaded once at registry construction so per-request
	// assembly only does memory lookups + random selection, no I/O.
	Parts map[int]CharacterPart
}

// CharacterLayer is one entry in ren.json: the absolute placement of a
// layer in the original PSD canvas.
type CharacterLayer struct {
	Name          string `json:"name"`
	Left          int    `json:"left"`
	Top           int    `json:"top"`
	Width         int    `json:"width"`
	Height        int    `json:"height"`
	Visible       int    `json:"visible"`
	LayerID       int    `json:"layer_id"`
	GroupLayerID  int    `json:"group_layer_id"`
}

// CharacterPart is a decoded layer image ready to overlay.
type CharacterPart struct {
	Left   int
	Top    int
	Width  int
	Height int
	Data   string
}

// partRange is a 1-based closed interval [First, Last] of layer indices
// in ren.json that belong to one part category.
type partRange struct {
	First, Last int
}

// Part categories by index range in ren.json (1-based, closed):
//
//	- brow  1-18   eyebrows
//	- eye   19-36  eyes
//	- mouth 37-56  mouths
//	- face  57-62  cheeks/blush (G頬 = no-blush base)
//	- lass  63-70  clothing/body
var characterRanges = map[string]partRange{
	"brow":  {1, 18},
	"eye":   {19, 36},
	"mouth": {37, 56},
	"face":  {57, 62},
	"lass":  {63, 70},
}

// assembly order for the z-stack (bottom -> top). Mirrors Loli.vue's
// <img> order: body(lass) -> eye -> brow -> mouth -> face.
var characterStack = []string{"lass", "eye", "brow", "mouth", "face"}

// ComposedPortrait is one randomly assembled portrait: the chosen parts
// and their shared bounding box.
type ComposedPortrait struct {
	Parts []CharacterPart
	BBox  struct {
		Left, Top, Width, Height int
	}
}

// Assemble randomly picks one layer from each part category and returns
// the composed portrait. It is pure memory work (no I/O).
func (c *Character) Assemble(r *rand.Rand) (*ComposedPortrait, error) {
	if c == nil || len(c.Layers) == 0 {
		return nil, fmt.Errorf("character: no layers")
	}
	chosen := make([]CharacterPart, 0, len(characterStack))
	for _, cat := range characterStack {
		rng, ok := characterRanges[cat]
		if !ok {
			continue
		}
		layer, err := c.pickLayer(r, rng)
		if err != nil {
			return nil, fmt.Errorf("character %s: %w", cat, err)
		}
		part, ok := c.Parts[layer.LayerID]
		if !ok {
			return nil, fmt.Errorf("character %s: layer_id %d not decoded", cat, layer.LayerID)
		}
		chosen = append(chosen, part)
	}

	p := &ComposedPortrait{Parts: chosen}
	left := chosen[0].Left
	top := chosen[0].Top
	right := chosen[0].Left + chosen[0].Width
	bottom := chosen[0].Top + chosen[0].Height
	for _, q := range chosen[1:] {
		if q.Left < left {
			left = q.Left
		}
		if q.Top < top {
			top = q.Top
		}
		if q.Left+q.Width > right {
			right = q.Left + q.Width
		}
		if q.Top+q.Height > bottom {
			bottom = q.Top + q.Height
		}
	}
	p.BBox.Left = left
	p.BBox.Top = top
	p.BBox.Width = right - left
	p.BBox.Height = bottom - top
	return p, nil
}

// pickLayer selects a random layer index in [rng.First, rng.Last]
// (1-based, closed) and returns the manifest entry.
func (c *Character) pickLayer(r *rand.Rand, rng partRange) (CharacterLayer, error) {
	if rng.First < 0 || rng.Last >= len(c.Layers) || rng.First > rng.Last {
		return CharacterLayer{}, fmt.Errorf("range [%d,%d] out of bounds (layers=%d)", rng.First, rng.Last, len(c.Layers))
	}
	idx := rng.First
	if rng.Last > rng.First {
		idx = rng.First + utils.RandomInt(r, rng.Last-rng.First+1)
	}
	return c.Layers[idx], nil
}

// Draw renders an assembled character portrait as the layer-0 background.
// Each portrait part is drawn at its ORIGINAL absolute left/top with its
// ORIGINAL width/height inside the PSD canvas. Scaling is applied to the
// whole canvas at once via an SVG viewBox -> viewport mapping (nested
// <svg>), NOT per-layer, so sub-pixel precision keeps layers aligned.
// Each layer is a data URI <image> (AGENTS.md Iron Rule 2).
func Draw(portrait *ComposedPortrait, scale float64) drawer.Layer {
	display := utils.DisplaySize(scale)
	imgW, imgH := utils.ScaledCanvasDims(CharacterCanvasW, CharacterCanvasH, display)

	var b strings.Builder
	fmt.Fprintf(&b, `  <svg x="0" y="0" width="%d" height="%d" viewBox="0 0 %d %d" preserveAspectRatio="none">`+"\n",
		imgW, imgH, CharacterCanvasW, CharacterCanvasH)
	for _, part := range portrait.Parts {
		fmt.Fprintf(&b, `    <image x="%d" y="%d" width="%d" height="%d" xlink:href="%s" />`+"\n",
			part.Left, part.Top, part.Width, part.Height, part.Data)
	}
	b.WriteString("  </svg>\n")
	return drawer.Layer{Fragment: b.String(), Width: imgW, Height: imgH}
}

// LoadCharacter reads ren.json + the ren/ layer directory from fsys
// (rooted at the theme dir) and pre-decodes every referenced layer into
// a data URI. The returned Character is ready for Assemble with no
// further I/O.
func LoadCharacter(fsys fs.FS, themeDir string) (*Character, error) {
	manifestPath := path.Join(themeDir, "ren.json")
	raw, err := fs.ReadFile(fsys, manifestPath)
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", manifestPath, err)
	}
	var layers []CharacterLayer
	if err := json.Unmarshal(raw, &layers); err != nil {
		return nil, fmt.Errorf("parse %s: %w", manifestPath, err)
	}
	if len(layers) == 0 {
		return nil, fmt.Errorf("%s: empty manifest", manifestPath)
	}

	renDir := path.Join(themeDir, "ren")
	parts := make(map[int]CharacterPart, len(layers))
	for _, l := range layers {
		if l.LayerID == 0 {
			continue
		}
		if _, dup := parts[l.LayerID]; dup {
			continue
		}
		data, imgW, imgH, err := readLayerDataURI(fsys, renDir, l.LayerID)
		if err != nil {
			// A manifest may reference layers without a shipped image
			// (e.g. group labels). Skip missing files so a partial set
			// still loads; assembly will error only if a selected part
			// is absent.
			continue
		}
		// Use the image file's ACTUAL pixel dimensions, not ren.json's
		// width/height: the webp files are grid-padded and larger than
		// the manifest content size.
		parts[l.LayerID] = CharacterPart{
			Left:   l.Left,
			Top:    l.Top,
			Width:  imgW,
			Height: imgH,
			Data:   data,
		}
	}
	if len(parts) == 0 {
		return nil, fmt.Errorf("%s: no layer images decoded", themeDir)
	}
	return &Character{Layers: layers, Parts: parts}, nil
}

// readLayerDataURI loads /ren/<layer_id>.<ext> and returns its data URI
// plus the image's ACTUAL pixel dimensions.
func readLayerDataURI(fsys fs.FS, renDir string, layerID int) (data string, w int, h int, err error) {
	for _, ext := range []string{".webp", ".png", ".gif"} {
		p := path.Join(renDir, fmt.Sprintf("%d%s", layerID, ext))
		raw, e := fs.ReadFile(fsys, p)
		if e != nil {
			continue
		}
		m, ok := supportedExts[strings.ToLower(ext)]
		if !ok {
			continue
		}
		cfg, _, e := image.DecodeConfig(bytes.NewReader(raw))
		if e != nil {
			return "", 0, 0, fmt.Errorf("layer %d: decode config: %w", layerID, e)
		}
		uri := "data:" + m + ";base64," + base64.StdEncoding.EncodeToString(raw)
		return uri, cfg.Width, cfg.Height, nil
	}
	return "", 0, 0, fmt.Errorf("layer %d: no image found", layerID)
}

// Registry resolves a character theme name to its Character.
type Registry interface {
	Get(name string) (*Character, bool)
	List() []string
}

// builtinCharRegistry loads character themes from assets/character.
type builtinCharRegistry struct {
	themes map[string]*Character
}

// NewBuiltinRegistry scans the embedded assets/character directory and
// loads every valid portrait theme into memory.
func NewBuiltinRegistry() (Registry, []error) {
	reg := &builtinCharRegistry{themes: make(map[string]*Character)}
	var errs []error

	sub, err := fs.Sub(assets.FS, "character")
	if err != nil {
		return reg, []error{fmt.Errorf("characterthemedrawer: open embedded character: %w", err)}
	}
	entries, err := fs.ReadDir(sub, ".")
	if err != nil {
		return reg, []error{fmt.Errorf("characterthemedrawer: read character: %w", err)}
	}
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		name := e.Name()
		ch, err := LoadCharacter(sub, name)
		if err != nil {
			errs = append(errs, fmt.Errorf("character %s: %w", name, err))
			continue
		}
		reg.themes[name] = ch
	}
	return reg, errs
}

func (r *builtinCharRegistry) Get(name string) (*Character, bool) {
	ch, ok := r.themes[name]
	return ch, ok
}

func (r *builtinCharRegistry) List() []string {
	out := make([]string, 0, len(r.themes))
	for k := range r.themes {
		out = append(out, k)
	}
	return out
}
