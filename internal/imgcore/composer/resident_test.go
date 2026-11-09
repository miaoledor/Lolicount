package composer

import (
	"testing"

	"github.com/miaoledor/lolicount/internal/imgcore/render"
)

// residentNames lists the themes that currently hold an assembly in the heap.
// It locks each entry because Warm may run concurrently with renders.
func residentNames(reg ThemeRegistry) []string {
	r := reg.(*unifiedRegistry)
	var names []string
	for name, e := range r.themes {
		e.mu.Lock()
		if e.base != nil {
			names = append(names, name)
		}
		e.mu.Unlock()
	}
	return names
}

// TestResidentRegistryBoundsAssemblies pins the core contract: the catalog
// stays complete (List/Get/random still see every theme) while only the
// requested number of assemblies is retained at startup.
func TestResidentRegistryBoundsAssemblies(t *testing.T) {
	reg, errs := NewThemeRegistryWithResident(3)
	if len(errs) > 0 {
		t.Fatalf("registry errors: %v", errs[:min(len(errs), 3)])
	}
	all := reg.List()
	resident := residentNames(reg)
	if len(resident) != 3 {
		t.Fatalf("resident assemblies = %d, want 3", len(resident))
	}
	if len(all) <= len(resident) {
		t.Fatalf("catalog should stay complete: %d listed, %d resident", len(all), len(resident))
	}

	var cold string
	for _, e := range all {
		found := false
		for _, name := range resident {
			if name == e.Name {
				found = true
			}
		}
		if !found {
			cold = e.Name
			break
		}
	}
	if cold == "" {
		t.Fatal("no cold theme found")
	}
	base, ok := reg.Get(cold)
	if !ok || base == nil {
		t.Fatalf("cold theme %s not available through Get", cold)
	}
	if _, err := Compose(ComposeParams{Theme: base, Seed: "test", CountText: "0123456789"}); err != nil {
		t.Fatalf("cold Get assembly does not compose: %v", err)
	}
	if got := len(residentNames(reg)); got != 3 {
		t.Errorf("cold Get made an assembly resident: %d", got)
	}
}

// TestWarmReplacesResidentSet verifies the hourly hot-list refresh frees the
// old assemblies and builds only the new ones.
func TestWarmReplacesResidentSet(t *testing.T) {
	reg, errs := NewThemeRegistryWithResident(2)
	if len(errs) > 0 {
		t.Fatalf("registry errors: %v", errs[:min(len(errs), 3)])
	}
	warmer, ok := reg.(ResidentController)
	if !ok {
		t.Fatal("registry does not implement ResidentController")
	}

	warmer.Warm([]string{"lian-ren", "wenders"})
	got := residentNames(reg)
	if len(got) != 2 {
		t.Fatalf("after Warm: %d resident assemblies (%v), want 2", len(got), got)
	}
	want := map[string]bool{"lian-ren": true, "wenders": true}
	for _, name := range got {
		if !want[name] {
			t.Errorf("unexpected resident theme %s", name)
		}
	}

	warmer.Warm(nil)
	if got := residentNames(reg); len(got) != 0 {
		t.Errorf("Warm(nil) left %d resident assemblies", len(got))
	}
	if base, ok := reg.Get("wenders"); !ok || base == nil {
		t.Error("theme removed from resident set is no longer renderable")
	}
}

// TestColdAssemblyIsDeterministic protects gallery thumbnails: a cold theme
// must assemble to the same frame set on every call, independent of swaps.
func TestColdAssemblyIsDeterministic(t *testing.T) {
	reg, errs := NewThemeRegistryWithResident(1)
	if len(errs) > 0 {
		t.Fatalf("registry errors: %v", errs[:min(len(errs), 3)])
	}
	first, ok := reg.Get("wenders")
	if !ok {
		t.Fatal("wenders missing")
	}
	second, ok := reg.Get("wenders")
	if !ok {
		t.Fatal("wenders missing on second Get")
	}
	a, b := srcsOf(t, first), srcsOf(t, second)
	if len(a) == 0 || len(a) != len(b) {
		t.Fatalf("cold assemblies differ in slot count: %d vs %d", len(a), len(b))
	}
	for i := range a {
		if a[i] != b[i] {
			t.Fatalf("cold assembly slot %d is not deterministic", i)
		}
	}
}

// TestIsMultiFrameMatchesLayers checks that the metadata-only cache decision
// agrees with the layer-based decision it replaces, for both frame and
// character themes.
func TestIsMultiFrameMatchesLayers(t *testing.T) {
	reg, errs := NewThemeRegistry()
	if len(errs) > 0 {
		t.Fatalf("registry errors: %v", errs[:min(len(errs), 3)])
	}
	info := reg.(ThemeInfo)
	for _, name := range []string{"wenders", "lian-ren"} {
		base, ok := reg.Get(name)
		if !ok {
			t.Fatalf("%s missing", name)
		}
		layerBased := false
		for _, layer := range base.Layers {
			if _, ok := layer.(*render.RandomPickLayer); ok {
				layerBased = true
			}
			if _, ok := layer.(*render.GroupLayer); ok {
				layerBased = true
			}
		}
		if got := info.IsMultiFrame(name); got != layerBased {
			t.Errorf("%s: IsMultiFrame=%v, layers say %v", name, got, layerBased)
		}
	}
	if info.IsMultiFrame("no-such-theme") {
		t.Error("unknown theme should not be multi-frame")
	}
}
