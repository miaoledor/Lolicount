package asset

import (
	"bytes"
	"image"
	"image/png"
	"testing"
	"testing/fstest"

	"github.com/miaoledor/lolicount/internal/imgcore/render"
)

// TestCharacterLayerDimensions verifies manifest display sizes take
// precedence while older manifests can fall back to decoded dimensions.
func TestCharacterLayerDimensions(t *testing.T) {
	tests := []struct {
		name       string
		manifestW  int
		manifestH  int
		wantWidth  int
		wantHeight int
	}{
		{name: "manifest dimensions", manifestW: 40, manifestH: 50, wantWidth: 40, wantHeight: 50},
		{name: "decoded fallback", wantWidth: 4, wantHeight: 5},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			width, height := characterLayerDimensions(
				CharacterManifest{Width: tt.manifestW, Height: tt.manifestH},
				4,
				5,
			)
			if width != tt.wantWidth || height != tt.wantHeight {
				t.Fatalf("dimensions = %dx%d, want %dx%d", width, height, tt.wantWidth, tt.wantHeight)
			}
		})
	}
}

// TestValidCharacterConfig verifies that configured sets take precedence
// over legacy top-level layout fields, matching the catalog loader.
func TestValidCharacterConfig(t *testing.T) {
	validRange := map[string]PartRange{"body": {First: 0, Last: 0}}
	tests := []struct {
		name   string
		config CharacterConfig
		want   bool
	}{
		{
			name:   "classic layout",
			config: CharacterConfig{CanvasW: 100, CanvasH: 200, Ranges: validRange},
			want:   true,
		},
		{
			name: "set layout",
			config: CharacterConfig{Sets: []CharacterSet{
				{Name: "default", CanvasW: 100, CanvasH: 200, Ranges: validRange},
				{Name: "alternate", CanvasW: 120, CanvasH: 220, Ranges: validRange},
			}},
			want: true,
		},
		{
			name: "invalid set overrides classic layout",
			config: CharacterConfig{
				CanvasW: 100,
				CanvasH: 200,
				Ranges:  validRange,
				Sets: []CharacterSet{
					{Name: "broken", CanvasH: 200, Ranges: validRange},
				},
			},
		},
		{name: "empty config"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := validCharacterConfig(&tt.config); got != tt.want {
				t.Fatalf("validCharacterConfig() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestLoadCharacterCatalogPreservesManifestDimensions verifies the lazy
// catalog keeps resized layer dimensions exported by theme tooling.
func TestLoadCharacterCatalogPreservesManifestDimensions(t *testing.T) {
	fsys := fstest.MapFS{
		"hero/ren.json": {
			Data: []byte(`[{"left":2,"top":3,"width":40,"height":50,"layer_id":1}]`),
		},
		"hero/config.json": {
			Data: []byte(`{"canvasW":100,"canvasH":200,"ranges":{"body":{"first":0,"last":0}}}`),
		},
		"hero/ren/1.png": {Data: encodeTestPNG(t, 4, 5)},
	}

	catalog, err := loadCharacterCatalog(fsys, "hero")
	if err != nil {
		t.Fatalf("loadCharacterCatalog failed: %v", err)
	}
	if len(catalog.Slots) != 1 || len(catalog.Slots[0].Candidates) != 1 {
		t.Fatalf("catalog slots = %#v, want one slot with one candidate", catalog.Slots)
	}
	candidate := catalog.Slots[0].Candidates[0]
	if candidate.Width != 40 || candidate.Height != 50 {
		t.Fatalf("candidate dimensions = %dx%d, want 40x50", candidate.Width, candidate.Height)
	}

	assembled, err := AssembleTheme(fsys, catalog, func(int) int { return 0 })
	if err != nil {
		t.Fatalf("AssembleTheme failed: %v", err)
	}
	group, ok := assembled.Layers[0].(*render.GroupLayer)
	if !ok || len(group.Parts) != 1 {
		t.Fatalf("assembled layer = %#v, want one group part", assembled.Layers)
	}
	if group.Parts[0].Width != 40 || group.Parts[0].Height != 50 {
		t.Fatalf("assembled dimensions = %dx%d, want 40x50", group.Parts[0].Width, group.Parts[0].Height)
	}
}

func encodeTestPNG(t *testing.T, width, height int) []byte {
	t.Helper()
	img := image.NewNRGBA(image.Rect(0, 0, width, height))
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		t.Fatalf("encode PNG: %v", err)
	}
	return buf.Bytes()
}
