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
// kind (frame/character) is no longer exposed — all themes go through
// the same compose path regardless of layer count.
type ThemeEntry struct {
	Name string
}

// ThemeRegistry provides unified access to all themes. The unified Get
// returns a *theme.Theme directly, eliminating the old dual-path
// GetCard/GetCharacter split. Theme kind is inferred at load time via
// Theme.IsCardTheme(), not from the source directory.
type ThemeRegistry interface {
	Get(name string) (*theme.Theme, bool)
	List() []ThemeEntry
}

// FThemeRegistry is the font-style registry interface.
type FThemeRegistry = theme.FThemeRegistry

// unifiedRegistry loads all themes (card + character) from the embedded
// assets and stores them as *theme.Theme. Card themes live under
// assets/theme/, character themes under assets/character/ — both are
// converted to the same *theme.Theme type at load time.
type unifiedRegistry struct {
	themes map[string]*theme.Theme
}

// NewThemeRegistry loads both card and character themes from the embedded
// assets and returns a unified ThemeRegistry. Each theme is loaded as a
// *theme.Theme regardless of its source directory.
func NewThemeRegistry() (ThemeRegistry, []error) {
	reg := &unifiedRegistry{themes: make(map[string]*theme.Theme)}
	var errs []error

	cards, cardErrs := asset.NewBuiltinCardRegistry()
	errs = append(errs, cardErrs...)
	for _, name := range cards.List() {
		ct, _ := cards.Get(name)
		reg.themes[name] = asset.CardThemeToTheme(ct)
	}

	characters, charErrs := asset.NewBuiltinCharacterRegistry()
	errs = append(errs, charErrs...)
	for _, name := range characters.List() {
		ct, _ := characters.Get(name)
		t, err := asset.CharacterThemeToTheme(ct)
		if err != nil {
			errs = append(errs, fmt.Errorf("character %s: %w", name, err))
			continue
		}
		reg.themes[name] = t
	}

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
	// Check existence without holding the *theme.Theme (List is cheaper
	// for the random path; for the normal path we just verify the name
	// exists and return the entry).
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
