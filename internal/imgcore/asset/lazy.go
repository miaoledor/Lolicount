// Package asset lazy.go implements the one-image-per-slot memory
// strategy: instead of decoding every candidate image of every theme at
// startup, it builds a lightweight catalog (file paths + placement +
// header dimensions only) and assembles each theme with exactly ONE
// loaded candidate per slot. Candidates are swapped in on demand at
// request time via DecodeImage (see composer's PrepareRender).
package asset

import (
	"fmt"
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
	Sets       []*CatalogSet

	// Frame-theme canvas (first frame dims) or character layout.
	CanvasW, CanvasH               int
	OutW, OutH, VbX, VbY, VbW, VbH int
	Display                        *theme.DisplayConfig

	// Variants is the product of candidate counts across slots, or the sum
	// of those products for mutually exclusive sets. Frame themes use the
	// frame count.
	Variants int
}

// CatalogSet is one mutually exclusive outfit and its independently
// selectable parts. Variants across sets are added rather than multiplied.
type CatalogSet struct {
	Name                           string
	Slots                          []*Slot
	CanvasW, CanvasH               int
	OutW, OutH, VbX, VbY, VbW, VbH int
	Variants                       int
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
	if len(cfg.Sets) > 0 {
		cat := &ThemeCatalog{Name: name, Display: display}
		for _, configured := range cfg.Sets {
			set, err := loadCharacterSet(fsys, dir, manifest, configured, display)
			if err != nil {
				return nil, fmt.Errorf("theme %s set %s: %w", name, configured.Name, err)
			}
			cat.Sets = append(cat.Sets, set)
			cat.Variants += set.Variants
		}
		first := cat.Sets[0]
		cat.CanvasW, cat.CanvasH = first.CanvasW, first.CanvasH
		cat.OutW, cat.OutH = first.OutW, first.OutH
		cat.VbX, cat.VbY, cat.VbW, cat.VbH = first.VbX, first.VbY, first.VbW, first.VbH
		return cat, nil
	}

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

func loadCharacterSet(fsys fs.FS, dir string, manifest []CharacterManifest, configured CharacterSet, display *theme.DisplayConfig) (*CatalogSet, error) {
	type rangeKey struct{ rng PartRange }
	var sortedRanges []rangeKey
	for _, rng := range configured.Ranges {
		sortedRanges = append(sortedRanges, rangeKey{rng: rng})
	}
	sort.Slice(sortedRanges, func(i, j int) bool { return sortedRanges[i].rng.First < sortedRanges[j].rng.First })

	set := &CatalogSet{Name: configured.Name, CanvasW: configured.CanvasW, CanvasH: configured.CanvasH, Variants: 1}
	for _, rk := range sortedRanges {
		if rk.rng.First < 0 || rk.rng.Last >= len(manifest) || rk.rng.First > rk.rng.Last {
			return nil, fmt.Errorf("invalid range %+v", rk.rng)
		}
		var candidates []Candidate
		for i := rk.rng.First; i <= rk.rng.Last; i++ {
			layer := manifest[i]
			if layer.LayerID == 0 {
				continue
			}
			rel, mime, err := FindImageFile(fsys, path.Join(dir, "ren", fmt.Sprintf("%d", layer.LayerID)))
			if err != nil {
				return nil, err
			}
			w, h, err := headerDims(fsys, rel)
			if err != nil {
				return nil, err
			}
			candidates = append(candidates, Candidate{RelPath: rel, Mime: mime, X: layer.Left, Y: layer.Top, Width: w, Height: h})
		}
		if len(candidates) == 0 {
			return nil, fmt.Errorf("range %+v has no candidates", rk.rng)
		}
		set.Slots = append(set.Slots, &Slot{Candidates: candidates})
		set.Variants *= len(candidates)
	}
	set.OutW, set.OutH = set.CanvasW, set.CanvasH
	set.VbW, set.VbH = set.CanvasW, set.CanvasH
	if display != nil && display.Size > 0 {
		if display.Crop != nil && display.Crop.Width > 0 && display.Crop.Height > 0 {
			set.VbX, set.VbY = display.Crop.Left, display.Crop.Top
			set.VbW, set.VbH = display.Crop.Width, display.Crop.Height
		}
		set.OutH = display.Size
		set.OutW = int(float64(set.VbW) * float64(set.OutH) / float64(set.VbH))
		if set.OutW < 1 {
			set.OutW = 1
		}
	}
	return set, nil
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
	if len(cat.Sets) > 0 {
		setIndex := pick(len(cat.Sets))
		if setIndex < 0 || setIndex >= len(cat.Sets) {
			setIndex = 0
		}
		return assembleCharacterSlots(fsys, cat, cat.Sets[setIndex].Slots, cat.Sets[setIndex], pick)
	}

	return assembleCharacterSlots(fsys, cat, cat.Slots, nil, pick)
}

func assembleCharacterSlots(fsys fs.FS, cat *ThemeCatalog, slots []*Slot, set *CatalogSet, pick func(n int) int) (*theme.Theme, error) {
	groupParts := make([]render.GroupPart, len(slots))
	for si, slot := range slots {
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
	outW, outH, vbX, vbY, vbW, vbH := cat.OutW, cat.OutH, cat.VbX, cat.VbY, cat.VbW, cat.VbH
	if set != nil {
		outW, outH, vbX, vbY, vbW, vbH = set.OutW, set.OutH, set.VbX, set.VbY, set.VbW, set.VbH
	}
	groupLayer := &render.GroupLayer{
		Parts: groupParts,
		OutW:  outW,
		OutH:  outH,
		VbX:   vbX,
		VbY:   vbY,
		VbW:   vbW,
		VbH:   vbH,
	}
	return &theme.Theme{
		Name:    cat.Name,
		Canvas:  theme.Canvas{Width: outW, Height: outH},
		BgW:     outW,
		BgH:     outH,
		Display: cat.Display,
		Layers:  []imgcore.Layer{groupLayer},
	}, nil
}
