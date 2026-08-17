package theme

import (
	"fmt"
	"io/fs"
	"sort"
	"strings"

	"github.com/miaoledor/lolicount/assets"
)

// digits are the required glyph slots every theme must provide.
var digits = []CharName{"0", "1", "2", "3", "4", "5", "6", "7", "8", "9"}

// decorations are optional glyph slots rendered around the digit run.
var decorations = []CharName{"_start", "_end"}

// builtinRegistry loads themes from the embedded assets/theme tree at
// construction time. Each subdirectory of assets/theme is one theme;
// its 0..9 glyphs are required, _start/_end are optional.
type builtinRegistry struct {
	themes map[string]*Theme
}

// NewBuiltinRegistry scans the embedded assets/theme directory and loads
// every valid theme into memory. A theme missing any of 0..9 is skipped
// with an error collected in the returned slice; the registry still
// returns successfully with the themes that did load.
func NewBuiltinRegistry() (Registry, []error) {
	root := "assets/theme"
	sub, err := fs.Sub(assets.FS, root)
	if err != nil {
		return nil, []error{fmt.Errorf("theme: open embedded %s: %w", root, err)}
	}

	entries, err := fs.ReadDir(sub, ".")
	if err != nil {
		return nil, []error{fmt.Errorf("theme: read %s: %w", root, err)}
	}

	reg := &builtinRegistry{themes: make(map[string]*Theme)}
	var errs []error
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

// loadTheme decodes one theme directory: requires 0..9, allows _start/_end.
func loadTheme(fsys fs.FS, name string) (*Theme, error) {
	entries, err := fs.ReadDir(fsys, name)
	if err != nil {
		return nil, fmt.Errorf("theme %s: read dir: %w", name, err)
	}

	th := &Theme{Name: name, Chars: make(map[CharName]ThemeChar)}
	for _, e := range entries {
		base := e.Name()
		ext, mime, ok := isSupportedGlyph(base)
		if !ok {
			continue
		}
		slot := CharName(strings.TrimSuffix(base, ext))
		if !isKnownSlot(slot) {
			continue
		}
		gc, err := decodeGlyph(fsys, name+"/"+base, mime)
		if err != nil {
			return nil, fmt.Errorf("theme %s: %w", name, err)
		}
		th.Chars[slot] = gc
	}

	for _, d := range digits {
		if _, ok := th.Chars[d]; !ok {
			return nil, fmt.Errorf("theme %s: missing required glyph %q", name, d)
		}
	}
	return th, nil
}

// isKnownSlot reports whether slot is a digit or a known decoration.
func isKnownSlot(slot CharName) bool {
	for _, d := range digits {
		if slot == d {
			return true
		}
	}
	for _, d := range decorations {
		if slot == d {
			return true
		}
	}
	return false
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
