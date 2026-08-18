package theme

import "testing"

// NewBuiltinRegistry must load at least one theme from the embedded
// assets and every loaded theme must be well-formed: frame themes have
// a non-empty frame set with valid data URIs, character themes carry
// their layered Character data. This asserts registry invariants only —
// it does NOT depend on which themes or how many are shipped, so adding
// or removing a theme cannot break it (AGENTS.md: theme content and
// count must not affect test results).
func TestBuiltinRegistryInvariants(t *testing.T) {
	reg, errs := NewBuiltinRegistry()
	for _, e := range errs {
		t.Errorf("load error: %v", e)
	}
	if reg == nil {
		t.Fatal("registry is nil")
	}
	list := reg.List()
	if len(list) == 0 {
		t.Fatal("no themes loaded; expected at least one")
	}
	for _, name := range list {
		th, ok := reg.Get(name)
		if !ok {
			t.Errorf("List contains %q but Get fails", name)
			continue
		}
		if th.Name != name {
			t.Errorf("theme %q has Name %q", name, th.Name)
		}
		switch th.Kind {
		case KindFrame:
			if th.Size() == 0 {
				t.Errorf("frame theme %q has no frames", name)
				continue
			}
			f, ok := th.Frame(0)
			if !ok {
				t.Errorf("theme %q: Frame(0) missing", name)
				continue
			}
			if f.Width <= 0 || f.Height <= 0 {
				t.Errorf("theme %q: frame 0 dims invalid %dx%d", name, f.Width, f.Height)
			}
			if f.Data == "" || f.Data[:5] != "data:" {
				t.Errorf("theme %q: frame 0 data URI malformed", name)
			}
		case KindCharacter:
			if th.Character == nil {
				t.Errorf("character theme %q has nil Character", name)
				continue
			}
			if len(th.Character.Layers) == 0 || len(th.Character.Parts) == 0 {
				t.Errorf("character theme %q has empty Layers/Parts", name)
			}
		default:
			t.Errorf("theme %q has unknown Kind %d", name, th.Kind)
		}
	}
}

func TestBuiltinListIsSorted(t *testing.T) {
	reg, _ := NewBuiltinRegistry()
	list := reg.List()
	if len(list) == 0 {
		t.Fatal("no themes listed")
	}
	for i := 1; i < len(list); i++ {
		if list[i-1] > list[i] {
			t.Errorf("List not sorted: %v", list)
		}
	}
}

func TestBuiltinGetMissing(t *testing.T) {
	reg, _ := NewBuiltinRegistry()
	if _, ok := reg.Get("does-not-exist"); ok {
		t.Error("expected missing theme to return false")
	}
}
