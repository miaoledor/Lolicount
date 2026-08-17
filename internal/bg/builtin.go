package bg

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"sort"
	"strings"

	"github.com/miaoledor/lolicount/assets"
)

// builtinRegistry loads background metadata from the embedded
// assets/bg tree at construction time. Each assets/bg/<name>.json file
// describes one Background (URL + dimensions). The image bytes are NOT
// embedded — only the metadata; the URL points to a CDN (Iron Rule 2).
type builtinRegistry struct {
	bgs map[string]Background
}

// NewBuiltinRegistry scans the embedded assets/bg directory and loads
// every *.json file as a Background. A malformed file is skipped with an
// error; the registry still returns with the backgrounds that loaded.
func NewBuiltinRegistry() (Registry, []error) {
	reg := &builtinRegistry{bgs: make(map[string]Background)}
	var errs []error

	root := "bg"
	sub, err := fs.Sub(assets.FS, root)
	if err != nil {
		return reg, []error{fmt.Errorf("bg: open embedded %s: %w", root, err)}
	}
	entries, err := fs.ReadDir(sub, ".")
	if err != nil {
		return reg, []error{fmt.Errorf("bg: read %s: %w", root, err)}
	}

	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".json") {
			continue
		}
		bg, err := loadBackground(sub, e.Name())
		if err != nil {
			errs = append(errs, err)
			continue
		}
		reg.bgs[bg.Name] = bg
	}
	return reg, errs
}

// loadBackground decodes one *.json metadata file into a Background.
// The file's own "name" field is authoritative (not the filename), so a
// renamed file still resolves by its declared name.
func loadBackground(fsys fs.FS, filename string) (Background, error) {
	raw, err := fs.ReadFile(fsys, filename)
	if err != nil {
		return Background{}, fmt.Errorf("bg: read %s: %w", filename, err)
	}
	var bg Background
	if err := json.Unmarshal(raw, &bg); err != nil {
		return Background{}, fmt.Errorf("bg: parse %s: %w", filename, err)
	}
	if bg.Name == "" {
		return Background{}, fmt.Errorf("bg: %s missing name", filename)
	}
	if bg.URL == "" {
		return Background{}, fmt.Errorf("bg %s: missing url", bg.Name)
	}
	if bg.Width <= 0 || bg.Height <= 0 {
		return Background{}, fmt.Errorf("bg %s: width/height must be > 0", bg.Name)
	}
	return bg, nil
}

// Get returns the background for name, or false if not registered.
func (r *builtinRegistry) Get(name string) (Background, bool) {
	bg, ok := r.bgs[name]
	return bg, ok
}

// List returns the names of all registered backgrounds, sorted for
// stable output.
func (r *builtinRegistry) List() []string {
	out := make([]string, 0, len(r.bgs))
	for name := range r.bgs {
		out = append(out, name)
	}
	sort.Strings(out)
	return out
}
