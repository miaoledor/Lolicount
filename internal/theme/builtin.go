package theme

import (
	"fmt"
	"io/fs"
	"sort"
	"strings"

	"github.com/miaoledor/lolicount/assets"
)

// builtinRegistry loads themes from the embedded asset trees at
// construction time. Frame themes live under assets/theme (each
// subdirectory is one theme; frame files are 0.png..size-1.png,
// gif/png/webp accepted). Character themes live under assets/character
// (each subdirectory is one layered-portrait theme with ren.json + ren/).
// The directory root determines the Kind, so a theme's location is its
// type — no per-directory manifest probing is needed.
type builtinRegistry struct {
	themes map[string]*Theme
}

// NewBuiltinRegistry scans the embedded assets/theme and assets/character
// directories and loads every valid theme into memory. A theme that
// fails to load is skipped with an error; the registry still returns
// successfully with the themes that did load.
func NewBuiltinRegistry() (Registry, []error) {
	reg := &builtinRegistry{themes: make(map[string]*Theme)}
	var errs []error

	// Frame themes: assets/theme/<name>/0.png ...
	if frameErrs := scanRoot(reg, "theme", loadFrameTheme); len(frameErrs) > 0 {
		errs = append(errs, frameErrs...)
	}
	// Character themes: assets/character/<name>/ren.json + ren/...
	if charErrs := scanRoot(reg, "character", loadCharacterTheme); len(charErrs) > 0 {
		errs = append(errs, charErrs...)
	}
	return reg, errs
}

// scanRoot reads one embedded root (e.g. "theme" or "character") and
// loads each subdirectory via the given loader. Errors are collected
// per-theme so one bad theme does not abort the rest.
func scanRoot(reg *builtinRegistry, root string, load func(fs.FS, string) (*Theme, error)) []error {
	sub, err := fs.Sub(assets.FS, root)
	if err != nil {
		return []error{fmt.Errorf("theme: open embedded %s: %w", root, err)}
	}
	entries, err := fs.ReadDir(sub, ".")
	if err != nil {
		return []error{fmt.Errorf("theme: read %s: %w", root, err)}
	}
	var errs []error
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		name := e.Name()
		th, err := load(sub, name)
		if err != nil {
			errs = append(errs, err)
			continue
		}
		reg.themes[name] = th
	}
	return errs
}

// loadCharacterTheme loads one character-theme directory as a
// KindCharacter theme (ren.json + ren/*.webp, M9).
func loadCharacterTheme(fsys fs.FS, name string) (*Theme, error) {
	ch, err := LoadCharacter(fsys, name)
	if err != nil {
		return nil, fmt.Errorf("character %s: %w", name, err)
	}
	return &Theme{Name: name, Kind: KindCharacter, Character: ch}, nil
}

// loadFrameTheme decodes one frame-theme directory into an ordered Frame
// slice. Files are named <index>.<ext>; indices need not be contiguous
// but must be non-negative. Frames are stored sorted by index.
func loadFrameTheme(fsys fs.FS, name string) (*Theme, error) {
	entries, err := fs.ReadDir(fsys, name)
	if err != nil {
		return nil, fmt.Errorf("theme %s: read dir: %w", name, err)
	}

	type indexed struct {
		idx int
		f   Frame
	}
	var frames []indexed
	for _, e := range entries {
		base := e.Name()
		idx := frameIndexFromName(base)
		if idx < 0 {
			continue // skip non-frame files (README, meta.json, _start, etc.)
		}
		ext := strings.ToLower(pathExt(base))
		mime, ok := supportedExts[ext]
		if !ok {
			continue
		}
		f, err := decodeFrame(fsys, name+"/"+base, mime)
		if err != nil {
			return nil, fmt.Errorf("theme %s: %w", name, err)
		}
		frames = append(frames, indexed{idx: idx, f: f})
	}
	if len(frames) == 0 {
		return nil, fmt.Errorf("theme %s: no frame images found", name)
	}

	sort.Slice(frames, func(i, j int) bool { return frames[i].idx < frames[j].idx })
	th := &Theme{Name: name, Kind: KindFrame, Frames: make([]Frame, len(frames))}
	// Re-index to 0..N-1 so Size() and Frame(i) are dense and contiguous.
	for i, fr := range frames {
		th.Frames[i] = fr.f
	}
	return th, nil
}

// pathExt is path.Ext without importing path here (kept local to avoid
// an extra import line in the file header block).
func pathExt(name string) string {
	for i := len(name) - 1; i >= 0 && name[i] != '/'; i-- {
		if name[i] == '.' {
			return name[i:]
		}
	}
	return ""
}

// Get returns the theme for name, or false if not registered.
func (r *builtinRegistry) Get(name string) (*Theme, bool) {
	t, ok := r.themes[name]
	return t, ok
}

// List returns registered theme names sorted for stable output.
func (r *builtinRegistry) List() []string {
	out := make([]string, 0, len(r.themes))
	for k := range r.themes {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

// ListWithKind returns every registered theme with its Kind, sorted by
// name for stable output (M9).
func (r *builtinRegistry) ListWithKind() []ThemeInfo {
	out := make([]ThemeInfo, 0, len(r.themes))
	for k, t := range r.themes {
		out = append(out, ThemeInfo{Name: k, Kind: t.Kind})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out
}
