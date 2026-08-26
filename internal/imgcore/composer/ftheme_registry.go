package composer

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"sort"
	"strings"

	"github.com/miaoledor/lolicount/assets"
	"github.com/miaoledor/lolicount/internal/imgcore/theme"
)

// builtinFThemeRegistry loads font-style themes from the embedded
// assets/f-theme tree at construction time. Each *.json file decodes into
// a theme.FStyle. Migrated from fdrawer.NewBuiltinRegistry.
type builtinFThemeRegistry struct {
	styles map[string]theme.FStyle
}

// newBuiltinFThemeRegistry scans the embedded assets/f-theme directory
// and loads every *.json file as a theme.FStyle. Files that fail to
// parse are collected as errors so a single bad file does not prevent
// the rest from loading.
func newBuiltinFThemeRegistry() (theme.FThemeRegistry, []error) {
	reg := &builtinFThemeRegistry{styles: make(map[string]theme.FStyle)}
	var errs []error

	root := "f-theme"
	sub, err := fs.Sub(assets.FS, root)
	if err != nil {
		return reg, []error{fmt.Errorf("f-theme loader: open embedded %s: %w", root, err)}
	}
	entries, err := fs.ReadDir(sub, ".")
	if err != nil {
		return reg, []error{fmt.Errorf("f-theme loader: read %s: %w", root, err)}
	}

	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".json") {
			continue
		}
		st, err := loadFStyle(sub, e.Name())
		if err != nil {
			errs = append(errs, err)
			continue
		}
		reg.styles[st.Name] = st
	}
	return reg, errs
}

// loadFStyle decodes one *.json file into a theme.FStyle.
func loadFStyle(fsys fs.FS, filename string) (theme.FStyle, error) {
	raw, err := fs.ReadFile(fsys, filename)
	if err != nil {
		return theme.FStyle{}, fmt.Errorf("f-theme loader: read %s: %w", filename, err)
	}
	var st theme.FStyle
	if err := json.Unmarshal(raw, &st); err != nil {
		return theme.FStyle{}, fmt.Errorf("f-theme loader: parse %s: %w", filename, err)
	}
	if st.Name == "" {
		return theme.FStyle{}, fmt.Errorf("f-theme loader: %s missing name", filename)
	}
	return st, nil
}

// Get returns the style for name, or false if not registered.
func (r *builtinFThemeRegistry) Get(name string) (theme.FStyle, bool) {
	st, ok := r.styles[name]
	return st, ok
}

// List returns registered style names sorted for stable output.
func (r *builtinFThemeRegistry) List() []string {
	out := make([]string, 0, len(r.styles))
	for name := range r.styles {
		out = append(out, name)
	}
	sort.Strings(out)
	return out
}
