package theme

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/fs"
	"math/rand"
	"path"
	"strings"
)

// Character is a layered portrait theme (M9): a PSD split into
// transparent webp layers described by a JSON manifest. Each request
// randomly picks one layer from each of five part categories and
// overlays them at absolute coordinates to compose a full portrait.
// Ported from kungal-forum's setting-panel Loli (getLoli.ts).
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

// CharacterPart is a decoded layer image ready to overlay: its placement
// and a data URI of its bytes.
type CharacterPart struct {
	Left   int
	Top    int
	Width  int
	Height int
	Data   string
}

// partRange is a 1-based closed interval [First, Last] of layer indices
// in ren.json that belong to one part category. Mirrors getLoli.ts.
type partRange struct {
	First, Last int
}

// Part categories by index range in ren.json (1-based, closed):
//   - brow  1-18   eyebrows
//   - eye   19-36  eyes
//   - mouth 37-56  mouths
//   - face  57-62  cheeks/blush (G頬 = no-blush base)
//   - lass  63-70  clothing/body
var characterRanges = map[string]partRange{
	"brow":  {1, 18},
	"eye":   {19, 36},
	"mouth": {37, 56},
	"face":  {57, 62},
	"lass":  {63, 70},
}

// assembly order for the z-stack (bottom → top). Mirrors Loli.vue's
// <img> order: body(lass) → eye → brow → mouth → face.
var characterStack = []string{"lass", "eye", "brow", "mouth", "face"}

// ComposedPortrait is one randomly assembled portrait: the chosen parts
// and their shared bounding box. The renderer overlays each part's
// data URI at Left/Top.
type ComposedPortrait struct {
	Parts []CharacterPart
	BBox  struct {
		Left, Top, Width, Height int
	}
}

// Assemble randomly picks one layer from each part category and returns
// the composed portrait. It is pure memory work (no I/O): the parts were
// decoded at load time. Returns an error only if the manifest is
// malformed (a category range out of bounds).
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
		idx = rng.First + randomInt(r, rng.Last-rng.First+1)
	}
	return c.Layers[idx], nil
}

// CanvasSize is the original PSD canvas dimensions used to place
// character layers. Layers carry absolute left/top within this canvas.
// These match the reference 莲 PSD (504 x 925).
const (
	CharacterCanvasW = 504
	CharacterCanvasH = 925
)

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
		data, err := readLayerDataURI(fsys, renDir, l.LayerID)
		if err != nil {
			// A manifest may reference layers without a shipped image
			// (e.g. group labels). Skip missing files so a partial set
			// still loads; assembly will error only if a selected part
			// is absent.
			continue
		}
		parts[l.LayerID] = CharacterPart{
			Left:   l.Left,
			Top:    l.Top,
			Width:  l.Width,
			Height: l.Height,
			Data:   data,
		}
	}
	if len(parts) == 0 {
		return nil, fmt.Errorf("%s: no layer images decoded", themeDir)
	}
	return &Character{Layers: layers, Parts: parts}, nil
}

// readLayerDataURI loads /ren/<layer_id>.<ext> and returns its data URI.
func readLayerDataURI(fsys fs.FS, renDir string, layerID int) (string, error) {
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
		return "data:" + m + ";base64," + base64.StdEncoding.EncodeToString(raw), nil
	}
	return "", fmt.Errorf("layer %d: no image found", layerID)
}
