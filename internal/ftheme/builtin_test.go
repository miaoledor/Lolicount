package ftheme

import "testing"

// NewBuiltinRegistry must load at least one f-theme and every loaded
// style must be well-formed: its Name matches the registry key. This
// asserts invariants only — it does NOT depend on which styles or how
// many are shipped, so adding or removing an f-theme cannot break it
// (AGENTS.md: theme content and count must not affect test results).
func TestBuiltinRegistryInvariants(t *testing.T) {
	reg, errs := NewBuiltinRegistry()
	for _, e := range errs {
		t.Errorf("load error: %v", e)
	}
	list := reg.List()
	if len(list) == 0 {
		t.Fatal("no f-themes loaded; expected at least one")
	}
	for _, name := range list {
		st, ok := reg.Get(name)
		if !ok {
			t.Errorf("List contains %q but Get fails", name)
			continue
		}
		if st.Name != name {
			t.Errorf("f-theme %q has Name %q", name, st.Name)
		}
	}
}

func TestBuiltinListSorted(t *testing.T) {
	reg, _ := NewBuiltinRegistry()
	list := reg.List()
	if len(list) == 0 {
		t.Fatal("no f-themes listed")
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
		t.Error("expected missing f-theme to return false")
	}
}
