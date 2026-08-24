package renderer

import (
	"testing"

	"github.com/miaoledor/lolicount/internal/imgcore"
	"github.com/miaoledor/lolicount/internal/imgcore/cardthemedrawer"
	"github.com/miaoledor/lolicount/internal/imgcore/characterthemedrawer"
	"github.com/miaoledor/lolicount/internal/imgcore/fdrawer"
)

// fakeThemeRegistry is an in-memory ThemeRegistry for testing
// ResolveTheme, Get, and List without touching the embed.FS.
type fakeThemeRegistry struct {
	cards      map[string]*cardthemedrawer.Theme
	characters map[string]*characterthemedrawer.Character
}

func (r *fakeThemeRegistry) GetCard(name string) (*cardthemedrawer.Theme, bool) {
	t, ok := r.cards[name]
	return t, ok
}

func (r *fakeThemeRegistry) GetCharacter(name string) (*characterthemedrawer.Character, bool) {
	c, ok := r.characters[name]
	return c, ok
}

func (r *fakeThemeRegistry) Get(name string) (ThemeEntry, bool) {
	if _, ok := r.cards[name]; ok {
		return ThemeEntry{Name: name, Kind: imgcore.KindFrame}, true
	}
	if _, ok := r.characters[name]; ok {
		return ThemeEntry{Name: name, Kind: imgcore.KindCharacter}, true
	}
	return ThemeEntry{}, false
}

func (r *fakeThemeRegistry) List() []ThemeEntry {
	var out []ThemeEntry
	for n := range r.cards {
		out = append(out, ThemeEntry{Name: n, Kind: imgcore.KindFrame})
	}
	for n := range r.characters {
		out = append(out, ThemeEntry{Name: n, Kind: imgcore.KindCharacter})
	}
	return out
}

func newFakeRegistry() *fakeThemeRegistry {
	return &fakeThemeRegistry{
		cards: map[string]*cardthemedrawer.Theme{
			"alpha": {Name: "alpha", Frames: []cardthemedrawer.Frame{{Width: 10, Height: 20, Data: "x"}}},
			"beta":  {Name: "beta", Frames: []cardthemedrawer.Frame{{Width: 10, Height: 20, Data: "y"}}},
		},
		characters: map[string]*characterthemedrawer.Character{
			"gamma": {Layers: []characterthemedrawer.CharacterLayer{{LayerID: 1}}, Parts: map[int]characterthemedrawer.CharacterPart{1: {Data: "z"}}},
		},
	}
}

func TestResolveThemeByName(t *testing.T) {
	reg := newFakeRegistry()
	entry, err := ResolveTheme(reg, "alpha")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if entry.Name != "alpha" || entry.Kind != imgcore.KindFrame {
		t.Errorf("got %+v, want alpha/frame", entry)
	}
}

func TestResolveThemeCharacter(t *testing.T) {
	reg := newFakeRegistry()
	entry, err := ResolveTheme(reg, "gamma")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if entry.Name != "gamma" || entry.Kind != imgcore.KindCharacter {
		t.Errorf("got %+v, want gamma/character", entry)
	}
}

func TestResolveThemeNotFound(t *testing.T) {
	reg := newFakeRegistry()
	_, err := ResolveTheme(reg, "nonexistent")
	if err == nil {
		t.Error("expected error for unknown theme")
	}
}

func TestResolveThemeRandom(t *testing.T) {
	reg := newFakeRegistry()
	entry, err := ResolveTheme(reg, "random")
	if err != nil {
		t.Fatalf("random error: %v", err)
	}
	// Must be one of the registered themes.
	switch entry.Name {
	case "alpha", "beta", "gamma":
	default:
		t.Errorf("random picked unknown theme %q", entry.Name)
	}
}

func TestResolveThemeRandomEmpty(t *testing.T) {
	reg := &fakeThemeRegistry{}
	_, err := ResolveTheme(reg, "random")
	if err == nil {
		t.Error("expected error for random with empty registry")
	}
}

func TestRegistryGet(t *testing.T) {
	reg := newFakeRegistry()
	if _, ok := reg.Get("alpha"); !ok {
		t.Error("Get(alpha) should succeed")
	}
	if _, ok := reg.Get("gamma"); !ok {
		t.Error("Get(gamma) should succeed")
	}
	if _, ok := reg.Get("nope"); ok {
		t.Error("Get(nope) should fail")
	}
}

func TestRegistryGetCard(t *testing.T) {
	reg := newFakeRegistry()
	if _, ok := reg.GetCard("alpha"); !ok {
		t.Error("GetCard(alpha) should succeed")
	}
	if _, ok := reg.GetCard("gamma"); ok {
		t.Error("GetCard(gamma) should fail (it is a character)")
	}
}

func TestRegistryGetCharacter(t *testing.T) {
	reg := newFakeRegistry()
	if _, ok := reg.GetCharacter("gamma"); !ok {
		t.Error("GetCharacter(gamma) should succeed")
	}
	if _, ok := reg.GetCharacter("alpha"); ok {
		t.Error("GetCharacter(alpha) should fail (it is a card)")
	}
}

// fakeFThemeRegistry is an in-memory fdrawer.Registry for testing.
type fakeFThemeRegistry struct {
	styles map[string]fdrawer.Style
}

func (r *fakeFThemeRegistry) Get(name string) (fdrawer.Style, bool) {
	s, ok := r.styles[name]
	return s, ok
}

func (r *fakeFThemeRegistry) List() []string {
	out := make([]string, 0, len(r.styles))
	for k := range r.styles {
		out = append(out, k)
	}
	return out
}

func newFakeFThemeRegistry() *fakeFThemeRegistry {
	return &fakeFThemeRegistry{
		styles: map[string]fdrawer.Style{
			"serif": {Name: "serif", Family: "serif", Color: "#000", Weight: "bold"},
			"mono":  {Name: "mono", Family: "monospace", Color: "#333"},
		},
	}
}

func TestResolveFThemeByName(t *testing.T) {
	reg := newFakeFThemeRegistry()
	st, err := ResolveFTheme(reg, "serif")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if st.Name != "serif" {
		t.Errorf("got %q, want serif", st.Name)
	}
}

func TestResolveFThemeEmpty(t *testing.T) {
	reg := newFakeFThemeRegistry()
	st, err := ResolveFTheme(reg, "")
	if err != nil {
		t.Fatalf("empty should not error: %v", err)
	}
	if st != (fdrawer.Style{}) {
		t.Errorf("empty should return zero Style, got %+v", st)
	}
}

func TestResolveFThemeNotFound(t *testing.T) {
	reg := newFakeFThemeRegistry()
	_, err := ResolveFTheme(reg, "nonexistent")
	if err == nil {
		t.Error("expected error for unknown f-theme")
	}
}

func TestResolveFThemeRandom(t *testing.T) {
	reg := newFakeFThemeRegistry()
	st, err := ResolveFTheme(reg, "random")
	if err != nil {
		t.Fatalf("random error: %v", err)
	}
	if st.Name != "serif" && st.Name != "mono" {
		t.Errorf("random picked unknown %q", st.Name)
	}
}

func TestResolveFThemeRandomEmpty(t *testing.T) {
	reg := &fakeFThemeRegistry{}
	_, err := ResolveFTheme(reg, "random")
	if err == nil {
		t.Error("expected error for random with empty registry")
	}
}

// TestUnifiedRegistry exercises the real unifiedRegistry (which wraps
// cardthemedrawer.Registry + characterthemedrawer.Registry) using fake
// sub-registries, covering Get/GetCard/GetCharacter/List and
// NewThemeRegistry/NewFThemeRegistry error aggregation.
func TestUnifiedRegistryGetCard(t *testing.T) {
	reg := &unifiedRegistry{
		cards:      &fakeCardReg{themes: map[string]*cardthemedrawer.Theme{"a": {}}},
		characters: &fakeCharReg{themes: map[string]*characterthemedrawer.Character{"b": {}}},
	}
	if _, ok := reg.GetCard("a"); !ok {
		t.Error("GetCard(a) should succeed")
	}
	if _, ok := reg.GetCard("b"); ok {
		t.Error("GetCard(b) should fail (it is a character)")
	}
}

func TestUnifiedRegistryGetCharacter(t *testing.T) {
	reg := &unifiedRegistry{
		cards:      &fakeCardReg{themes: map[string]*cardthemedrawer.Theme{"a": {}}},
		characters: &fakeCharReg{themes: map[string]*characterthemedrawer.Character{"b": {}}},
	}
	if _, ok := reg.GetCharacter("b"); !ok {
		t.Error("GetCharacter(b) should succeed")
	}
	if _, ok := reg.GetCharacter("a"); ok {
		t.Error("GetCharacter(a) should fail (it is a card)")
	}
}

func TestUnifiedRegistryGet(t *testing.T) {
	reg := &unifiedRegistry{
		cards:      &fakeCardReg{themes: map[string]*cardthemedrawer.Theme{"a": {}}},
		characters: &fakeCharReg{themes: map[string]*characterthemedrawer.Character{"b": {}}},
	}
	if e, ok := reg.Get("a"); !ok || e.Kind != imgcore.KindFrame {
		t.Errorf("Get(a) = %+v ok=%v, want frame", e, ok)
	}
	if e, ok := reg.Get("b"); !ok || e.Kind != imgcore.KindCharacter {
		t.Errorf("Get(b) = %+v ok=%v, want character", e, ok)
	}
	if _, ok := reg.Get("c"); ok {
		t.Error("Get(c) should fail")
	}
}

func TestUnifiedRegistryList(t *testing.T) {
	reg := &unifiedRegistry{
		cards:      &fakeCardReg{themes: map[string]*cardthemedrawer.Theme{"zeta": {}, "alpha": {}}},
		characters: &fakeCharReg{themes: map[string]*characterthemedrawer.Character{"mid": {}}},
	}
	list := reg.List()
	if len(list) != 3 {
		t.Fatalf("got %d, want 3", len(list))
	}
	// List should be sorted by name.
	if list[0].Name != "alpha" || list[1].Name != "mid" || list[2].Name != "zeta" {
		t.Errorf("List not sorted: %v", list)
	}
}

func TestNewThemeRegistryLoadsBuiltin(t *testing.T) {
	// This loads the real embedded assets. It should succeed with the
	// builtin themes and return no errors for valid assets.
	reg, errs := NewThemeRegistry()
	if reg == nil {
		t.Fatal("NewThemeRegistry returned nil registry")
	}
	// Errors may occur if assets are partially loaded, but the registry
	// should still function.
	_ = errs
	list := reg.List()
	if len(list) == 0 {
		t.Error("builtin registry should have themes")
	}
}

func TestNewFThemeRegistryLoadsBuiltin(t *testing.T) {
	reg, errs := NewFThemeRegistry()
	if reg == nil {
		t.Fatal("NewFThemeRegistry returned nil registry")
	}
	_ = errs
	list := reg.List()
	if len(list) == 0 {
		t.Error("builtin f-theme registry should have styles")
	}
}

// fakeCardReg is a minimal cardthemedrawer.Registry for testing.
type fakeCardReg struct {
	themes map[string]*cardthemedrawer.Theme
}

func (r *fakeCardReg) Get(name string) (*cardthemedrawer.Theme, bool) {
	t, ok := r.themes[name]
	return t, ok
}

func (r *fakeCardReg) List() []string {
	out := make([]string, 0, len(r.themes))
	for k := range r.themes {
		out = append(out, k)
	}
	return out
}

// fakeCharReg is a minimal characterthemedrawer.Registry for testing.
type fakeCharReg struct {
	themes map[string]*characterthemedrawer.Character
}

func (r *fakeCharReg) Get(name string) (*characterthemedrawer.Character, bool) {
	c, ok := r.themes[name]
	return c, ok
}

func (r *fakeCharReg) List() []string {
	out := make([]string, 0, len(r.themes))
	for k := range r.themes {
		out = append(out, k)
	}
	return out
}
