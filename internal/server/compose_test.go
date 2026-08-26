package server

import "testing"

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
