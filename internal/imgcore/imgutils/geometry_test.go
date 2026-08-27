package imgutils

import "testing"
func TestDisplaySize(t *testing.T) {
	if DisplaySize(0) != DefaultDisplaySize {
		t.Errorf("scale=0 should use default %d", DefaultDisplaySize)
	}
	if DisplaySize(2) != 800 {
		t.Errorf("scale=2 should give 800, got %d", DisplaySize(2))
	}
	if DisplaySize(-1) != DefaultDisplaySize {
		t.Errorf("negative scale should use default")
	}
}

func TestScaledDimsPreservesAspect(t *testing.T) {
	// 10x20 frame, longest edge 20 -> 400, ratio 20 -> 200x400
	w, h := ScaledDims(10, 20, 400)
	if w != 200 || h != 400 {
		t.Errorf("ScaledDims(10,20,400) = %dx%d, want 200x400", w, h)
	}
}

func TestScaledDimsWideFrame(t *testing.T) {
	// 2000x1000 -> longest 2000 -> 400, ratio 0.2 -> 400x200
	w, h := ScaledDims(2000, 1000, 400)
	if w != 400 || h != 200 {
		t.Errorf("ScaledDims(2000,1000,400) = %dx%d, want 400x200", w, h)
	}
}

func TestScaledDimsZeroReturnsOriginal(t *testing.T) {
	w, h := ScaledDims(0, 0, 400)
	if w != 0 || h != 0 {
		t.Errorf("zero dims should pass through, got %dx%d", w, h)
	}
}

func TestScaledDimsMinOne(t *testing.T) {
	// Very small but positive dims clamp to at least 1.
	w, h := ScaledDims(1, 1, 400)
	if w < 1 || h < 1 {
		t.Errorf("clamped dims should be >= 1, got %dx%d", w, h)
	}
}
