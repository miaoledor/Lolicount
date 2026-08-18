package theme

import (
	"fmt"
	"io/fs"
	"sort"
	"strings"

	"github.com/miaoledor/lolicount/assets"
)

// builtinRegistry loads themes from the embedded assets/theme tree at
// construction time. Each subdirectory of assets/theme is one theme; its
// frame files are named 0.png .. size-1.png (gif/png/webp all accepted).
type builtinRegistry struct {
	themes map[string]*Theme
}

// NewBuiltinRegistry scans the embedded assets/theme directory and loads
// every valid theme into memory. A theme with zero frame files is
// skipped with an error; the registry still returns successfully with
// the themes that did load.
func NewBuiltinRegistry() (Registry, []error) {
	reg := &builtinRegistry{themes: make(map[string]*Theme)}
	var errs []error

	root := "theme"
	sub, err := fs.Sub(assets.FS, root)
	if err != nil {
		return reg, []error{fmt.Errorf("theme: open embedded %s: %w", root, err)}
	}
	entries, err := fs.ReadDir(sub, ".")
	if err != nil {
		return reg, []error{fmt.Errorf("theme: read %s: %w", root, err)}
	}

	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		name := e.Name()
		th, err := loadTheme(sub, name)
		if err != nil {
			errs = append(errs, err)
			continue
		}
		reg.themes[name] = th
	}
	return reg, errs
}

// hasCharacterManifest reports whether the theme directory contains a
// ren.json, marking it as a KindCharacter (layered portrait) theme.
func hasCharacterManifest(fsys fs.FS, name string) bool {
	f, err := fsys.Open(name + "/ren.json")
	if err != nil {
		return false
	}
	f.Close()
	return true
}

// loadTheme decodes one theme directory into a Theme. A directory with
// ren.json is loaded as KindCharacter (layered portrait, M9); otherwise
// it is loaded as KindFrame with ordered 0.png..size-1.png frames.
func loadTheme(fsys fs.FS, name string) (*Theme, error) {
	if hasCharacterManifest(fsys, name) {
		ch, err := LoadCharacter(fsys, name)
		if err != nil {
			return nil, fmt.Errorf("theme %s: %w", name, err)
		}
		return &Theme{Name: name, Kind: KindCharacter, Character: ch}, nil
	}
	return loadFrameTheme(fsys, name)
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
	th := &Theme{Name: name, Frames: make([]Frame, len(frames))}
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
