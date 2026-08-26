package imgutils

import (
	"testing"

	"github.com/miaoledor/lolicount/internal/imgcore"
)

// TestPRNGDeterminism verifies that the same seed string produces the
// same sequence of random values across separate PRNG instances.
func TestPRNGDeterminism(t *testing.T) {
	p1 := NewPRNG("counter-name")
	p2 := NewPRNG("counter-name")

	r := imgcore.Range{Min: 0, Max: 100}
	for i := 0; i < 20; i++ {
		v1 := p1.FloatRange(r)
		v2 := p2.FloatRange(r)
		if v1 != v2 {
			t.Fatalf("iteration %d: same seed gave different values %v != %v", i, v1, v2)
		}
	}
}

// TestPRNGDifferentSeeds verifies that different seeds produce
// different values (collision is astronomically unlikely with FNV-1a).
func TestPRNGDifferentSeeds(t *testing.T) {
	p1 := NewPRNG("alice")
	p2 := NewPRNG("bob")

	r := imgcore.Range{Min: 0, Max: 1000}
	different := false
	for i := 0; i < 10; i++ {
		if p1.FloatRange(r) != p2.FloatRange(r) {
			different = true
			break
		}
	}
	if !different {
		t.Fatal("different seeds produced identical sequences")
	}
}

// TestFloatRangeFixed verifies that Min == Max returns the fixed value.
func TestFloatRangeFixed(t *testing.T) {
	p := NewPRNG("test")
	v := p.FloatRange(imgcore.Range{Min: 42, Max: 42})
	if v != 42 {
		t.Fatalf("expected 42, got %v", v)
	}
}

// TestFloatRangeBounds verifies that random values stay within [Min, Max).
func TestFloatRangeBounds(t *testing.T) {
	p := NewPRNG("bounds-test")
	r := imgcore.Range{Min: 10, Max: 20}
	for i := 0; i < 100; i++ {
		v := p.FloatRange(r)
		if v < 10 || v >= 20 {
			t.Fatalf("value %v out of bounds [10, 20)", v)
		}
	}
}

// TestWeightedPick verifies that zero-weight items are never selected
// and that the pick falls within the valid index range.
func TestWeightedPick(t *testing.T) {
	p := NewPRNG("weighted-test")
	weights := []float64{0, 0, 5, 0, 3}
	for i := 0; i < 100; i++ {
		idx := p.WeightedPick(weights)
		if idx != 2 && idx != 4 {
			t.Fatalf("picked index %d, expected 2 or 4 (only non-zero weights)", idx)
		}
	}
}

// TestWeightedPickAllZero verifies the degenerate case.
func TestWeightedPickAllZero(t *testing.T) {
	p := NewPRNG("degenerate")
	idx := p.WeightedPick([]float64{0, 0, 0})
	if idx != 0 {
		t.Fatalf("expected 0 for all-zero weights, got %d", idx)
	}
}

// TestWeightedPickEmpty verifies the empty slice case.
func TestWeightedPickEmpty(t *testing.T) {
	p := NewPRNG("empty")
	idx := p.WeightedPick([]float64{})
	if idx != 0 {
		t.Fatalf("expected 0 for empty weights, got %d", idx)
	}
}
