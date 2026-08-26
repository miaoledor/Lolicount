package asset

import (
	"bytes"
	"image"
	"image/color"
	"image/png"
	"testing"
)

// makeTestPNG creates a small PNG image for testing.
func makeTestPNG(t *testing.T, w, h int) []byte {
	t.Helper()
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	img.Set(0, 0, color.RGBA{255, 0, 0, 255})
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		t.Fatal(err)
	}
	return buf.Bytes()
}

// TestReEncodeImagePNG verifies round-trip PNG re-encoding.
func TestReEncodeImagePNG(t *testing.T) {
	raw := makeTestPNG(t, 10, 20)
	encoded, mime, err := ReEncodeImage(raw, EncodePNG)
	if err != nil {
		t.Fatalf("ReEncodeImage failed: %v", err)
	}
	if mime != "image/png" {
		t.Fatalf("expected image/png, got %s", mime)
	}
	img, _, err := image.Decode(bytes.NewReader(encoded))
	if err != nil {
		t.Fatalf("decode re-encoded image: %v", err)
	}
	bounds := img.Bounds()
	if bounds.Dx() != 10 || bounds.Dy() != 20 {
		t.Fatalf("expected 10x20, got %dx%d", bounds.Dx(), bounds.Dy())
	}
}

// TestReEncodeImageWebPFallback verifies WebP format falls back to PNG.
func TestReEncodeImageWebPFallback(t *testing.T) {
	raw := makeTestPNG(t, 5, 5)
	encoded, mime, err := ReEncodeImage(raw, EncodeWebP)
	if err != nil {
		t.Fatalf("ReEncodeImage failed: %v", err)
	}
	if mime != "image/png" {
		t.Fatalf("expected image/png (webp fallback), got %s", mime)
	}
	if len(encoded) == 0 {
		t.Fatal("expected non-empty encoded bytes")
	}
}

// TestReEncodeImageTooLarge verifies size limit enforcement.
func TestReEncodeImageTooLarge(t *testing.T) {
	raw := make([]byte, MaxUploadBytes+1)
	_, _, err := ReEncodeImage(raw, EncodePNG)
	if err == nil {
		t.Fatal("expected error for oversized image")
	}
}

// TestReEncodeImageInvalidData verifies invalid image data is rejected.
func TestReEncodeImageInvalidData(t *testing.T) {
	_, _, err := ReEncodeImage([]byte("not an image"), EncodePNG)
	if err == nil {
		t.Fatal("expected error for invalid image data")
	}
}

// TestFrameIndexFromName verifies frame index parsing.
func TestFrameIndexFromName(t *testing.T) {
	tests := []struct {
		input string
		want  int
	}{
		{"0.png", 0},
		{"1.webp", 1},
		{"10.gif", 10},
		{"abc.png", -1},
		{"-1.png", -1},
		{"1.txt", -1},
	}
	for _, tt := range tests {
		got := FrameIndexFromName(tt.input)
		if got != tt.want {
			t.Errorf("FrameIndexFromName(%q) = %d, want %d", tt.input, got, tt.want)
		}
	}
}
