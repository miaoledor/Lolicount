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
)

// makePNGDataURI creates a small PNG and returns it as a base64 data URI.
func makePNGDataURI(t *testing.T, w, h int) string {
	t.Helper()
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	img.Set(0, 0, color.RGBA{255, 0, 0, 255})
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		t.Fatal(err)
	}
	return "data:image/png;base64," + base64.StdEncoding.EncodeToString(buf.Bytes())
}

// TestValidateExportName verifies name validation rules.
func TestValidateExportName(t *testing.T) {
	cases := []struct {
		name    string
		wantErr bool
	}{
		{"mytheme", false},
		{"my-theme-2", false},
		{"ABCxyz", false},
		{"", true},
		{"demo", true},
		{"random", true},
		{"under_score", true},
		{"café", true},
		{"hello world", true},
	}
	for _, tc := range cases {
		err := validateExportName(tc.name)
		if tc.wantErr && err == nil {
			t.Errorf("validateExportName(%q): expected error", tc.name)
		}
		if !tc.wantErr && err != nil {
			t.Errorf("validateExportName(%q): unexpected error: %v", tc.name, err)
		}
	}
}

// TestExtractDataURI verifies base64 data URI extraction.
func TestExtractDataURI(t *testing.T) {
	raw, err := extractDataURI("data:image/png;base64," + base64.StdEncoding.EncodeToString([]byte("hello")))
	if err != nil {
		t.Fatalf("extractDataURI: %v", err)
	}
	if string(raw) != "hello" {
		t.Errorf("got %q, want hello", string(raw))
	}

	if _, err := extractDataURI("https://example.com/img.png"); err == nil {
		t.Error("expected error for non-data-URI")
	}
}

// TestExportCardTheme verifies single-layer export produces a valid zip.
func TestExportCardTheme(t *testing.T) {
	pngURI := makePNGDataURI(t, 10, 20)
	layers := []EditorLayer{
		{ID: 1, Category: "lass", Images: []EditorImage{
			{Src: pngURI, Left: 0, Top: 0, Width: 10, Height: 20},
		}},
	}
	imgs, err := decodeAndReencodeImages(layers)
	if err != nil {
		t.Fatalf("decodeAndReencodeImages: %v", err)
	}

	zipBytes, err := exportCardTheme("test-card", layers, imgs)
	if err != nil {
		t.Fatalf("exportCardTheme: %v", err)
	}

	zr, err := zip.NewReader(bytes.NewReader(zipBytes), int64(len(zipBytes)))
	if err != nil {
		t.Fatalf("zip read: %v", err)
	}

	if len(zr.File) != 1 {
		t.Fatalf("zip entries = %d, want 1", len(zr.File))
	}
	if !strings.HasPrefix(zr.File[0].Name, "test-card/0.") {
		t.Errorf("zip entry = %q, want test-card/0.<ext>", zr.File[0].Name)
	}
}

// TestExportCharacterTheme verifies multi-layer export produces ren.json +
// config.json + images.
func TestExportCharacterTheme(t *testing.T) {
	pngURI := makePNGDataURI(t, 50, 50)
	layers := []EditorLayer{
		{ID: 1, Category: "lass", Images: []EditorImage{
			{Src: pngURI, Left: 0, Top: 0, Width: 50, Height: 50},
		}},
		{ID: 2, Category: "eye", Images: []EditorImage{
			{Src: pngURI, Left: 10, Top: 20, Width: 30, Height: 15},
			{Src: pngURI, Left: 10, Top: 20, Width: 30, Height: 15},
		}},
	}
	imgs, err := decodeAndReencodeImages(layers)
	if err != nil {
		t.Fatalf("decodeAndReencodeImages: %v", err)
	}

	canvas := EditorCanvas{Width: 500, Height: 800}
	zipBytes, err := exportCharacterTheme("test-char", canvas, nil, layers, imgs)
	if err != nil {
		t.Fatalf("exportCharacterTheme: %v", err)
	}

	zr, err := zip.NewReader(bytes.NewReader(zipBytes), int64(len(zipBytes)))
	if err != nil {
		t.Fatalf("zip read: %v", err)
	}

	// Should have: ren.json + config.json + 3 images (1 + 2) = 5 entries
	if len(zr.File) != 5 {
		t.Fatalf("zip entries = %d, want 5", len(zr.File))
	}

	// Verify ren.json exists and is valid JSON
	var manifest []manifestEntry
	for _, f := range zr.File {
		if strings.HasSuffix(f.Name, "ren.json") {
			rc, _ := f.Open()
			defer rc.Close()
			if err := json.NewDecoder(rc).Decode(&manifest); err != nil {
				t.Fatalf("decode ren.json: %v", err)
			}
		}
	}
	if len(manifest) != 3 {
		t.Errorf("manifest entries = %d, want 3", len(manifest))
	}

	// Verify config.json exists and has ranges
	for _, f := range zr.File {
		if strings.HasSuffix(f.Name, "config.json") {
			rc, _ := f.Open()
			defer rc.Close()
			var cfg map[string]any
			if err := json.NewDecoder(rc).Decode(&cfg); err != nil {
				t.Fatalf("decode config.json: %v", err)
			}
			if cfg["canvasW"] != float64(500) {
				t.Errorf("canvasW = %v, want 500", cfg["canvasW"])
			}
			ranges, ok := cfg["ranges"].(map[string]any)
			if !ok || len(ranges) < 2 {
				t.Errorf("ranges missing or too few: %v", ranges)
			}
		}
	}
}

// TestBuildRanges verifies range computation from manifest.
func TestBuildRanges(t *testing.T) {
	manifest := []manifestEntry{
		{Name: "lass", LayerID: 0},
		{Name: "lass", LayerID: 1},
		{Name: "eye", LayerID: 2},
		{Name: "eye", LayerID: 3},
		{Name: "eye", LayerID: 4},
		{Name: "mouth", LayerID: 5},
	}
	ranges := buildRanges(manifest)
	if ranges["lass"].First != 0 || ranges["lass"].Last != 1 {
		t.Errorf("lass range = [%d,%d], want [0,1]", ranges["lass"].First, ranges["lass"].Last)
	}
	if ranges["eye"].First != 2 || ranges["eye"].Last != 4 {
		t.Errorf("eye range = [%d,%d], want [2,4]", ranges["eye"].First, ranges["eye"].Last)
	}
	if ranges["mouth"].First != 5 || ranges["mouth"].Last != 5 {
		t.Errorf("mouth range = [%d,%d], want [5,5]", ranges["mouth"].First, ranges["mouth"].Last)
	}
}
