package cardthemedrawer

import "testing"

// NewBuiltinRegistry must load at least one card theme from the embedded
// assets and every loaded theme must be well-formed. Asserts invariants
// only — does not depend on which themes or how many are shipped
// (AGENTS.md: theme content and count must not affect test results).
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
		if th.Size() == 0 {
			t.Errorf("frame theme %q has no frames", name)
			continue
		}
		f, ok := th.FrameAt(0)
		if !ok {
			t.Errorf("theme %q: FrameAt(0) missing", name)
			continue
		}
		if f.Width <= 0 || f.Height <= 0 {
			t.Errorf("theme %q: frame 0 dims invalid %dx%d", name, f.Width, f.Height)
		}
		if f.Data == "" || f.Data[:5] != "data:" {
			t.Errorf("theme %q: frame 0 data URI malformed", name)
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
