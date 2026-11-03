package asset

import "testing"

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
