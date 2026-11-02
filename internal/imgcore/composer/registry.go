// Package composer is the sole rendering entry point for imgcore: it
// iterates a theme's layer stack, calls each layer's Render method, and
// concatenates the SVG fragments into the final document.
package composer

import (
	"fmt"
	"io/fs"
	"math/rand"
	"sort"
	"sync"

	"github.com/miaoledor/lolicount/assets"
	"github.com/miaoledor/lolicount/internal/imgcore"
	"github.com/miaoledor/lolicount/internal/imgcore/asset"
	"github.com/miaoledor/lolicount/internal/imgcore/imgutils"
	"github.com/miaoledor/lolicount/internal/imgcore/render"
	"github.com/miaoledor/lolicount/internal/imgcore/theme"
)

// ThemeEntry is a registry entry surfaced to the front-end. The theme
// kind (frame/character) is not exposed — all themes go through the
// same compose path regardless of layer count.
type ThemeEntry struct {
	Name     string
	Variants int
}

// ThemeRegistry provides unified access to all themes. The unified Get
// returns the theme's resident assembly (one loaded candidate per slot,
// per the lazy memory strategy) — callers must treat it as read-only;
// rendering paths should use RenderPreparer.PrepareRender instead, which
// swaps one slot and returns a private snapshot.
type ThemeRegistry interface {
	Get(name string) (*theme.Theme, bool)
	List() []ThemeEntry
}

// RenderPreparer is the lazy-registry capability used by the server's
// compose path: each call swaps in ONE freshly loaded slot candidate
// (on-demand DecodeImage) and returns a deep snapshot safe to render
// concurrently.
type RenderPreparer interface {
	PrepareRender(name string) (*theme.Theme, error)
}

// FThemeRegistry is the font-style registry interface.
type FThemeRegistry = theme.FThemeRegistry

// lazyEntry is one theme under the one-image-per-slot strategy: a
// metadata catalog, the resident assembly (exactly one loaded candidate
// per slot), and a mutex serializing swaps against snapshots.
type lazyEntry struct {
	cat  *asset.ThemeCatalog
	base *theme.Theme
	prng *imgutils.PRNG
	mu   sync.Mutex
}

// unifiedRegistry holds lazy entries for all themes from the embedded
// assets/theme/ tree.
type unifiedRegistry struct {
	themes map[string]*lazyEntry
}

// NewThemeRegistry loads all themes with the one-image-per-slot memory
// strategy: every theme's full candidate set is cataloged as metadata
// (paths + placement + header dims), and the resident theme holds a
// single loaded candidate per slot. The initial pick per slot is seeded
// from the theme name, so a given asset tree always assembles
// identically (thumbnails included).
func NewThemeRegistry() (ThemeRegistry, []error) {
	fsys, err := fs.Sub(assets.FS, "theme")
	if err != nil {
		return nil, []error{fmt.Errorf("registry: open embedded theme: %w", err)}
	}
	cats, errs := asset.LoadThemeCatalogs(fsys)
	reg := &unifiedRegistry{themes: make(map[string]*lazyEntry, len(cats))}
	for name, cat := range cats {
		prng := imgutils.NewPRNG(name)
		pick := func(n int) int {
			if n <= 1 {
				return 0
			}
			weights := make([]float64, n)
			for i := range weights {
				weights[i] = 1
			}
			return prng.WeightedPick(weights)
		}
		t, err := asset.AssembleTheme(fsys, cat, pick)
		if err != nil {
			errs = append(errs, err)
			continue
		}
		reg.themes[name] = &lazyEntry{cat: cat, base: t, prng: prng}
	}
	return reg, errs
}

// NewFThemeRegistry loads font-style themes from the embedded assets.
func NewFThemeRegistry() (FThemeRegistry, []error) {
	return newBuiltinFThemeRegistry()
}

// Get returns the theme's resident assembly. Read-only: the caller must
// not mutate it — use PrepareRender for the render path.
func (r *unifiedRegistry) Get(name string) (*theme.Theme, bool) {
	e, ok := r.themes[name]
	if !ok {
		return nil, false
	}
	return e.base, true
}

// List returns all registered themes sorted by name for stable output.
// Variants comes from the catalog (products within mutually exclusive
// sets, summed across sets), not from the resident layers.
func (r *unifiedRegistry) List() []ThemeEntry {
	out := make([]ThemeEntry, 0, len(r.themes))
	for name, e := range r.themes {
		out = append(out, ThemeEntry{Name: name, Variants: e.cat.Variants})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out
}

// PrepareRender swaps one slot then returns a deep snapshot. The swap
// picks a random multi-candidate slot, loads a different candidate from
// the embedded tree (DecodeImage — the only time those bytes touch the
// heap), and patches the resident assembly, so consecutive requests each
// evolve exactly one part. The snapshot decouples rendering from later
// swaps: every concurrent renderer owns its layer copies.
func (r *unifiedRegistry) PrepareRender(name string) (*theme.Theme, error) {
	e, ok := r.themes[name]
	if !ok {
		return nil, fmt.Errorf("theme %q not found", name)
	}
	e.mu.Lock()
	defer e.mu.Unlock()
	if len(e.cat.Sets) > 0 {
		return asset.AssembleTheme(themeFS, e.cat, func(n int) int {
			weights := make([]float64, n)
			for i := range weights {
				weights[i] = 1
			}
			return e.prng.WeightedPick(weights)
		})
	}
	r.swapOneLocked(e)
	return deepCopyTheme(e.base), nil
}

// swapOneLocked mutates the resident assembly by one slot. Load failure
// keeps the previous candidate (render still succeeds with the resident
// look). Caller holds e.mu.
func (r *unifiedRegistry) swapOneLocked(e *lazyEntry) {
	var swappable []int
	for i, s := range e.cat.Slots {
		if len(s.Candidates) > 1 {
			swappable = append(swappable, i)
		}
	}
	if len(swappable) == 0 {
		return
	}
	weights := make([]float64, len(swappable))
	for i := range weights {
		weights[i] = 1
	}
	slot := e.cat.Slots[swappable[e.prng.WeightedPick(weights)]]

	cw := make([]float64, len(slot.Candidates))
	for i := range cw {
		cw[i] = 1
	}
	next := e.prng.WeightedPick(cw)
	if next == slot.Loaded {
		next = (next + 1) % len(slot.Candidates)
	}
	cand := slot.Candidates[next]

	decoded, err := asset.DecodeImage(themeFS, cand.RelPath, cand.Mime)
	if err != nil {
		return
	}
	slot.Loaded = next
	if e.cat.FrameTheme {
		if rp, ok := e.base.Layers[0].(*render.RandomPickLayer); ok && len(rp.Options) > 0 {
			rp.Options[0].Src = decoded.Data
			rp.Options[0].Width = decoded.Width
			rp.Options[0].Height = decoded.Height
		}
		return
	}
	if gl, ok := e.base.Layers[0].(*render.GroupLayer); ok {
		for si := range gl.Parts {
			if e.cat.Slots[si] == slot {
				// Index assignment: ranging copies the GroupPart value.
				gl.Parts[si].Src = decoded.Data
				gl.Parts[si].X = cand.X
				gl.Parts[si].Y = cand.Y
				gl.Parts[si].Width = cand.Width
				gl.Parts[si].Height = cand.Height
				return
			}
		}
	}
}

// deepCopyTheme clones the layer stack so a snapshot can be rendered
// while the registry keeps swapping slots on the original.
func deepCopyTheme(t *theme.Theme) *theme.Theme {
	if t == nil {
		return nil
	}
	out := &theme.Theme{
		Name:    t.Name,
		Canvas:  t.Canvas,
		BgW:     t.BgW,
		BgH:     t.BgH,
		Display: t.Display,
	}
	out.Layers = make([]imgcore.Layer, len(t.Layers))
	for i, l := range t.Layers {
		switch layer := l.(type) {
		case *render.ImageLayer:
			cp := *layer
			out.Layers[i] = &cp
		case *render.RandomPickLayer:
			cp := *layer
			cp.Options = append([]render.ImageOption(nil), layer.Options...)
			cp.FrameDims = append([]render.FrameDim(nil), layer.FrameDims...)
			out.Layers[i] = &cp
		case *render.GroupLayer:
			cp := *layer
			cp.Parts = append([]render.GroupPart(nil), layer.Parts...)
			out.Layers[i] = &cp
		default:
			out.Layers[i] = l
		}
	}
	return out
}

// themeFS is the cached embedded theme subtree used for on-demand
// candidate loads (swaps); it cannot fail for a valid build.
var themeFS = func() fs.FS {
	sub, err := fs.Sub(assets.FS, "theme")
	if err != nil {
		panic(fmt.Sprintf("assets: theme subtree missing: %v", err))
	}
	return sub
}()

// ResolveTheme handles the reserved "random" value by picking from the
// registry, then returns the resolved theme entry.
func ResolveTheme(reg ThemeRegistry, name string) (ThemeEntry, error) {
	if name == "random" {
		list := reg.List()
		if len(list) == 0 {
			return ThemeEntry{}, fmt.Errorf("no themes available for random")
		}
		return list[rand.Intn(len(list))], nil
	}
	list := reg.List()
	for _, e := range list {
		if e.Name == name {
			return e, nil
		}
	}
	return ThemeEntry{}, fmt.Errorf("theme %q not found", name)
}

// ResolveFTheme handles the reserved "random" value by picking from the
// registry, then returns the resolved FStyle.
func ResolveFTheme(reg FThemeRegistry, name string) (theme.FStyle, error) {
	if name == "" {
		return theme.FStyle{}, nil
	}
	if name == "random" {
		list := reg.List()
		if len(list) == 0 {
			return theme.FStyle{}, fmt.Errorf("no f-themes available for random")
		}
		st, ok := reg.Get(list[rand.Intn(len(list))])
		if !ok {
			return theme.FStyle{}, fmt.Errorf("random f-theme missing")
		}
		return st, nil
	}
	st, ok := reg.Get(name)
	if !ok {
		return theme.FStyle{}, fmt.Errorf("f-theme %q not found", name)
	}
	return st, nil
}
