package renderer

import (
	"fmt"
	"math/rand"
	"sort"

	"github.com/miaoledor/lolicount/internal/imgcore"
	"github.com/miaoledor/lolicount/internal/imgcore/cardthemedrawer"
	"github.com/miaoledor/lolicount/internal/imgcore/characterthemedrawer"
	"github.com/miaoledor/lolicount/internal/imgcore/fdrawer"
)

// ThemeEntry is a registry entry surfaced to the front-end: the theme
// name and whether it is a frame or character theme.
type ThemeEntry struct {
	Name string
	Kind imgcore.LegacyKind
}

// ThemeRegistry provides unified access to both card and character
// themes. The server depends on this interface instead of importing the
// individual drawer packages, keeping the server decoupled from the
// drawer layer.
type ThemeRegistry interface {
	// GetCard returns the card theme for name, or false if absent.
	GetCard(name string) (*cardthemedrawer.Theme, bool)
	// GetCharacter returns the character theme for name, or false.
	GetCharacter(name string) (*characterthemedrawer.Character, bool)
	// Get resolves any theme name to its Kind + a ThemeEntry. Returns
	// false if the name is not registered as either kind.
	Get(name string) (ThemeEntry, bool)
	// List returns every registered theme (card + character) with its
	// Kind, sorted by name.
	List() []ThemeEntry
}

// FThemeRegistry is the font-style registry (alias of fdrawer.Registry).
type FThemeRegistry = fdrawer.Registry

// unifiedRegistry holds both card and character registries.
type unifiedRegistry struct {
	cards     cardthemedrawer.Registry
	characters characterthemedrawer.Registry
}

// NewThemeRegistry loads both card and character themes from the embedded
// assets and returns a unified ThemeRegistry.
func NewThemeRegistry() (ThemeRegistry, []error) {
	cards, cardErrs := cardthemedrawer.NewBuiltinRegistry()
	characters, charErrs := characterthemedrawer.NewBuiltinRegistry()
	return &unifiedRegistry{cards: cards, characters: characters}, append(cardErrs, charErrs...)
}

// NewFThemeRegistry loads font-style themes from the embedded assets.
func NewFThemeRegistry() (FThemeRegistry, []error) {
	return fdrawer.NewBuiltinRegistry()
}

func (r *unifiedRegistry) GetCard(name string) (*cardthemedrawer.Theme, bool) {
	return r.cards.Get(name)
}

func (r *unifiedRegistry) GetCharacter(name string) (*characterthemedrawer.Character, bool) {
	return r.characters.Get(name)
}

func (r *unifiedRegistry) Get(name string) (ThemeEntry, bool) {
	if _, ok := r.cards.Get(name); ok {
		return ThemeEntry{Name: name, Kind: imgcore.LegacyKindFrame}, true
	}
	if _, ok := r.characters.Get(name); ok {
		return ThemeEntry{Name: name, Kind: imgcore.LegacyKindCharacter}, true
	}
	return ThemeEntry{}, false
}

func (r *unifiedRegistry) List() []ThemeEntry {
	var out []ThemeEntry
	for _, name := range r.cards.List() {
		out = append(out, ThemeEntry{Name: name, Kind: imgcore.LegacyKindFrame})
	}
	for _, name := range r.characters.List() {
		out = append(out, ThemeEntry{Name: name, Kind: imgcore.LegacyKindCharacter})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out
}

// ResolveTheme handles the reserved "random" value by picking from the
// registry, then returns the resolved theme entry. For a non-random
// name it just looks up the registry.
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
// registry, then returns the resolved Style.
func ResolveFTheme(reg FThemeRegistry, name string) (fdrawer.Style, error) {
	if name == "" {
		return fdrawer.Style{}, nil
	}
	if name == "random" {
		list := reg.List()
		if len(list) == 0 {
			return fdrawer.Style{}, fmt.Errorf("no f-themes available for random")
		}
		st, ok := reg.Get(list[rand.Intn(len(list))])
		if !ok {
			return fdrawer.Style{}, fmt.Errorf("random f-theme %q missing", list)
		}
		return st, nil
	}
	st, ok := reg.Get(name)
	if !ok {
		return fdrawer.Style{}, fmt.Errorf("f-theme %q not found", name)
	}
	return st, nil
}
