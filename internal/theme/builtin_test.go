package theme

import "testing"

// NewBuiltinRegistry must load the embedded lian theme with its frames.
// This also exercises decodeFrame end-to-end against the real embed.FS.
func TestBuiltinLoadsLian(t *testing.T) {
	reg, errs := NewBuiltinRegistry()
	for _, e := range errs {
		t.Logf("load warning: %v", e)
	}
	if reg == nil {
		t.Fatal("registry is nil")
	}
	th, ok := reg.Get("lian")
	if !ok {
		t.Fatal("lian theme not loaded")
	}
	if th.Size() != 12 {
		t.Fatalf("lian size = %d, want 12", th.Size())
	}
	f, ok := th.Frame(0)
	if !ok {
		t.Fatal("lian frame 0 missing")
	}
	if f.Width != 508 || f.Height != 512 {
		t.Errorf("lian frame 0 dims = %dx%d, want 508x512", f.Width, f.Height)
	}
	if f.Data == "" || f.Data[:5] != "data:" {
		t.Errorf("lian frame 0 data uri malformed")
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

// M5.6: NewBuiltinRegistry must auto-scan every theme directory under
// assets/theme at startup. Both shipped themes (lian, lian-st) must be
// present so they are usable without extra registration.
func TestBuiltinScansAllThemes(t *testing.T) {
	reg, errs := NewBuiltinRegistry()
	for _, e := range errs {
		t.Errorf("load error: %v", e)
	}
	want := []string{"lian", "lian-st"}
	list := reg.List()
	for _, w := range want {
		found := false
		for _, got := range list {
			if got == w {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("theme %q not auto-scanned; got list %v", w, list)
		}
		if th, ok := reg.Get(w); !ok || th.Size() == 0 {
			t.Errorf("theme %q loaded but empty", w)
		}
	}
}
