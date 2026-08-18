package theme

import "testing"

// Lock the render defaults so a casual edit to defaults.go is caught.
// These values are part of the rendering contract (M5.6): uniform image
// size, readable font, consistent spacing.
func TestDefaultsContract(t *testing.T) {
	if DefaultTheme != "lian" {
		t.Errorf("DefaultTheme = %q, want lian", DefaultTheme)
	}
	if DefaultDisplaySize != 400 {
		t.Errorf("DefaultDisplaySize = %d, want 400", DefaultDisplaySize)
	}
	if DefaultFontSize != 16 {
		t.Errorf("DefaultFontSize = %d, want 16", DefaultFontSize)
	}
	if MonoCharWidthFactor != 0.6 {
		t.Errorf("MonoCharWidthFactor = %v, want 0.6", MonoCharWidthFactor)
	}
	if DefaultFontFamily != "monospace" {
		t.Errorf("DefaultFontFamily = %q, want monospace", DefaultFontFamily)
	}
	if DefaultFontColor != "#333" {
		t.Errorf("DefaultFontColor = %q, want #333", DefaultFontColor)
	}
	if TextGapBelowImage != 4 {
		t.Errorf("TextGapBelowImage = %d, want 4", TextGapBelowImage)
	}
}
