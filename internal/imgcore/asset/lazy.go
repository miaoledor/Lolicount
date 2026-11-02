// Package asset lazy.go implements the one-image-per-slot memory
// strategy: instead of decoding every candidate image of every theme at
// startup, it builds a lightweight catalog (file paths + placement +
// header dimensions only) and assembles each theme with exactly ONE
// loaded candidate per slot. Candidates are swapped in on demand at
// request time via DecodeImage (see composer's PrepareRender).
package asset

import (
	"bytes"
	"fmt"
	"image"
	"io/fs"
	"path"
	"sort"

	"github.com/miaoledor/lolicount/internal/imgcore"
	"github.com/miaoledor/lolicount/internal/imgcore/render"
	"github.com/miaoledor/lolicount/internal/imgcore/theme"
)

// Candidate is one selectable image of a slot, described by metadata
// only — the pixels stay on disk (embed.FS) until this exact candidate
// is swapped in.
type Candidate struct {
	RelPath string // relative to the theme fs root, e.g. "lian-ren/ren/123.webp"
	Mime    string
	X, Y    int // placement on the theme canvas (frames: 0,0)
	Width   int
	Height  int
}

// Slot is one selectable part position of a theme (a frame-theme's frame
// picker, or a character theme's part category). Loaded is the index of
// the candidate currently resident in the assembled theme.
type Slot struct {
	Candidates []Candidate
	Loaded     int
}

// ThemeCatalog is the full metadata of a theme: every slot and every
// candidate, with no image bytes. Character themes carry the GroupLayer
// layout derived from config/display; frame themes carry the first
// frame's canvas.
type ThemeCatalog struct {
	Name       string
	FrameTheme bool
	Slots      []*Slot

	// Frame-theme canvas (first frame dims) or character layout.
	CanvasW, CanvasH               int
	OutW, OutH, VbX, VbY, VbW, VbH int
	Display                        *theme.DisplayConfig

	// Variants is the product of candidate counts across slots (frame
	// themes: the frame count) — the number of distinct renders the theme
	// can produce.
	Variants int
}

// LoadThemeCatalogs scans the theme tree and returns a catalog per valid
// theme. It reads every image file once to parse its header (dimensions)
// but never decodes pixels or builds base64 data, so startup cost stays
// low and no image bytes are retained.
func LoadThemeCatalogs(fsys fs.FS) (map[string]*ThemeCatalog, []error) {
	entries, err := fs.ReadDir(fsys, ".")
	if err != nil {
		return nil, []error{fmt.Errorf("theme catalog: read root: %w", err)}
	}
	cats := make(map[string]*ThemeCatalog)
	var errs []error
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		name := e.Name()
		cat, err := loadThemeCatalog(fsys, name)
		if err != nil {
			errs = append(errs, err)
			continue
		}
		cats[name] = cat
	}
	return cats, errs
}

// loadThemeCatalog dispatches by ren.json presence.
func loadThemeCatalog(fsys fs.FS, name string) (*ThemeCatalog, error) {
	if isManifestTheme(fsys, name) {
		return loadCharacterCatalog(fsys, name)
	}
	return loadFrameCatalog(fsys, name)
}

// loadFrameCatalog builds a single-slot catalog from the directory's
// numbered frame files.
func loadFrameCatalog(fsys fs.FS, name string) (*ThemeCatalog, error) {
	entries, err := fs.ReadDir(fsys, name)
	if err != nil {
		return nil, fmt.Errorf("theme %s: read dir: %w", name, err)
	}
	type indexed struct {
		idx  int
		cand Candidate
	}
	var frames []indexed
	for _, e := range entries {
		base := e.Name()
		idx := FrameIndexFromName(base)
		if idx < 0 {
			continue
		}
		ext := pathExt(base)
		mime, ok := SupportedExts[ext]
		if !ok {
			continue
		}
		rel := path.Join(name, base)
		w, h, err := headerDims(fsys, rel)
		if err != nil {
			return nil, fmt.Errorf("theme %s: %w", name, err)
		}
		frames = append(frames, indexed{idx: idx, cand: Candidate{
			RelPath: rel, Mime: mime, Width: w, Height: h,
		}})
	}
	if len(frames) == 0 {
		return nil, fmt.Errorf("theme %s: no frame images found", name)
	}
	sort.Slice(frames, func(i, j int) bool { return frames[i].idx < frames[j].idx })

	// Canvas = the maximum dimensions across all frames: the
	// resident assembly holds one frame at a time, and without this floor
	// the canvas would shrink/grow as swaps rotate through frames of
	// differing sizes, making the embedded image's aspect ratio jump
	// between requests.
	maxW, maxH := 0, 0
	for _, fr := range frames {
		if fr.cand.Width > maxW {
			maxW = fr.cand.Width
		}
		if fr.cand.Height > maxH {
			maxH = fr.cand.Height
		}
	}

	slot := &Slot{Candidates: make([]Candidate, len(frames))}
	for i, fr := range frames {
		slot.Candidates[i] = fr.cand
	}
	return &ThemeCatalog{
		Name:       name,
		FrameTheme: true,
		Slots:      []*Slot{slot},
		CanvasW:    maxW,
		CanvasH:    maxH,
		Variants:   len(frames),
	}, nil
}

// loadCharacterCatalog builds a per-category slot catalog from ren.json
// + config.json (+ display.json), preserving manifest range order and
// display layout.
func loadCharacterCatalog(fsys fs.FS, name string) (*ThemeCatalog, error) {
	dir := name
	manifest, cfg, err := readCharacterManifest(fsys, dir)
	if err != nil {
		return nil, fmt.Errorf("theme %s: %w", name, err)
	}
	display := readCharacterDisplay(fsys, dir)

	renDir := path.Join(dir, "ren")
	type rangeKey struct {
		rng PartRange
	}
	var sortedRanges []rangeKey
	for _, rng := range cfg.Ranges {
		sortedRanges = append(sortedRanges, rangeKey{rng: rng})
	}
	sort.Slice(sortedRanges, func(i, j int) bool {
		return sortedRanges[i].rng.First < sortedRanges[j].rng.First
	})

	var slots []*Slot
	for _, rk := range sortedRanges {
		if rk.rng.First < 0 || rk.rng.Last >= len(manifest) || rk.rng.First > rk.rng.Last {
			continue
		}
		var cands []Candidate
		for i := rk.rng.First; i <= rk.rng.Last; i++ {
			layer := manifest[i]
			if layer.LayerID == 0 {
				continue
			}
			rel, mime, err := FindImageFile(fsys, path.Join(renDir, fmt.Sprintf("%d", layer.LayerID)))
			if err != nil {
				continue
			}
			w, h, err := headerDims(fsys, rel)
			if err != nil {
				continue
			}
			w, h = characterLayerDimensions(layer, w, h)
			cands = append(cands, Candidate{
				RelPath: rel, Mime: mime,
				X: layer.Left, Y: layer.Top, Width: w, Height: h,
			})
		}
		if len(cands) > 0 {
			slots = append(slots, &Slot{Candidates: cands})
		}
	}
	if len(slots) == 0 {
		return nil, fmt.Errorf("theme %s: no part slots assembled", name)
	}

	canvasW, canvasH := cfg.CanvasW, cfg.CanvasH
	outW, outH := canvasW, canvasH
	vbX, vbY, vbW, vbH := 0, 0, canvasW, canvasH
	if display != nil && display.Size > 0 {
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
	}

	variants := 1
	for _, s := range slots {
		variants *= len(s.Candidates)
	}
	return &ThemeCatalog{
		Name:    name,
		Slots:   slots,
		CanvasW: canvasW, CanvasH: canvasH,
		OutW: outW, OutH: outH, VbX: vbX, VbY: vbY, VbW: vbW, VbH: vbH,
		Display:  display,
		Variants: variants,
	}, nil
}

// AssembleTheme builds a renderable *theme.Theme holding exactly ONE
// loaded candidate per slot (picked deterministically from the prng, so
// a given catalog + seed assemble identically). The caller retains the
// catalog; swapping slots later mutates the returned theme through the
// registry's swap path.
func AssembleTheme(fsys fs.FS, cat *ThemeCatalog, pick func(n int) int) (*theme.Theme, error) {
	if cat.FrameTheme {
		slot := cat.Slots[0]
		idx := pick(len(slot.Candidates))
		if idx < 0 || idx >= len(slot.Candidates) {
			idx = 0
		}
		cand := slot.Candidates[idx]
		decoded, err := DecodeImage(fsys, cand.RelPath, cand.Mime)
		if err != nil {
			return nil, fmt.Errorf("theme %s: %w", cat.Name, err)
		}
		slot.Loaded = idx
		layer := render.ImageLayer{
			Src:       decoded.Data,
			Width:     decoded.Width,
			Height:    decoded.Height,
			Transform: imgcore.DefaultTransform(),
		}
		var out imgcore.Layer
		if len(slot.Candidates) == 1 {
			out = &layer
		} else {
			frameDims := make([]render.FrameDim, len(slot.Candidates))
			for i, c := range slot.Candidates {
				frameDims[i] = render.FrameDim{W: c.Width, H: c.Height}
			}
			out = &render.RandomPickLayer{
				Category:  cat.Name,
				Options:   []render.ImageOption{{ImageLayer: layer, Weight: 1}},
				FrameDims: frameDims,
				Transform: imgcore.DefaultTransform(),
			}
		}
		return &theme.Theme{
			Name:   cat.Name,
			Canvas: theme.Canvas{Width: cat.CanvasW, Height: cat.CanvasH},
			BgW:    cat.CanvasW,
			BgH:    cat.CanvasH,
			Layers: []imgcore.Layer{out},
		}, nil
	}

	groupParts := make([]render.GroupPart, len(cat.Slots))
	for si, slot := range cat.Slots {
		idx := pick(len(slot.Candidates))
		if idx < 0 || idx >= len(slot.Candidates) {
			idx = 0
		}
		cand := slot.Candidates[idx]
		decoded, err := DecodeImage(fsys, cand.RelPath, cand.Mime)
		if err != nil {
			return nil, fmt.Errorf("theme %s: %w", cat.Name, err)
		}
		slot.Loaded = idx
		// Candidates MUST stay empty: GroupPart.Render treats non-empty
		// Candidates as a per-request random pick, which would bypass the
		// swap-driven one-part-per-request strategy.
		groupParts[si] = render.GroupPart{
			Src:    decoded.Data,
			X:      cand.X,
			Y:      cand.Y,
			Width:  cand.Width,
			Height: cand.Height,
		}
	}
	groupLayer := &render.GroupLayer{
		Parts: groupParts,
		OutW:  cat.OutW,
		OutH:  cat.OutH,
		VbX:   cat.VbX,
		VbY:   cat.VbY,
		VbW:   cat.VbW,
		VbH:   cat.VbH,
	}
	return &theme.Theme{
		Name:    cat.Name,
		Canvas:  theme.Canvas{Width: cat.OutW, Height: cat.OutH},
		BgW:     cat.OutW,
		BgH:     cat.OutH,
		Display: cat.Display,
		Layers:  []imgcore.Layer{groupLayer},
	}, nil
}

// headerDims reads an image file and returns its pixel dimensions from
// the header only — no pixel decode, no retained bytes.
func headerDims(fsys fs.FS, relPath string) (int, int, error) {
	raw, err := fs.ReadFile(fsys, relPath)
	if err != nil {
		return 0, 0, fmt.Errorf("read %s: %w", relPath, err)
	}
	cfg, _, err := image.DecodeConfig(bytes.NewReader(raw))
	if err != nil {
		return 0, 0, fmt.Errorf("decode config %s: %w", relPath, err)
	}
	return cfg.Width, cfg.Height, nil
}
