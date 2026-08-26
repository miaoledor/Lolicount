package composer

import (
	"fmt"
	"math/rand"
	"sort"

	"github.com/miaoledor/lolicount/internal/imgcore/asset"
	"github.com/miaoledor/lolicount/internal/imgcore/theme"
)

// ThemeEntry is a registry entry surfaced to the front-end.
type ThemeEntry struct {
	Name string
	Kind string // "frame" or "character"
}

// ThemeRegistry provides unified access to both card and character
// themes. The server depends on this interface instead of importing the
// individual loader packages.
type ThemeRegistry interface {
	GetCard(name string) (*asset.CardTheme, bool)
	GetCharacter(name string) (*asset.CharacterTheme, bool)
	Get(name string) (ThemeEntry, bool)
	List() []ThemeEntry
}

// FThemeRegistry is the font-style registry interface.
type FThemeRegistry = theme.FThemeRegistry

// unifiedRegistry holds both card and character registries.
type unifiedRegistry struct {
	cards      asset.CardRegistry
	characters asset.CharacterRegistry
}

// NewThemeRegistry loads both card and character themes from the
// embedded assets and returns a unified ThemeRegistry.
func NewThemeRegistry() (ThemeRegistry, []error) {
	cards, cardErrs := asset.NewBuiltinCardRegistry()
	characters, charErrs := asset.NewBuiltinCharacterRegistry()
	return &unifiedRegistry{cards: cards, characters: characters}, append(cardErrs, charErrs...)
}

// NewFThemeRegistry loads font-style themes from the embedded assets.
func NewFThemeRegistry() (FThemeRegistry, []error) {
	return newBuiltinFThemeRegistry()
}

func (r *unifiedRegistry) GetCard(name string) (*asset.CardTheme, bool) {
	return r.cards.Get(name)
}

func (r *unifiedRegistry) GetCharacter(name string) (*asset.CharacterTheme, bool) {
	return r.characters.Get(name)
}

func (r *unifiedRegistry) Get(name string) (ThemeEntry, bool) {
	if _, ok := r.cards.Get(name); ok {
		return ThemeEntry{Name: name, Kind: "frame"}, true
	}
	if _, ok := r.characters.Get(name); ok {
		return ThemeEntry{Name: name, Kind: "character"}, true
	}
	return ThemeEntry{}, false
}

func (r *unifiedRegistry) List() []ThemeEntry {
	var out []ThemeEntry
	for _, name := range r.cards.List() {
		out = append(out, ThemeEntry{Name: name, Kind: "frame"})
	}
	for _, name := range r.characters.List() {
		out = append(out, ThemeEntry{Name: name, Kind: "character"})
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
	entry, ok := reg.Get(name)
	if !ok {
		return ThemeEntry{}, fmt.Errorf("theme %q not found", name)
	}
	return entry, nil
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

// builtinFThemeRegistry loads font-style themes from assets/f-theme.
type builtinFThemeRegistry struct {
	styles map[string]theme.FStyle
}

func newBuiltinFThemeRegistry() (theme.FThemeRegistry, []error) {
	return &builtinFThemeRegistry{styles: make(map[string]theme.FStyle)}, nil
}

func (r *builtinFThemeRegistry) Get(name string) (theme.FStyle, bool) {
	st, ok := r.styles[name]
	return st, ok
}

func (r *builtinFThemeRegistry) List() []string {
	out := make([]string, 0, len(r.styles))
	for name := range r.styles {
		out = append(out, name)
	}
	sort.Strings(out)
	return out
}
