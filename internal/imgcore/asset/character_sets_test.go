package asset

import (
	"io/fs"
	"testing"

	"github.com/miaoledor/lolicount/assets"
)

// TestMaskedOutfitKeepsFaceOverlaySeparate locks the layer contract of the
// two Senren Banka Mako themes. Ninja bodies already draw a scarf over the
// lower face, so exposing the positive-face mouth/blush slot on that set
// paints a second mouth on top of the scarf. The casual bodies have a bare
// face and need that slot. Splitting the bodies into mutually exclusive
// sets also stops eyes and mouths from being picked across outfit
// orientations. The assertions describe configuration only: adding or
// removing artwork, frames or other themes cannot break this test.
func TestMaskedOutfitKeepsFaceOverlaySeparate(t *testing.T) {
	fsys, err := fs.Sub(assets.FS, "theme")
	if err != nil {
		t.Fatalf("open embedded theme tree: %v", err)
	}

	for _, themeName := range []string{"senren-mako", "senren-mako-b"} {
		t.Run(themeName, func(t *testing.T) {
			manifest, cfg, err := readCharacterManifest(fsys, themeName)
			if err != nil {
				t.Fatalf("read manifest: %v", err)
			}
			if len(cfg.Sets) != 2 {
				t.Fatalf("sets = %d, want exactly one casual and one ninja set", len(cfg.Sets))
			}

			sets := make(map[string]CharacterSet, len(cfg.Sets))
			for _, set := range cfg.Sets {
				sets[set.Name] = set
			}
			ninja, hasNinja := sets["ninja"]
			casual, hasCasual := sets["casual"]
			if !hasNinja || !hasCasual {
				t.Fatalf("sets = %v, want a ninja and a casual set", sets)
			}

			if _, ok := ninja.Ranges["face"]; ok {
				t.Error("ninja set must not expose a face overlay: the scarf already covers the lower face")
			}
			if _, ok := casual.Ranges["face"]; !ok {
				t.Error("casual set must keep its face overlay")
			}

			// Every set owns a distinct slice of the body range, so a
			// single render can never mix parts from both outfits.
			nrange, crange := ninja.Ranges["lass"], casual.Ranges["lass"]
			if nrange.First <= crange.Last && crange.First <= nrange.Last {
				t.Errorf("lass ranges overlap: ninja [%d,%d], casual [%d,%d]",
					nrange.First, nrange.Last, crange.First, crange.Last)
			}
			for _, set := range cfg.Sets {
				for category, r := range set.Ranges {
					if r.First < 0 || r.Last >= len(manifest) || r.First > r.Last {
						t.Errorf("set %q range %q [%d,%d] is outside ren.json indexes [0,%d]",
							set.Name, category, r.First, r.Last, len(manifest)-1)
					}
				}
			}
		})
	}
}
