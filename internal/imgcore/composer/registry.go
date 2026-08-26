// Package composer is the sole rendering entry point for imgcore: it
// iterates a theme's layer stack, calls each layer's Render method, and
// concatenates the SVG fragments into the final document.
package composer

import (
	"fmt"
	"math/rand"
	"sort"

	"github.com/miaoledor/lolicount/internal/imgcore/asset"
	"github.com/miaoledor/lolicount/internal/imgcore/theme"
)

// ThemeEntry is a registry entry surfaced to the front-end. The theme
// kind (frame/character) is not exposed — all themes go through the
// same compose path regardless of layer count.
type ThemeEntry struct {
	Name string
}

// ThemeRegistry provides unified access to all themes. The unified Get
// returns a *theme.Theme directly. Both frame and character themes are
// converted to *theme.Theme at load time; the registry does not
// distinguish them.
type ThemeRegistry interface {
	Get(name string) (*theme.Theme, bool)
	List() []ThemeEntry
}

// FThemeRegistry is the font-style registry interface.
type FThemeRegistry = theme.FThemeRegistry

// unifiedRegistry loads all themes from the embedded assets/theme/ tree
// and stores them as *theme.Theme. Each theme is loaded as a *theme.Theme
// regardless of its internal structure (frame vs character).
type unifiedRegistry struct {
	themes map[string]*theme.Theme
}

// NewThemeRegistry loads all themes from the embedded assets and returns
// a unified ThemeRegistry. Each theme is loaded as a *theme.Theme
// regardless of its source structure.
func NewThemeRegistry() (ThemeRegistry, []error) {
	loaded, errs := asset.LoadThemes()
	reg := &unifiedRegistry{themes: loaded}
	return reg, errs
}

// NewFThemeRegistry loads font-style themes from the embedded assets.
func NewFThemeRegistry() (FThemeRegistry, []error) {
	return newBuiltinFThemeRegistry()
}

// Get returns the *theme.Theme for name, or false if not registered.
func (r *unifiedRegistry) Get(name string) (*theme.Theme, bool) {
	t, ok := r.themes[name]
	return t, ok
}

// List returns all registered themes sorted by name for stable output.
func (r *unifiedRegistry) List() []ThemeEntry {
	out := make([]ThemeEntry, 0, len(r.themes))
	for name := range r.themes {
		out = append(out, ThemeEntry{Name: name})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out
}

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
