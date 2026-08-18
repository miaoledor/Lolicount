package ftheme

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"sort"
	"strings"

	"github.com/miaoledor/lolicount/assets"
)

// builtinRegistry loads font-style themes from the embedded
// assets/f-theme tree at construction time. Each assets/f-theme/<name>.json
// file describes one Style.
type builtinRegistry struct {
	styles map[string]Style
}

// NewBuiltinRegistry scans the embedded assets/f-theme directory and
// loads every *.json file as a Style. A malformed file is skipped with
// an error; the registry still returns with the styles that loaded.
func NewBuiltinRegistry() (Registry, []error) {
	reg := &builtinRegistry{styles: make(map[string]Style)}
	var errs []error

	root := "f-theme"
	sub, err := fs.Sub(assets.FS, root)
	if err != nil {
		return reg, []error{fmt.Errorf("ftheme: open embedded %s: %w", root, err)}
	}
	entries, err := fs.ReadDir(sub, ".")
	if err != nil {
		return reg, []error{fmt.Errorf("ftheme: read %s: %w", root, err)}
	}

	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".json") {
			continue
		}
		st, err := loadStyle(sub, e.Name())
		if err != nil {
			errs = append(errs, err)
			continue
		}
		reg.styles[st.Name] = st
	}
	return reg, errs
}

// loadStyle decodes one *.json file into a Style. The file's own "name"
// field is authoritative (not the filename).
func loadStyle(fsys fs.FS, filename string) (Style, error) {
	raw, err := fs.ReadFile(fsys, filename)
	if err != nil {
		return Style{}, fmt.Errorf("ftheme: read %s: %w", filename, err)
	}
	var st Style
	if err := json.Unmarshal(raw, &st); err != nil {
		return Style{}, fmt.Errorf("ftheme: parse %s: %w", filename, err)
	}
	if st.Name == "" {
		return Style{}, fmt.Errorf("ftheme: %s missing name", filename)
	}
	return st, nil
}

// Get returns the style for name, or false if not registered.
func (r *builtinRegistry) Get(name string) (Style, bool) {
	st, ok := r.styles[name]
	return st, ok
}

// List returns registered style names sorted for stable output.
func (r *builtinRegistry) List() []string {
	out := make([]string, 0, len(r.styles))
	for name := range r.styles {
		out = append(out, name)
	}
	sort.Strings(out)
	return out
}
