// Package server counter.go handles GET /@:name (and /get/@:name).
//
// The theme is the background (layer 0); the count is shown by the
// overlaid <text> layer. All themes go through the same unified compose
// path — no card/character branching. Name-level rate limiting degrades
// to read-only (Iron Rule 3); Cache-Control no-store for real counters
// (Iron Rule 1); demo is long cache.
package server

import (
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v3"

	"github.com/miaoledor/lolicount/internal/imgcore/composer"
	"github.com/miaoledor/lolicount/internal/imgcore/theme"
)

// counterHandler renders GET /@:name (and the /get/@:name alias).
func (s *Server) counterHandler(c fiber.Ctx) error {
	// Fiber/fasthttp route params can reference a per-request buffer that
	// the runtime reuses across requests. The name is later stored as a
	// map key inside counter.Buffer, so it MUST outlive the request.
	name := strings.Clone(c.Params("name"))
	if name == "" {
		return fiber.NewError(fiber.StatusBadRequest, "missing name")
	}

	q, err := parseParams(c)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}

	entry, err := composer.ResolveTheme(s.themes, q.Theme)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}
	fst, err := composer.ResolveFTheme(s.fthemes, q.FTheme)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}
	style := theme.TextStyle{Family: fst.Family, Color: fst.Color, Weight: fst.Weight}

	// Resolve the final text string.
	var text string
	switch {
	case name == "demo":
		if q.Number > 0 {
			text = strconv.FormatInt(q.Number, 10)
		} else {
			text = "0123456789"
		}
	case q.Number > 0:
		text = strconv.FormatInt(q.Number, 10)
	default:
		if s.counter == nil {
			return fiber.NewError(fiber.StatusServiceUnavailable, "counter not configured")
		}
		count, err := s.incrementOrDegrade(c, name)
		if err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, err.Error())
		}
		text = strconv.FormatInt(count, 10)
	}

	// Unified compose path for all theme types.
	svg, err := s.compose(entry, q, text, style)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}

	c.Set("Content-Type", "image/svg+xml")
	// Iron Rule 1: demo with a fixed value (number>0) or a single-frame
	// theme is long-cached because the output is deterministic. Demo with
	// a multi-frame theme uses random selection, so it must be no-store.
	// Real counters are always no-store.
	if name == "demo" && (q.Number > 0 || !s.themeIsMultiFrame(entry.Name)) {
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
