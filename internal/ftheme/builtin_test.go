package ftheme

import "testing"

// NewBuiltinRegistry must load the embedded f-theme JSON files.
func TestBuiltinLoadsStyles(t *testing.T) {
	reg, errs := NewBuiltinRegistry()
	for _, e := range errs {
		t.Errorf("load error: %v", e)
	}
	want := []string{"default", "neon", "pink", "serif"}
	for _, w := range want {
		st, ok := reg.Get(w)
		if !ok {
			t.Errorf("f-theme %q not loaded", w)
			continue
		}
		if st.Name != w {
			t.Errorf("f-theme name = %q, want %q", st.Name, w)
		}
	}
}

func TestBuiltinListSorted(t *testing.T) {
	reg, _ := NewBuiltinRegistry()
	list := reg.List()
	if len(list) < 4 {
		t.Fatalf("expected >=4 f-themes, got %d", len(list))
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
