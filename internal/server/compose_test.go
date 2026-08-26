package server

import "testing"

// TestFrameIndexForCount verifies the M2.5 frame-advancement rule:
// display frame[(count+1) % size].
func TestFrameIndexForCount(t *testing.T) {
	cases := []struct {
		name  string
		count int64
		size  int
		want  int
	}{
		{"size1", 5, 1, 0},
		{"size3-count0", 0, 3, 1},
		{"size3-count1", 1, 3, 2},
		{"size3-count2", 2, 3, 0},
		{"size3-count5", 5, 3, 0},
		{"size10-count9", 9, 10, 0},
		{"size10-count8", 8, 10, 9},
		{"size0-guard", 5, 0, 0},
		{"size-negative", 5, -1, 0},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := frameIndexForCount(tc.count, tc.size)
			if got != tc.want {
				t.Errorf("frameIndexForCount(%d,%d) = %d, want %d", tc.count, tc.size, got, tc.want)
			}
		})
	}
}

// TestScaleOrOne verifies the scale fallback.
func TestScaleOrOne(t *testing.T) {
	if v := scaleOrOne(0); v != 1 {
		t.Errorf("scaleOrOne(0) = %v, want 1", v)
	}
	if v := scaleOrOne(-1); v != 1 {
		t.Errorf("scaleOrOne(-1) = %v, want 1", v)
	}
	if v := scaleOrOne(1.5); v != 1.5 {
		t.Errorf("scaleOrOne(1.5) = %v, want 1.5", v)
	}
}
