// Package imgutils provides shared helpers used across imgcore render
// sub-packages. prng.go implements a seed-driven PRNG that produces
// deterministic output for the same seed (the counter name), so the
// same counter renders the same random layer selection across requests.
package imgutils

import (
	"math/rand/v2"

	"github.com/miaoledor/lolicount/internal/imgcore"
)

// PRNG is a deterministic random number generator seeded by a string
// (typically the counter name). It implements imgcore.PRNGSource so
// render layers can resolve Range values and make weighted picks
// without importing math/rand directly.
type PRNG struct {
	rng *rand.Rand
}

// NewPRNG creates a PRNG seeded from a string via FNV-1a hashing.
// The same seed string always produces the same sequence of random
// values, so the same counter name yields the same random portrait.
func NewPRNG(seed string) *PRNG {
	return &PRNG{rng: rand.New(rand.NewPCG(fnv1a(seed), 0))}
}

// FloatRange resolves a Range to a concrete float64. When Min == Max
// the fixed value is returned directly; otherwise a uniform random
// value in [Min, Max) is generated.
func (p *PRNG) FloatRange(r imgcore.Range) float64 {
	if r.Min == r.Max {
		return r.Min
	}
	return r.Min + p.rng.Float64()*(r.Max-r.Min)
}

// WeightedPick returns the index of a randomly selected item from a
// slice of non-negative weights. Weights of zero are never selected.
// Returns 0 when the slice is empty or all weights are zero.
func (p *PRNG) WeightedPick(weights []float64) int {
	var total float64
	for _, w := range weights {
		if w > 0 {
			total += w
		}
	}
	if total <= 0 {
		return 0
	}
	target := p.rng.Float64() * total
	var acc float64
	for i, w := range weights {
		if w <= 0 {
			continue
		}
		acc += w
		if target < acc {
			return i
		}
	}
	return len(weights) - 1
}

// fnv1a computes the 64-bit FNV-1a hash of s, producing a deterministic
// integer seed from a string.
func fnv1a(s string) uint64 {
	const offset = 14695981039346656037
	const prime = 1099511628211
	h := uint64(offset)
	for i := 0; i < len(s); i++ {
		h ^= uint64(s[i])
		h *= prime
	}
	return h
}
