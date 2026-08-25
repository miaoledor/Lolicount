package characterthemedrawer

import (
	"bytes"
	"image"
	"image/png"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// A Character with a custom config uses its own canvas dimensions and
// part ranges for both assembly and drawing, instead of the 莲 defaults.
func TestCustomConfigUsedByAssembleAndDraw(t *testing.T) {
	const canvasW, canvasH = 999, 888
	cfg := &CharacterConfig{
		CanvasW: canvasW,
		CanvasH: canvasH,
		Ranges: map[string]partRange{
			"lass":  {1, 1},
			"brow":  {2, 2},
			"eye":   {3, 3},
			"mouth": {4, 4},
			"face":  {5, 5},
		},
	}
	layers := make([]CharacterLayer, 6)
	parts := make(map[int]CharacterPart, 5)
	for i := 1; i <= 5; i++ {
		layers[i] = CharacterLayer{
			Name: "L", Left: 10 * i, Top: 20 * i,
			Width: 30, Height: 40, LayerID: 100 + i,
		}
		parts[100+i] = CharacterPart{
			Left: 10 * i, Top: 20 * i, Width: 30, Height: 40,
			Data: "data:image/png;base64,QQ",
		}
	}
	c := &Character{Layers: layers, Parts: parts, Config: cfg}

	p, err := c.Assemble(rand.New(rand.NewSource(1)))
	if err != nil {
		t.Fatalf("Assemble: %v", err)
	}
	if p.Config != cfg {
		t.Fatal("portrait should carry the custom config")
	}
	layer := Draw(p, 0)
	if !strings.Contains(layer.Fragment, `viewBox="0 0 999 888"`) {
		t.Errorf("Draw should use custom canvas viewBox 999x888: %s", layer.Fragment)
	}
}

// LoadCharacter reads an optional config.json alongside ren.json and
// applies its canvas dimensions + part ranges. A theme without
// config.json falls back to the default 莲 layout.
func TestLoadCharacterReadsConfigJSON(t *testing.T) {
	dir := t.TempDir()
	renDir := filepath.Join(dir, "ren")
	if err := os.MkdirAll(renDir, 0755); err != nil {
		t.Fatal(err)
	}
	// Build a minimal manifest: index 0 placeholder, then 5 part layers.
	manifest := `[
		{"name":"placeholder","left":0,"top":0,"width":0,"height":0,"visible":1,"layer_id":0,"group_layer_id":0},
		{"name":"lass_1","left":0,"top":0,"width":10,"height":10,"visible":1,"layer_id":1,"group_layer_id":1},
		{"name":"brow_2","left":0,"top":0,"width":10,"height":10,"visible":1,"layer_id":2,"group_layer_id":2},
		{"name":"eye_3","left":0,"top":0,"width":10,"height":10,"visible":1,"layer_id":3,"group_layer_id":3},
		{"name":"mouth_4","left":0,"top":0,"width":10,"height":10,"visible":1,"layer_id":4,"group_layer_id":4},
		{"name":"face_5","left":0,"top":0,"width":10,"height":10,"visible":1,"layer_id":5,"group_layer_id":5}
	]`
	if err := os.WriteFile(filepath.Join(dir, "ren.json"), []byte(manifest), 0644); err != nil {
		t.Fatal(err)
	}
	// Write tiny 1x1 transparent PNGs for each layer.
	png1x1 := tinyPNG(t)
	for i := 1; i <= 5; i++ {
		if err := os.WriteFile(filepath.Join(renDir, pngLayerName(i)), png1x1, 0644); err != nil {
			t.Fatal(err)
		}
	}
	// Write a config.json with a custom canvas.
	config := `{"canvasW":1234,"canvasH":5678,"ranges":{"lass":{"first":1,"last":1},"brow":{"first":2,"last":2},"eye":{"first":3,"last":3},"mouth":{"first":4,"last":4},"face":{"first":5,"last":5}}}`
	if err := os.WriteFile(filepath.Join(dir, "config.json"), []byte(config), 0644); err != nil {
		t.Fatal(err)
	}

	ch, err := LoadCharacter(os.DirFS(dir), ".")
	if err != nil {
		t.Fatalf("LoadCharacter: %v", err)
	}
	if ch.Config == nil {
		t.Fatal("config should be loaded, not nil")
	}
	if ch.Config.CanvasW != 1234 || ch.Config.CanvasH != 5678 {
		t.Errorf("canvas: got %dx%d, want 1234x5678", ch.Config.CanvasW, ch.Config.CanvasH)
	}
	// Verify Draw uses the custom canvas.
	p, err := ch.Assemble(rand.New(rand.NewSource(1)))
	if err != nil {
		t.Fatalf("Assemble: %v", err)
	}
	layer := Draw(p, 0)
	if !strings.Contains(layer.Fragment, `viewBox="0 0 1234 5678"`) {
		t.Errorf("Draw should use config canvas viewBox: %s", layer.Fragment)
	}
}

// LoadCharacter without a config.json falls back to defaultConfig.
func TestLoadCharacterFallsBackToDefaultConfig(t *testing.T) {
	dir := t.TempDir()
	renDir := filepath.Join(dir, "ren")
	if err := os.MkdirAll(renDir, 0755); err != nil {
		t.Fatal(err)
	}
	// Minimal manifest with at least one real layer.
	manifest := `[
		{"name":"placeholder","left":0,"top":0,"width":0,"height":0,"visible":1,"layer_id":0,"group_layer_id":0},
		{"name":"lass_1","left":0,"top":0,"width":10,"height":10,"visible":1,"layer_id":1,"group_layer_id":1}
	]`
	if err := os.WriteFile(filepath.Join(dir, "ren.json"), []byte(manifest), 0644); err != nil {
		t.Fatal(err)
	}
	png1x1 := tinyPNG(t)
	if err := os.WriteFile(filepath.Join(renDir, pngLayerName(1)), png1x1, 0644); err != nil {
		t.Fatal(err)
	}
	// No config.json — should fall back to default.
	ch, err := LoadCharacter(os.DirFS(dir), ".")
	if err != nil {
		t.Fatalf("LoadCharacter: %v", err)
	}
	if ch.Config != nil {
		t.Error("Config should be nil when no config.json is present (uses default at runtime)")
	}
	// config() method should return the default.
	if got := ch.config(); got != defaultConfig {
		t.Error("config() should return defaultConfig when Config is nil")
	}
}

// pngLayerName returns the filename for a layer_id PNG.
func pngLayerName(id int) string {
	return intToStr(id) + ".png"
}

// intToStr converts an int to its decimal string without importing strconv
// (kept simple for test helper).
func intToStr(n int) string {
	if n == 0 {
		return "0"
	}
	var buf [20]byte
	i := len(buf)
	for n > 0 {
		i--
		buf[i] = byte('0' + n%10)
		n /= 10
	}
	return string(buf[i:])
}

// tinyPNG returns a valid 1x1 transparent PNG for test fixtures.
func tinyPNG(t *testing.T) []byte {
	t.Helper()
	var buf bytes.Buffer
	img := image.NewNRGBA(image.Rect(0, 0, 1, 1))
	if err := png.Encode(&buf, img); err != nil {
		t.Fatalf("encode png: %v", err)
	}
	return buf.Bytes()
}


// TestDrawUsesDisplayJSON verifies that when a display.json sets a target
// size (height), Draw scales the canvas proportionally: the output height
// equals size and the width follows the PSD canvas aspect ratio (no
// stretching). For a 100x500 canvas with size=400, the output is 80x400.
func TestDrawUsesDisplayJSON(t *testing.T) {
	dir := t.TempDir()
	renDir := filepath.Join(dir, "ren")
	if err := os.MkdirAll(renDir, 0755); err != nil {
		t.Fatal(err)
	}
	manifest := `[
		{"name":"placeholder","left":0,"top":0,"width":0,"height":0,"visible":1,"layer_id":0,"group_layer_id":0},
		{"name":"lass_1","left":0,"top":0,"width":10,"height":10,"visible":1,"layer_id":1,"group_layer_id":1},
		{"name":"brow_2","left":0,"top":0,"width":10,"height":10,"visible":1,"layer_id":2,"group_layer_id":2},
		{"name":"eye_3","left":0,"top":0,"width":10,"height":10,"visible":1,"layer_id":3,"group_layer_id":3},
		{"name":"mouth_4","left":0,"top":0,"width":10,"height":10,"visible":1,"layer_id":4,"group_layer_id":4},
		{"name":"face_5","left":0,"top":0,"width":10,"height":10,"visible":1,"layer_id":5,"group_layer_id":5}
	]`
	if err := os.WriteFile(filepath.Join(dir, "ren.json"), []byte(manifest), 0644); err != nil {
		t.Fatal(err)
	}
	png1x1 := tinyPNG(t)
	for i := 1; i <= 5; i++ {
		if err := os.WriteFile(filepath.Join(renDir, pngLayerName(i)), png1x1, 0644); err != nil {
			t.Fatal(err)
		}
	}
	// Canvas 100x500 (aspect ratio 1:5); display.json size=400 -> 80x400.
	config := `{"canvasW":100,"canvasH":500,"ranges":{"lass":{"first":1,"last":1},"brow":{"first":2,"last":2},"eye":{"first":3,"last":3},"mouth":{"first":4,"last":4},"face":{"first":5,"last":5}}}`
	if err := os.WriteFile(filepath.Join(dir, "config.json"), []byte(config), 0644); err != nil {
		t.Fatal(err)
	}
	display := `{"size":400}`
	if err := os.WriteFile(filepath.Join(dir, "display.json"), []byte(display), 0644); err != nil {
		t.Fatal(err)
	}
	ch, err := LoadCharacter(os.DirFS(dir), ".")
	if err != nil {
		t.Fatalf("LoadCharacter: %v", err)
	}
	if ch.Display == nil || ch.Display.Size != 400 {
		t.Fatalf("display.json not loaded: %+v", ch.Display)
	}
	p, err := ch.Assemble(rand.New(rand.NewSource(1)))
	if err != nil {
		t.Fatalf("Assemble: %v", err)
	}
	if p.Display == nil || p.Display.Size != 400 {
		t.Fatalf("Display not propagated to portrait: %+v", p.Display)
	}
	layer := Draw(p, 0)
	// Proportional: height=400, width=100*400/500=80.
	if !strings.Contains(layer.Fragment, `width="80" height="400"`) {
		t.Errorf("Draw should scale proportionally to 80x400: %s", layer.Fragment)
	}
	if layer.Width != 80 || layer.Height != 400 {
		t.Errorf("layer dims: got %dx%d, want 80x400", layer.Width, layer.Height)
	}
}

// TestDrawUsesDisplayCrop verifies that when display.json sets a crop
// rect, Draw maps only that sub-rectangle to the output viewport,
// trimming blank PSD canvas margins. For a 100x500 canvas with a
// 50x200 crop at (10,50) and size=400, the output is 100x400 with
// viewBox "10 50 50 200".
func TestDrawUsesDisplayCrop(t *testing.T) {
	dir := t.TempDir()
	renDir := filepath.Join(dir, "ren")
	if err := os.MkdirAll(renDir, 0755); err != nil {
		t.Fatal(err)
	}
	manifest := `[
		{"name":"placeholder","left":0,"top":0,"width":0,"height":0,"visible":1,"layer_id":0,"group_layer_id":0},
		{"name":"lass_1","left":0,"top":0,"width":10,"height":10,"visible":1,"layer_id":1,"group_layer_id":1},
		{"name":"brow_2","left":0,"top":0,"width":10,"height":10,"visible":1,"layer_id":2,"group_layer_id":2},
		{"name":"eye_3","left":0,"top":0,"width":10,"height":10,"visible":1,"layer_id":3,"group_layer_id":3},
		{"name":"mouth_4","left":0,"top":0,"width":10,"height":10,"visible":1,"layer_id":4,"group_layer_id":4},
		{"name":"face_5","left":0,"top":0,"width":10,"height":10,"visible":1,"layer_id":5,"group_layer_id":5}
	]`
	if err := os.WriteFile(filepath.Join(dir, "ren.json"), []byte(manifest), 0644); err != nil {
		t.Fatal(err)
	}
	png1x1 := tinyPNG(t)
	for i := 1; i <= 5; i++ {
		if err := os.WriteFile(filepath.Join(renDir, pngLayerName(i)), png1x1, 0644); err != nil {
			t.Fatal(err)
		}
	}
	config := `{"canvasW":100,"canvasH":500,"ranges":{"lass":{"first":1,"last":1},"brow":{"first":2,"last":2},"eye":{"first":3,"last":3},"mouth":{"first":4,"last":4},"face":{"first":5,"last":5}}}`
	if err := os.WriteFile(filepath.Join(dir, "config.json"), []byte(config), 0644); err != nil {
		t.Fatal(err)
	}
	// Crop 50x200 at (10,50); size=400 -> width=50*400/200=100.
	display := `{"size":400,"crop":{"left":10,"top":50,"width":50,"height":200}}`
	if err := os.WriteFile(filepath.Join(dir, "display.json"), []byte(display), 0644); err != nil {
		t.Fatal(err)
	}
	ch, err := LoadCharacter(os.DirFS(dir), ".")
	if err != nil {
		t.Fatalf("LoadCharacter: %v", err)
	}
	if ch.Display == nil || ch.Display.Crop == nil {
		t.Fatalf("crop not loaded: %+v", ch.Display)
	}
	p, err := ch.Assemble(rand.New(rand.NewSource(1)))
	if err != nil {
		t.Fatalf("Assemble: %v", err)
	}
	layer := Draw(p, 0)
	// viewBox should be the crop rect, dims 100x400.
	if !strings.Contains(layer.Fragment, `viewBox="10 50 50 200"`) {
		t.Errorf("Draw should use crop viewBox: %s", layer.Fragment)
	}
	if !strings.Contains(layer.Fragment, `width="100" height="400"`) {
		t.Errorf("Draw should scale crop to 100x400: %s", layer.Fragment)
	}
	if layer.Width != 100 || layer.Height != 400 {
		t.Errorf("layer dims: got %dx%d, want 100x400", layer.Width, layer.Height)
	}
}
