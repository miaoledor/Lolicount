package theme

import "testing"

// NewBuiltinRegistry must load the embedded loli theme with frames
// derived from its 0..9 gif files. This also exercises decodeFrame
// end-to-end against the real embed.FS.
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
	if th.Size() != 10 {
		t.Fatalf("loli size = %d, want 10", th.Size())
	}
	f, ok := th.Frame(0)
	if !ok {
		t.Fatal("loli frame 0 missing")
	}
	if f.Height != 64 {
		t.Errorf("loli frame 0 height = %d, want 64", f.Height)
	}
	if f.Data == "" || f.Data[:5] != "data:" {
		t.Errorf("loli frame 0 data uri malformed")
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
