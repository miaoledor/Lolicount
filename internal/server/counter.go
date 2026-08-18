package server

import (
	"strings"
	"time"

	"github.com/gofiber/fiber/v3"

	"github.com/miaoledor/lolicount/internal/theme"
)

// counterHandler renders GET /@:name (and the /get/@:name alias).
//
// M5.5: theme IS the background (layer 0); the count is shown only by
// the overlaid <text> (layer 1). There is no separate bg concept — the
// theme frame is the sole background image.
// M5.6: the frame is scaled to a uniform display size (scale param);
// the text sits below the image and can be hidden via ?unshowf=true.
// The frame advances with the count: frameIndex = (count+1) % size, so
// the background image cycles as the counter grows (M2.5 frame model).
//
// M4 scope still applies: name-level rate limiting with read-only
// degradation (Iron Rule 3) and Cache-Control no-store for real counters
// (Iron Rule 1); demo is long cache.
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

	var rp theme.RenderParams
	switch {
	case name == "demo":
		// Reserved: never count, long cache (Iron Rule 1). Frame stays 0.
		if q.Number > 0 {
			rp = theme.RenderParams{Count: q.Number, Number: q.Number, FrameIndex: 0, FontSize: q.FSize, Scale: q.Scale, UnshowFont: q.UnshowF}
		} else {
			rp = theme.RenderParams{Count: 0, Number: -1, FrameIndex: 0, FontSize: q.FSize, Scale: q.Scale, UnshowFont: q.UnshowF}
		}
	case q.Number > 0:
		// Preview mode: show the given number, no increment, frame 0.
		rp = theme.RenderParams{Count: q.Number, Number: q.Number, FrameIndex: 0, FontSize: q.FSize, Scale: q.Scale, UnshowFont: q.UnshowF}
	default:
		if s.counter == nil {
			return fiber.NewError(fiber.StatusServiceUnavailable, "counter not configured")
		}
		count, err := s.incrementOrDegrade(c, name)
		if err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, err.Error())
		}
		// Frame advances with the count: (count+1) % size (M2.5).
		rp = theme.RenderParams{Count: count, Number: -1, FrameIndex: frameIndexForCount(count, th.Size()), FontSize: q.FSize, Scale: q.Scale, UnshowFont: q.UnshowF}
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

// frameIndexForCount picks the background frame for a given count.
// Per M2.5: display frame[(count+1) % size]. size<=1 guards against a
// single-frame theme (always 0) and division by zero.
func frameIndexForCount(count int64, size int) int {
	if size <= 1 {
		return 0
	}
	return int((count + 1) % int64(size))
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
