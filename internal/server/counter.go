package server

import (
	"strings"
	"time"

	"github.com/gofiber/fiber/v3"

	"github.com/miaoledor/lolicount/internal/theme"
)

// counterHandler renders GET /@:name (and the /get/@:name alias).
//
// M4 scope: name-level rate limiting with read-only degradation. A name
// exceeding 5 req/s is served its current count WITHOUT incrementing
// (AGENTS.md Iron Rule 3) — never 429, which would break the embedded
// image on the referrer's page. IP-level 429 is handled by the
// ipRateLimit middleware on the route group.
//
// Cache-Control: no-store for all real counters (Iron Rule 1); only the
// reserved "demo" name (fixed value, never persisted) gets long cache.
func (s *Server) counterHandler(c fiber.Ctx) error {
	// Fiber/fasthttp route params can reference a per-request buffer that
	// the runtime reuses across requests. The name is later stored as a
	// map key inside counter.Buffer, so it MUST outlive the request:
	// clone it into an owned string before anything caches it.
	name := strings.Clone(c.Params("name"))
	if name == "" {
		return fiber.NewError(fiber.StatusBadRequest, "missing name")
	}

	q, err := parseParams(c)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}

	th, err := resolveTheme(s.themes, q.Theme)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}
	size := th.Size()

	var rp theme.RenderParams
	switch {
	case name == "demo":
		// Reserved: never count, long cache (Iron Rule 1).
		if q.Number > 0 {
			rp = theme.RenderParams{Count: q.Number, Number: q.Number, FrameIndex: int(q.Number) % size}
		} else {
			rp = theme.RenderParams{Count: 0, Number: -1, FrameIndex: frameOf(0, size)}
		}
	case q.Number > 0:
		// Preview mode: show the given number, no increment.
		rp = theme.RenderParams{Count: q.Number, Number: q.Number, FrameIndex: int(q.Number) % size}
	default:
		if s.counter == nil {
			return fiber.NewError(fiber.StatusServiceUnavailable, "counter not configured")
		}
		count, err := s.incrementOrDegrade(c, name)
		if err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, err.Error())
		}
		// M2.5: frame = (count+1) % size, text = count.
		rp = theme.RenderParams{Count: count, Number: -1, FrameIndex: frameOf(count+1, size)}
	}

	svg, err := theme.Render(th, rp)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}

	c.Set("Content-Type", "image/svg+xml")
	// Iron Rule 1: real counters are no-store; only demo is long cache.
	if name == "demo" {
		c.Set("Cache-Control", "public, max-age=31536000")
	} else {
		c.Set("Cache-Control", "no-store")
	}
	return c.Status(fiber.StatusOK).SendString(svg)
}

// incrementOrDegrade applies the name-level rate limit. Within quota it
// increments; over quota it returns the current count WITHOUT +1
// (AGENTS.md Iron Rule 3: degrade read-only, not 429).
func (s *Server) incrementOrDegrade(c fiber.Ctx, name string) (int64, error) {
	if s.nameLimiter != nil && !s.nameLimiter.Allow(name, time.Now()) {
		return s.counter.Get(c.Context(), name)
	}
	return s.counter.Incr(c.Context(), name)
}

// frameOf returns (v) % size, guarding against size==0.
func frameOf(v int64, size int) int {
	if size <= 0 {
		return 0
	}
	return int(v % int64(size))
}
