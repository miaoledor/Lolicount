package bg

import (
	"testing"
)

func TestNewBuiltinRegistryLoadsLoliStand(t *testing.T) {
	reg, errs := NewBuiltinRegistry()
	for _, e := range errs {
		t.Logf("bg load err: %v", e)
	}
	b, ok := reg.Get("loli-stand")
	if !ok {
		t.Fatal("expected loli-stand background to be loaded")
	}
	if b.URL == "" {
		t.Error("URL should not be empty")
	}
	if b.Width != 400 || b.Height != 300 {
		t.Errorf("dimensions: got %dx%d want 400x300", b.Width, b.Height)
	}
	if list := reg.List(); len(list) == 0 {
		t.Error("List should not be empty")
	}
}

func TestNewBuiltinRegistryMissingName(t *testing.T) {
	// The loli-stand.json has a name, so this just verifies a well-formed
	// registry rejects a missing-name file at load time (covered by
	// loadBackground). Sanity-check Get on unknown name.
	reg, _ := NewBuiltinRegistry()
	if _, ok := reg.Get("does-not-exist"); ok {
		t.Error("Get unknown name should return false")
	}
}
