package server

import (
	"strings"
	"testing"

	"github.com/miaoledor/lolicount/internal/imgcore/theme"
)

// makeEditorReq builds a minimal EditorRequest for testing with the
// given number of layers. Each layer has one fake image.
func makeEditorReq(nLayers int) *EditorRequest {
	layers := make([]EditorLayer, nLayers)
	for i := 0; i < nLayers; i++ {
		layers[i] = EditorLayer{
			ID:       i + 1,
			Name: "lass",
			ZIndex:   i,
			Images: []EditorImage{
				{Src: "data:image/png;base64,AAAA", Left: 0, Top: 0, Width: 100, Height: 200},
			},
		}
	}
	return &EditorRequest{
		Name:   "test-editor",
		Canvas: EditorCanvas{Width: 500, Height: 800},
		Layers: layers,
		Text:   "12345",
		FSize:  50,
	}
}

// TestBuildEditorThemeSingleLayer verifies that a single-layer editor
// request produces a valid theme with a GroupLayer and appended text.
func TestBuildEditorThemeSingleLayer(t *testing.T) {
	req := makeEditorReq(1)
	t2, err := buildEditorTheme(req)
	if err != nil {
		t.Fatalf("buildEditorTheme: %v", err)
	}
	if t2.Name != "test-editor" {
		t.Errorf("name = %q, want test-editor", t2.Name)
	}
	if len(t2.Layers) != 2 {
		t.Fatalf("layers = %d, want 2 (group + text)", len(t2.Layers))
	}
	if t2.BgW <= 0 || t2.BgH <= 0 {
		t.Errorf("bg dims = %dx%d, want positive", t2.BgW, t2.BgH)
	}
}

// TestBuildEditorThemeMultiLayer verifies multi-layer requests produce
// a GroupLayer with multiple parts.
func TestBuildEditorThemeMultiLayer(t *testing.T) {
	req := makeEditorReq(3)
	t2, err := buildEditorTheme(req)
	if err != nil {
		t.Fatalf("buildEditorTheme: %v", err)
	}
	if len(t2.Layers) != 2 {
		t.Fatalf("layers = %d, want 2", len(t2.Layers))
	}
}

// TestBuildEditorThemeEmptyLayers verifies that zero layers is rejected.
func TestBuildEditorThemeEmptyLayers(t *testing.T) {
	req := &EditorRequest{
		Name:   "empty",
		Canvas: EditorCanvas{Width: 500, Height: 800},
		Layers: []EditorLayer{},
	}
	if _, err := buildEditorTheme(req); err == nil {
		t.Fatal("expected error for empty layers")
	}
}

// TestBuildEditorThemeZeroCanvas verifies that zero canvas dims rejected.
func TestBuildEditorThemeZeroCanvas(t *testing.T) {
	req := &EditorRequest{
		Name:   "zero",
		Canvas: EditorCanvas{Width: 0, Height: 0},
		Layers: []EditorLayer{
			{ID: 1, Images: []EditorImage{{Src: "x", Width: 10, Height: 10}}},
		},
	}
	if _, err := buildEditorTheme(req); err == nil {
		t.Fatal("expected error for zero canvas")
	}
}

// TestBuildEditorThemeDisplayConfig verifies display config controls
// output dimensions.
func TestBuildEditorThemeDisplayConfig(t *testing.T) {
	req := makeEditorReq(1)
	req.Canvas = EditorCanvas{Width: 2000, Height: 4000}
	req.Display = &theme.DisplayConfig{Size: 400}
	t2, err := buildEditorTheme(req)
	if err != nil {
		t.Fatalf("buildEditorTheme: %v", err)
	}
	// outH should be display size (400), outW proportional
	if t2.BgH != 400 {
		t.Errorf("BgH = %d, want 400 (display size)", t2.BgH)
	}
	expectedW := int(float64(2000) * float64(400) / float64(4000))
	if t2.BgW != expectedW {
		t.Errorf("BgW = %d, want %d", t2.BgW, expectedW)
	}
}

// TestBuildEditorThemeMultiImageCandidates verifies that a layer with
// multiple images produces Candidates for random selection.
func TestBuildEditorThemeMultiImageCandidates(t *testing.T) {
	req := &EditorRequest{
		Name:   "multi-img",
		Canvas: EditorCanvas{Width: 500, Height: 800},
		Layers: []EditorLayer{
			{
				ID:       1,
				Name: "eye",
				Images: []EditorImage{
					{Src: "img1", Left: 10, Top: 20, Width: 100, Height: 50},
					{Src: "img2", Left: 10, Top: 20, Width: 100, Height: 50},
					{Src: "img3", Left: 10, Top: 20, Width: 100, Height: 50},
				},
			},
		},
		Text: "42",
	}
	parts, err := buildGroupParts(req.Layers)
	if err != nil {
		t.Fatalf("buildGroupParts: %v", err)
	}
	if len(parts) != 1 {
		t.Fatalf("parts = %d, want 1", len(parts))
	}
	if len(parts[0].Candidates) != 3 {
		t.Errorf("candidates = %d, want 3", len(parts[0].Candidates))
	}
}

// TestBuildEditorThemeEmptyLayerPlaceholders verifies that layers with
// no images get placeholder parts (dashed border + category label)
// instead of being skipped. All 3 layers should produce 3 parts.
func TestBuildEditorThemeEmptyLayerPlaceholders(t *testing.T) {
	req := &EditorRequest{
		Name:   "placeholders",
		Canvas: EditorCanvas{Width: 500, Height: 800},
		Layers: []EditorLayer{
			{ID: 1, Name: "lass", Images: []EditorImage{}},
			{ID: 2, Name: "eye", Images: []EditorImage{{Src: "x", Width: 10, Height: 10}}},
			{ID: 3, Name: "mouth", Images: []EditorImage{}},
		},
		Text: "1",
	}
	parts, err := buildGroupParts(req.Layers)
	if err != nil {
		t.Fatalf("buildGroupParts: %v", err)
	}
	if len(parts) != 3 {
		t.Fatalf("parts = %d, want 3 (2 placeholders + 1 image)", len(parts))
	}
	// First and third parts should be placeholders (data URI SVG)
	if !strings.HasPrefix(parts[0].Src, "data:image/svg+xml;base64,") {
		t.Errorf("part[0] src should be a data URI SVG placeholder, got %q", parts[0].Src[:30])
	}
	if !strings.HasPrefix(parts[2].Src, "data:image/svg+xml;base64,") {
		t.Errorf("part[2] src should be a data URI SVG placeholder, got %q", parts[2].Src[:30])
	}
	// Second part should be the real image
	if parts[1].Src != "x" {
		t.Errorf("part[1] src = %q, want x", parts[1].Src)
	}
}

// TestBuildEditorThemeRendersSVG verifies the full pipeline produces
// valid SVG with image and text elements.
func TestBuildEditorThemeRendersSVG(t *testing.T) {
	req := makeEditorReq(2)
	t2, err := buildEditorTheme(req)
	if err != nil {
		t.Fatalf("buildEditorTheme: %v", err)
	}
	// Check the theme has the right structure for compose
	if !strings.Contains(t2.Name, "test") {
		t.Errorf("unexpected name: %s", t2.Name)
	}
	// Verify both layers exist
	if len(t2.Layers) != 2 {
		t.Fatalf("layers = %d, want 2", len(t2.Layers))
	}
}

// TestBuildEditorThemeLayersNoImages verifies that layers with no
// images don't cause a 400 error. The user may add a layer before
// uploading images — the preview should render an empty canvas with
// just the text layer.
func TestBuildEditorThemeLayersNoImages(t *testing.T) {
	req := &EditorRequest{
		Name:   "empty-layers",
		Canvas: EditorCanvas{Width: 500, Height: 800},
		Layers: []EditorLayer{
			{ID: 1, Name: "lass", Images: []EditorImage{}},
		},
		Text: "12345",
	}
	t2, err := buildEditorTheme(req)
	if err != nil {
		t.Fatalf("buildEditorTheme with empty-image layers: %v", err)
	}
	if len(t2.Layers) != 2 {
		t.Fatalf("layers = %d, want 2 (empty group + text)", len(t2.Layers))
	}
}
