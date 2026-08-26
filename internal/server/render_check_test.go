package server

import (
	"strings"
	"testing"

	"github.com/miaoledor/lolicount/internal/imgcore/composer"
	"github.com/miaoledor/lolicount/internal/imgcore/theme"
)

// TestRenderDimensionsUnified checks that after the theme unification
// refactor, card themes (single-frame and multi-frame) and character
// themes all produce reasonable, scaled output dimensions. This catches
// regressions where a theme type is missed by the scale logic.
func TestRenderDimensionsUnified(t *testing.T) {
	reg, errs := composer.NewThemeRegistry()
	for _, e := range errs {
		t.Logf("registry warning: %v", e)
	}

	cases := []struct {
		name    string
		text    string
		fsize   int
	}{
		{"lian", "12345", 50},      // multi-frame card
		{"ao", "12345", 50},        // multi-frame card
		{"shiroha", "12345", 50},   // multi-frame card
		{"hinata", "12345", 50},    // character
		{"lian-ren", "12345", 50},  // character
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			entry, err := composer.ResolveTheme(reg, tc.name)
			if err != nil {
				t.Fatalf("resolve %s: %v", tc.name, err)
			}
			base, ok := reg.Get(entry.Name)
			if !ok || base == nil {
				t.Fatalf("get %s: not found", tc.name)
			}

			pos := theme.TextPos{}
			th, err := buildThemeLayers(base, 0, tc.text, tc.fsize, false, theme.TextStyle{}, pos)
			if err != nil {
				t.Fatalf("buildThemeLayers %s: %v", tc.name, err)
			}

			// BgW/BgH should be the scaled display size (longest edge = 400),
			// not the original pixel dimensions. Original frames are
			// typically 800+px; unscaled would be a clear regression.
			maxBg := th.BgW
			if th.BgH > maxBg {
				maxBg = th.BgH
			}
			if maxBg > 500 {
				t.Errorf("%s: bg=%dx%d, longest edge %d > 500 (scale not applied?)",
					tc.name, th.BgW, th.BgH, maxBg)
			}
			if maxBg < 100 {
				t.Errorf("%s: bg=%dx%d, longest edge %d < 100 (over-scaled?)",
					tc.name, th.BgW, th.BgH, maxBg)
			}

			svg, err := composer.Compose(composer.ComposeParams{
				Theme: th, Seed: tc.name + ":test", CountText: tc.text,
			})
			if err != nil {
				t.Fatalf("compose %s: %v", tc.name, err)
			}
			if !strings.Contains(svg, "viewBox=") {
				t.Errorf("%s: SVG missing viewBox", tc.name)
			}
			if !strings.Contains(svg, "<image") {
				t.Errorf("%s: SVG missing <image> element", tc.name)
			}
			if !strings.Contains(svg, "<text") {
				t.Errorf("%s: SVG missing <text> element", tc.name)
			}
		})
	}
}
