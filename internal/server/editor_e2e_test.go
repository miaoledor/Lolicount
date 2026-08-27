package server

import (
	"archive/zip"
	"bytes"
	"encoding/base64"
	"encoding/json"
	"image"
	"image/color"
	"image/png"
	"strings"
	"testing"

	"github.com/miaoledor/lolicount/internal/imgcore/composer"
	"github.com/miaoledor/lolicount/internal/imgcore/theme"
)

// makePNGBytes creates a small PNG and returns its raw bytes.
func makePNGBytes(t *testing.T, w, h int) []byte {
	t.Helper()
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	img.Set(0, 0, color.RGBA{255, 0, 0, 255})
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		t.Fatal(err)
	}
	return buf.Bytes()
}

// makeDataURI wraps raw bytes as a base64 data URI.
func makeDataURI(raw []byte) string {
	return "data:image/png;base64," + base64.StdEncoding.EncodeToString(raw)
}

// TestE2ECardThemeFlow simulates the full editor flow for a card theme:
// build request → preview render → export zip → verify structure.
func TestE2ECardThemeFlow(t *testing.T) {
	pngData := makePNGBytes(t, 100, 200)
	uri := makeDataURI(pngData)

	req := &EditorRequest{
		Name:   "test-card-e2e",
		Canvas: EditorCanvas{Width: 100, Height: 200},
		Layers: []EditorLayer{
			{ID: 1, Category: "lass", ZIndex: 0, Images: []EditorImage{
				{Src: uri, Left: 0, Top: 0, Width: 100, Height: 200},
			}},
		},
		Text:  "12345",
		FSize: 50,
	}

	// Step 1: Preview render
	t2, err := buildEditorTheme(req)
	if err != nil {
		t.Fatalf("buildEditorTheme: %v", err)
	}
	svg, err := composer.Compose(composer.ComposeParams{Theme: t2, Seed: "e2e", CountText: "12345"})
	if err != nil {
		t.Fatalf("compose: %v", err)
	}
	if !strings.Contains(svg, "<svg") {
		t.Error("SVG missing <svg tag")
	}
	if !strings.Contains(svg, "<image") {
		t.Error("SVG missing <image element")
	}

	// Step 2: Export
	if err := validateExportName(req.Name); err != nil {
		t.Fatalf("validateExportName: %v", err)
	}
	imgs, err := decodeAndReencodeImages(req.Layers)
	if err != nil {
		t.Fatalf("decodeAndReencodeImages: %v", err)
	}
	zipBytes, err := exportCardTheme(req.Name, req.Layers, imgs)
	if err != nil {
		t.Fatalf("exportCardTheme: %v", err)
	}

	// Step 3: Verify zip structure
	zr, err := zip.NewReader(bytes.NewReader(zipBytes), int64(len(zipBytes)))
	if err != nil {
		t.Fatalf("zip read: %v", err)
	}
	if len(zr.File) != 1 {
		t.Errorf("zip entries = %d, want 1", len(zr.File))
	}
	if !strings.HasPrefix(zr.File[0].Name, "test-card-e2e/0.") {
		t.Errorf("zip entry = %q", zr.File[0].Name)
	}
}

// TestE2ECharacterThemeFlow simulates the full editor flow for a
// character theme: multi-layer → preview → export → verify ren.json +
// config.json + images.
func TestE2ECharacterThemeFlow(t *testing.T) {
	pngData := makePNGBytes(t, 50, 50)
	uri := makeDataURI(pngData)

	req := &EditorRequest{
		Name:   "test-char-e2e",
		Canvas: EditorCanvas{Width: 500, Height: 800},
		Display: &theme.DisplayConfig{Size: 400},
		Layers: []EditorLayer{
			{ID: 1, Category: "lass", ZIndex: 0, Images: []EditorImage{
				{Src: uri, Left: 0, Top: 0, Width: 50, Height: 50},
			}},
			{ID: 2, Category: "eye", ZIndex: 1, Images: []EditorImage{
				{Src: uri, Left: 10, Top: 20, Width: 30, Height: 15},
				{Src: uri, Left: 10, Top: 20, Width: 30, Height: 15},
			}},
			{ID: 3, Category: "mouth", ZIndex: 2, Images: []EditorImage{
				{Src: uri, Left: 20, Top: 40, Width: 20, Height: 10},
			}},
		},
		Text:  "42",
		FSize: 30,
	}

	// Step 1: Preview render
	t2, err := buildEditorTheme(req)
	if err != nil {
		t.Fatalf("buildEditorTheme: %v", err)
	}
	svg, err := composer.Compose(composer.ComposeParams{Theme: t2, Seed: "e2e", CountText: "42"})
	if err != nil {
		t.Fatalf("compose: %v", err)
	}
	if !strings.Contains(svg, "<svg") {
		t.Error("SVG missing <svg tag")
	}

	// Step 2: Export
	imgs, err := decodeAndReencodeImages(req.Layers)
	if err != nil {
		t.Fatalf("decodeAndReencodeImages: %v", err)
	}
	zipBytes, err := exportCharacterTheme(req.Name, req.Canvas, req.Display, req.Layers, imgs)
	if err != nil {
		t.Fatalf("exportCharacterTheme: %v", err)
	}

	// Step 3: Verify zip
	zr, err := zip.NewReader(bytes.NewReader(zipBytes), int64(len(zipBytes)))
	if err != nil {
		t.Fatalf("zip read: %v", err)
	}
	// ren.json + config.json + 4 images = 6
	if len(zr.File) != 7 {
		t.Errorf("zip entries = %d, want 7", len(zr.File))
	}

	// Verify ren.json
	for _, f := range zr.File {
		if strings.HasSuffix(f.Name, "ren.json") {
			rc, _ := f.Open()
			var manifest []manifestEntry
			if err := json.NewDecoder(rc).Decode(&manifest); err != nil {
				t.Fatalf("decode ren.json: %v", err)
			}
			rc.Close()
			if len(manifest) != 4 {
				t.Errorf("manifest entries = %d, want 4", len(manifest))
			}
		}
	}

	// Verify config.json has ranges
	for _, f := range zr.File {
		if strings.HasSuffix(f.Name, "config.json") {
			rc, _ := f.Open()
			var cfg map[string]any
			json.NewDecoder(rc).Decode(&cfg)
			rc.Close()
			ranges, ok := cfg["ranges"].(map[string]any)
			if !ok {
				t.Fatal("config.json missing ranges")
			}
			if len(ranges) != 3 {
				t.Errorf("ranges count = %d, want 3", len(ranges))
			}
		}
	}
}

// TestE2EExportNameValidation verifies that export name validation
// catches all illegal names before export.
func TestE2EExportNameValidation(t *testing.T) {
	illegal := []string{"", "demo", "random", "my theme", "café", "under_score"}
	for _, name := range illegal {
		if err := validateExportName(name); err == nil {
			t.Errorf("validateExportName(%q) should fail", name)
		}
	}
	legal := []string{"mytheme", "my-theme-1", "ABC"}
	for _, name := range legal {
		if err := validateExportName(name); err != nil {
			t.Errorf("validateExportName(%q) should pass: %v", name, err)
		}
	}
}
