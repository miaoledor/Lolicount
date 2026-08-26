package composer

import (
	"testing"
)

// TestBuiltinFThemeRegistryLoadsJSON verifies the embedded f-theme JSON
// files are loaded into the registry at construction time.
func TestBuiltinFThemeRegistryLoadsJSON(t *testing.T) {
	reg, errs := newBuiltinFThemeRegistry()
	if len(errs) != 0 {
		t.Fatalf("unexpected errors: %v", errs)
	}

	want := []string{"default", "neon", "pink", "serif"}
	got := reg.List()
	if len(got) != len(want) {
		t.Fatalf("f-theme count = %d (%v), want %d (%v)", len(got), got, len(want), want)
	}
	for _, name := range want {
		if _, ok := reg.Get(name); !ok {
			t.Errorf("f-theme %q not found", name)
		}
	}
}

// TestBuiltinFThemeRegistryNeonColor verifies a known style's color is
// loaded correctly, ensuring the JSON fields map to theme.FStyle.
func TestBuiltinFThemeRegistryNeonColor(t *testing.T) {
	reg, _ := newBuiltinFThemeRegistry()
	st, ok := reg.Get("neon")
	if !ok {
		t.Fatal("neon f-theme not found")
	}
	if st.Color != "#39ff14" {
		t.Errorf("neon color = %q, want #39ff14", st.Color)
	}
	if st.Weight != "bold" {
		t.Errorf("neon weight = %q, want bold", st.Weight)
	}
}

// TestBuiltinFThemeRegistryMissingName verifies Get returns false for an
// unregistered name.
func TestBuiltinFThemeRegistryMissingName(t *testing.T) {
	reg, _ := newBuiltinFThemeRegistry()
	if _, ok := reg.Get("does-not-exist"); ok {
		t.Error("expected false for unknown f-theme")
	}
}

// TestResolveFThemeRandom verifies ResolveFTheme picks a registered style
// for the reserved "random" value.
func TestResolveFThemeRandom(t *testing.T) {
	reg, _ := newBuiltinFThemeRegistry()
	st, err := ResolveFTheme(reg, "random")
	if err != nil {
		t.Fatalf("ResolveFTheme random: %v", err)
	}
	if st.Name == "" {
		t.Error("random returned empty style")
	}
}

// TestResolveFThemeEmpty verifies an empty name returns a zero style
// (the server treats empty as "use defaults").
func TestResolveFThemeEmpty(t *testing.T) {
	reg, _ := newBuiltinFThemeRegistry()
	st, err := ResolveFTheme(reg, "")
	if err != nil {
		t.Fatalf("ResolveFTheme empty: %v", err)
	}
	if st.Name != "" {
		t.Errorf("empty name returned non-zero style: %+v", st)
	}
}
