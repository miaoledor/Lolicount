package asset_test

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path"
	"regexp"
	"strings"
	"testing"

	"github.com/miaoledor/lolicount/assets"
	"github.com/miaoledor/lolicount/internal/imgcore/asset"
	"github.com/miaoledor/lolicount/internal/imgcore/composer"
)

// minimumPriconneVariants is the variant floor below which a Princess Connect
// portrait is dropped entirely instead of shipping a nearly static theme.
const minimumPriconneVariants = 10

// TestPriconnePortraits exercises the real loader and SVG composition for every
// contributed portrait, including alternate draw orders and body-only models.
func TestPriconnePortraits(t *testing.T) {
	source, err := os.ReadFile("../../../scripts/priconne-manifest.txt")
	if err != nil {
		t.Fatal(err)
	}
	exclusionJSON, err := os.ReadFile("../../../scripts/priconne-exclusions.json")
	if err != nil {
		t.Fatal(err)
	}
	var exclusions map[string]string
	if err := json.Unmarshal(exclusionJSON, &exclusions); err != nil {
		t.Fatal(err)
	}
	nameJSON, err := os.ReadFile("../../../scripts/priconne-names.json")
	if err != nil {
		t.Fatal(err)
	}
	var names map[string]string
	if err := json.Unmarshal(nameJSON, &names); err != nil {
		t.Fatal(err)
	}
	if len(names) == 0 {
		t.Fatal("scripts/priconne-names.json holds no names")
	}
	silhouetteJSON, err := os.ReadFile("../../../scripts/priconne-silhouettes.json")
	if err != nil {
		t.Fatal(err)
	}
	var silhouettes map[string]string
	if err := json.Unmarshal(silhouetteJSON, &silhouettes); err != nil {
		t.Fatal(err)
	}
	singleVariantJSON, err := os.ReadFile("../../../scripts/priconne-excluded-themes.json")
	if err != nil {
		t.Fatal(err)
	}
	var singleVariantNames []string
	if err := json.Unmarshal(singleVariantJSON, &singleVariantNames); err != nil {
		t.Fatal(err)
	}
	singleVariants := make(map[string]bool, len(singleVariantNames))
	for _, name := range singleVariantNames {
		singleVariants[name] = true
	}
	// Portraits that show the same skin of one character are consolidated into a
	// single theme whose config keeps one mutually exclusive set per portrait, so
	// the plan below maps every portrait to the theme that ships it.
	mergedJSON, err := os.ReadFile("../../../scripts/priconne-merged-themes.json")
	if err != nil {
		t.Fatal(err)
	}
	var mergedGroups map[string][]string
	if err := json.Unmarshal(mergedJSON, &mergedGroups); err != nil {
		t.Fatal(err)
	}
	exceptionJSON, err := os.ReadFile("../../../scripts/priconne-merge-exceptions.json")
	if err != nil {
		t.Fatal(err)
	}
	var exceptions struct {
		Deleted    []string `json:"deleted"`
		Standalone []string `json:"standalone"`
	}
	if err := json.Unmarshal(exceptionJSON, &exceptions); err != nil {
		t.Fatal(err)
	}
	representative := make(map[string]string)
	groupMembers := make(map[string][]string, len(mergedGroups))
	for themeName, group := range mergedGroups {
		want := strings.TrimPrefix(themeName, "pcr-")
		if len(group) < 2 {
			t.Fatalf("merge plan keeps %s with %d portrait(s)", themeName, len(group))
		}
		if group[0] != want {
			t.Fatalf("merge plan names %s after %s, want the first portrait", themeName, group[0])
		}
		groupMembers[want] = group
		representative[want] = want
		for _, member := range group {
			representative[member] = want
		}
	}
	standalone := make(map[string]bool, len(exceptions.Standalone))
	for _, model := range exceptions.Standalone {
		standalone[model] = true
		if _, merged := representative[model]; merged {
			t.Fatalf("portrait %s is declared standalone but appears in the merge plan", model)
		}
	}
	for _, model := range exceptions.Deleted {
		if !singleVariants["pcr-"+model] {
			t.Fatalf("deleted portrait %s must be recorded in priconne-excluded-themes.json", model)
		}
		if _, merged := representative[model]; merged {
			t.Fatalf("deleted portrait %s must not appear in the merge plan", model)
		}
	}
	// Duplicated artwork ships once: the plan keeps the survivor and drops the
	// portraits whose identical art appears under another costume.
	dedupeJSON, err := os.ReadFile("../../../scripts/priconne-deduped-themes.json")
	if err != nil {
		t.Fatal(err)
	}
	var deduped map[string][]string
	if err := json.Unmarshal(dedupeJSON, &deduped); err != nil {
		t.Fatal(err)
	}
	for survivor, dropped := range deduped {
		if len(dropped) == 0 {
			t.Fatalf("dedupe plan keeps %s with nothing dropped", survivor)
		}
		survivorModel := strings.TrimPrefix(survivor, "pcr-")
		if singleVariants[survivor] {
			t.Fatalf("dedupe survivor %s is recorded as excluded", survivor)
		}
		for _, model := range dropped {
			if !singleVariants["pcr-"+model] {
				t.Fatalf("duplicate portrait %s must be recorded in priconne-excluded-themes.json", model)
			}
			if model == survivorModel {
				t.Fatalf("dedupe plan drops the survivor %s", survivor)
			}
		}
	}
	root, err := fs.Sub(assets.FS, "theme")
	if err != nil {
		t.Fatal(err)
	}
	entries, err := fs.ReadDir(root, ".")
	if err != nil {
		t.Fatal(err)
	}
	expected := make(map[string]bool)
	for _, match := range regexp.MustCompile(`spine_full_(\d+)`).FindAllStringSubmatch(string(source), -1) {
		modelID := match[1]
		if _, excluded := exclusions[modelID]; excluded {
			continue
		}
		if _, silhouette := silhouettes[modelID]; silhouette {
			continue
		}
		// A portrait ships either as its own theme or inside the theme of the
		// portrait that opens its merge group; same-name same-skin portraits are
		// the only ones grouped, so a card never shuffles between characters.
		shippedAs := modelID
		if value, merged := representative[modelID]; merged {
			shippedAs = value
		}
		themeName := "pcr-" + shippedAs
		if !singleVariants[themeName] {
			expected[themeName] = true
		}
	}
	if len(expected) == 0 {
		t.Fatal("source manifest contains no portrait models")
	}
	catalogs, catalogErrors := asset.LoadThemeCatalogs(root)
	if len(catalogErrors) > 0 {
		t.Fatal(catalogErrors)
	}
	for _, entry := range entries {
		if !entry.IsDir() || !strings.HasPrefix(entry.Name(), "pcr-") {
			continue
		}
		if !expected[entry.Name()] {
			t.Fatalf("portrait %s is absent from the documented source selection", entry.Name())
		}
		delete(expected, entry.Name())
		t.Run(entry.Name(), func(t *testing.T) {
			manifestJSON, err := fs.ReadFile(root, path.Join(entry.Name(), asset.ManifestName))
			if err != nil {
				t.Fatal(err)
			}
			var manifest []asset.CharacterManifest
			if err := json.Unmarshal(manifestJSON, &manifest); err != nil {
				t.Fatal(err)
			}
			configJSON, err := fs.ReadFile(root, path.Join(entry.Name(), "config.json"))
			if err != nil {
				t.Fatal(err)
			}
			var config asset.CharacterConfig
			if err := json.Unmarshal(configJSON, &config); err != nil {
				t.Fatal(err)
			}
			type configSet struct {
				canvasW int
				canvasH int
				ranges  map[string]asset.PartRange
			}
			sets := []configSet{{canvasW: config.CanvasW, canvasH: config.CanvasH, ranges: config.Ranges}}
			if len(config.Sets) > 0 {
				sets = nil
				for _, configured := range config.Sets {
					sets = append(sets, configSet{canvasW: configured.CanvasW, canvasH: configured.CanvasH, ranges: configured.Ranges})
				}
			}
			multiple := len(sets) > 1
			expectedVariants := 0
			for _, configured := range sets {
				setVariants := 1
				for category, span := range configured.ranges {
					if span.First < 1 || span.Last < span.First || span.Last >= len(manifest) {
						t.Fatalf("%s: invalid manifest range %+v", category, span)
					}
					for i := span.First; i <= span.Last; i++ {
						layer := manifest[i]
						if category == "body" && !strings.HasSuffix(layer.Name, "_body") || category == "eye" && !strings.Contains(layer.Name, "_eye") || category == "mouth" && !strings.Contains(layer.Name, "_mouth") {
							t.Fatalf("%s range includes unrelated layer %q", category, layer.Name)
						}
						if _, _, err := asset.FindImageFile(root, path.Join(entry.Name(), "ren", fmt.Sprint(layer.LayerID))); err != nil {
							t.Fatalf("layer %d image missing: %v", layer.LayerID, err)
						}
						if layer.Left < 0 || layer.Top < 0 || layer.Left+layer.Width > configured.canvasW || layer.Top+layer.Height > configured.canvasH {
							t.Fatalf("layer %d falls outside canvas", layer.LayerID)
						}
					}
					multiple = multiple || span.Last > span.First
					setVariants *= span.Last - span.First + 1
				}
				expectedVariants += setVariants
			}
			catalog := catalogs[entry.Name()]
			if catalog == nil {
				t.Fatal("portrait is absent from the production theme catalog")
			}
			if group, merged := groupMembers[strings.TrimPrefix(entry.Name(), "pcr-")]; merged {
				var shipped []string
				for _, configured := range config.Sets {
					shipped = append(shipped, configured.Name)
				}
				if strings.Join(shipped, ",") != strings.Join(group, ",") {
					t.Fatalf("mutually exclusive sets %v do not match the merge plan %v", shipped, group)
				}
			} else if len(config.Sets) > 0 {
				t.Fatalf("%s ships mutually exclusive sets but is absent from the merge plan", entry.Name())
			}
			if catalog.Variants != expectedVariants {
				t.Fatalf("catalog variants = %d, want additive set total %d", catalog.Variants, expectedVariants)
			}
			if catalog.Variants < minimumPriconneVariants {
				t.Fatalf("Princess Connect themes below %d variants must be excluded, got %d", minimumPriconneVariants, catalog.Variants)
			}
			var first string
			varies := false
			// Seeds 0 and 1 are sufficient here: the deterministic picker uses
			// seed%n, so any slot or set with alternatives selects a different
			// candidate. Parsing one composed SVG per theme keeps the exhaustive
			// asset test below the race build's ten-minute package timeout.
			for seed := 0; seed < 2; seed++ {
				theme, err := asset.AssembleTheme(root, catalog, func(n int) int { return seed % n })
				if err != nil {
					t.Fatal(err)
				}
				svg, err := composer.Compose(composer.ComposeParams{Theme: theme, Seed: fmt.Sprint(seed)})
				if err != nil {
					t.Fatal(err)
				}
				selected := sets[0]
				if len(config.Sets) > 0 {
					selected = sets[seed%len(sets)]
				}
				if got := strings.Count(svg, "<image "); got != len(selected.ranges) {
					t.Fatalf("rendered %d parts, want one per category", got)
				}
				if seed == 0 {
					decoder := xml.NewDecoder(strings.NewReader(svg))
					for {
						_, err := decoder.Token()
						if err == io.EOF {
							break
						}
						if err != nil {
							t.Fatal(err)
						}
					}
				}
				if seed == 0 {
					first = svg
				} else {
					varies = varies || svg != first
				}
			}
			if multiple && !varies {
				t.Fatal("all seeds produced identical SVG despite alternative parts")
			}
		})
	}
	if len(expected) != 0 {
		t.Fatalf("source portraits have not been exported: %v", expected)
	}
}
