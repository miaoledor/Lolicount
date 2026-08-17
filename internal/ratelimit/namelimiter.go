package ratelimit

import (
	"sync"
	"time"
)

// NameLimiter is a per-name sliding-window counter with a 1-second
// window. When a name exceeds perSec requests within the current
// window, Allow returns false — the caller must degrade to read-only
// (return the current count WITHOUT incrementing) per AGENTS.md Iron
// Rule 3. This is deliberately NOT a 429: a 429 on an embedded counter
// SVG would render a broken image on the referrer's page.
type NameLimiter struct {
	mu      sync.Mutex
	windows map[string]*slidingWindow
	perSec  int
}

type slidingWindow struct {
	count  int
	expiry time.Time
}

// NewNameLimiter builds a name limiter allowing perSec increments per
// second per name.
func NewNameLimiter(perSec int) *NameLimiter {
	return &NameLimiter{
		windows: make(map[string]*slidingWindow),
		perSec:  perSec,
	}
}

// Allow returns true if the name is within its per-second quota. A false
// return means the caller must serve read-only (no increment).
func (l *NameLimiter) Allow(name string, now time.Time) bool {
	l.mu.Lock()
	defer l.mu.Unlock()
	w, ok := l.windows[name]
	if !ok || now.After(w.expiry) {
		l.windows[name] = &slidingWindow{count: 1, expiry: now.Add(time.Second)}
		return true
	}
	if w.count >= l.perSec {
		return false
	}
	w.count++
	return true
}
