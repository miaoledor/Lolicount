package theme

import "testing"

// NewBuiltinRegistry must load the embedded loli theme with all 10
// digits and correct dimensions (gif 46x64). This also exercises
// decodeGlyph end-to-end against the real embed.FS.
func TestBuiltinLoadsLoli(t *testing.T) {
	reg, errs := NewBuiltinRegistry()
	for _, e := range errs {
		t.Logf("load warning: %v", e)
	}
	if reg == nil {
		t.Fatal("registry is nil")
	}
	th, ok := reg.Get("loli")
	if !ok {
		t.Fatal("loli theme not loaded")
	}
	for _, d := range digits {
		c, ok := th.Lookup(d)
		if !ok {
			t.Errorf("loli missing glyph %q", d)
			continue
		}
		if c.Width <= 0 || c.Height != 64 {
			t.Errorf("loli %q dims = %dx%d, want W>0 x64", d, c.Width, c.Height)
		}
		if c.Data == "" || c.Data[:5] != "data:" {
			t.Errorf("loli %q data uri malformed", d)
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
