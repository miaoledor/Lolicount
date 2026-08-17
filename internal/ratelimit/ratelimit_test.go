package ratelimit

import (
	"testing"
	"time"
)

func TestTokenBucketBurstThenRefill(t *testing.T) {
	tb := newTokenBucket(2, 2, time.Unix(0, 0)) // 2/s, burst 2
	now := time.Unix(0, 0)
	if !tb.allow(now) {
		t.Fatal("first token should be allowed")
	}
	if !tb.allow(now) {
		t.Fatal("second token should be allowed (burst)")
	}
	if tb.allow(now) {
		t.Fatal("third token within same instant should be denied")
	}
	// After 1s, 2 tokens refill.
	if !tb.allow(now.Add(time.Second)) {
		t.Fatal("token should refill after 1s")
	}
}

func TestIPLimiterDualBuckets(t *testing.T) {
	l := NewIPLimiter(2, 4) // 2/s burst, 4/min burst
	now := time.Unix(0, 0)
	// Burst 2 at t0 (per-sec bucket), but per-min allows 4.
	for i := 0; i < 2; i++ {
		if !l.Allow("1.2.3.4", now) {
			t.Fatalf("req %d should pass", i)
		}
	}
	// 3rd at t0: per-sec exhausted.
	if l.Allow("1.2.3.4", now) {
		t.Fatal("3rd req same second should be denied")
	}
	// After 1s: per-sec refills 2, but per-min bucket now has 2 left
	// (used 2 of 4). Two more should pass.
	pass := 0
	for i := 0; i < 4; i++ {
		if l.Allow("1.2.3.4", now.Add(time.Second)) {
			pass++
		}
	}
	if pass != 2 {
		t.Errorf("after refill, expected 2 passes, got %d", pass)
	}
}

func TestIPLimiterIsolatesIPs(t *testing.T) {
	l := NewIPLimiter(1, 100)
	now := time.Unix(0, 0)
	if !l.Allow("a", now) {
		t.Fatal("IP a first should pass")
	}
	if l.Allow("a", now) {
		t.Fatal("IP a second should be denied")
	}
	if !l.Allow("b", now) {
		t.Fatal("IP b should be independent")
	}
}

func TestNameLimiterDegrades(t *testing.T) {
	l := NewNameLimiter(2)
	now := time.Unix(0, 0)
	if !l.Allow("x", now) {
		t.Fatal("first should pass")
	}
	if !l.Allow("x", now) {
		t.Fatal("second should pass")
	}
	if l.Allow("x", now) {
		t.Fatal("third should be denied (degrade)")
	}
	// New window resets.
	if !l.Allow("x", now.Add(time.Second+time.Millisecond)) {
		t.Fatal("should pass after window reset")
	}
}

func TestNameLimiterIsolatesNames(t *testing.T) {
	l := NewNameLimiter(1)
	now := time.Unix(0, 0)
	if !l.Allow("a", now) {
		t.Fatal("a first should pass")
	}
	if l.Allow("a", now) {
		t.Fatal("a second should be denied")
	}
	if !l.Allow("b", now) {
		t.Fatal("b should be independent")
	}
}
