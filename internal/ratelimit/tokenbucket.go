// Package ratelimit provides the two independent limiters required by
// AGENTS.md Iron Rule 3:
//
//   - IP-level token buckets (10/s burst, 300/min) that reject with 429.
//   - Name-level sliding-window counters (5/s) that degrade to read-only
//     (return the current count without incrementing) instead of 429.
//
// The two must NOT be unified: IP limits protect the server; name limits
// protect a single embedded counter from a burst without breaking the
// embedding (a 429 would show a broken image on the referrer's page).
package ratelimit

import (
	"sync"
	"time"
)

// tokenBucket is a classic continuous-refill token bucket. It is safe
// for concurrent use. tokens are refilled at rate per second up to burst.
type tokenBucket struct {
	mu         sync.Mutex
	tokens     float64
	last       time.Time
	ratePerSec float64
	burst      float64
}

func newTokenBucket(ratePerSec, burst float64, now time.Time) *tokenBucket {
	return &tokenBucket{
		tokens:     burst,
		last:       now,
		ratePerSec: ratePerSec,
		burst:      burst,
	}
}

// allow consumes one token if available and returns true, else false.
func (b *tokenBucket) allow(now time.Time) bool {
	b.mu.Lock()
	defer b.mu.Unlock()
	elapsed := now.Sub(b.last).Seconds()
	b.tokens += elapsed * b.ratePerSec
	if b.tokens > b.burst {
		b.tokens = b.burst
	}
	b.last = now
	if b.tokens >= 1 {
		b.tokens--
		return true
	}
	return false
}

// idleSince returns the last time the bucket was touched; used by the
// reaper to drop buckets for IPs that have gone quiet.
func (b *tokenBucket) idleSince() time.Time {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.last
}

// IPLimiter applies two independent token buckets per IP: a short-burst
// bucket (per-second rate) and a sustained bucket (per-minute rate). A
// request is allowed only if BOTH buckets have a token. This implements
// the "10/s, 300/min" contract from AGENTS.md.
type IPLimiter struct {
	mu       sync.Mutex
	buckets  map[string]*ipBuckets
	rateSec  float64
	burstSec float64
	rateMin  float64
	burstMin float64
	stop     chan struct{}
}

type ipBuckets struct {
	sec *tokenBucket
	min *tokenBucket
}

// NewIPLimiter builds an IP limiter with perSec (tokens/sec, burst=perSec)
// and perMin (tokens/min, burst=perMin). A background reaper drops idle
// IP buckets (no traffic for 10 min) so long-running servers don't leak
// memory on cardinality growth.
func NewIPLimiter(perSec, perMin int) *IPLimiter {
	l := &IPLimiter{
		buckets:  make(map[string]*ipBuckets),
		rateSec:  float64(perSec),
		burstSec: float64(perSec),
		rateMin:  float64(perMin) / 60.0,
		burstMin: float64(perMin),
		stop:     make(chan struct{}),
	}
	go l.reaper()
	return l
}

// Allow returns true if the IP is within both limits.
func (l *IPLimiter) Allow(ip string, now time.Time) bool {
	l.mu.Lock()
	bs, ok := l.buckets[ip]
	if !ok {
		bs = &ipBuckets{
			sec: newTokenBucket(l.rateSec, l.burstSec, now),
			min: newTokenBucket(l.rateMin, l.burstMin, now),
		}
		l.buckets[ip] = bs
	}
	l.mu.Unlock()
	return bs.sec.allow(now) && bs.min.allow(now)
}

// Stop halts the background reaper. Safe to call multiple times.
func (l *IPLimiter) Stop() {
	select {
	case <-l.stop:
	default:
		close(l.stop)
	}
}

// reaper periodically evicts IP buckets idle for over 10 minutes. Without
// this, every distinct IP ever seen would stay in memory forever.
func (l *IPLimiter) reaper() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()
	const idleTTL = 10 * time.Minute
	for {
		select {
		case <-ticker.C:
			l.evict(time.Now(), idleTTL)
		case <-l.stop:
			return
		}
	}
}

func (l *IPLimiter) evict(now time.Time, idleTTL time.Duration) {
	l.mu.Lock()
	defer l.mu.Unlock()
	for ip, bs := range l.buckets {
		if now.Sub(bs.sec.idleSince()) > idleTTL {
			delete(l.buckets, ip)
		}
	}
}
