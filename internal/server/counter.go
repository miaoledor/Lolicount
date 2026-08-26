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
//
// The theme is the background (layer 0); the count is shown only by the
// overlaid <text> (layer 1). The frame advances with the count:
// frameIndex = (count+1) % size. Name-level rate limiting degrades to
// read-only (Iron Rule 3); Cache-Control no-store for real counters
// (Iron Rule 1); demo is long cache.
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

	// Resolve the final text string and background params.
	var text string
	var frameIndex int
	switch {
	case name == "demo":
		if q.Number > 0 {
			text = strconv.FormatInt(q.Number, 10)
		} else {
			text = "0123456789"
		}
		frameIndex = 0
	case q.Number > 0:
		text = strconv.FormatInt(q.Number, 10)
		frameIndex = 0
	default:
		if s.counter == nil {
			return fiber.NewError(fiber.StatusServiceUnavailable, "counter not configured")
		}
		count, err := s.incrementOrDegrade(c, name)
		if err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, err.Error())
		}
		text = strconv.FormatInt(count, 10)
		size := 0
		if th, ok := s.themes.GetCard(entry.Name); ok {
			size = th.Size()
		}
		frameIndex = frameIndexForCount(count, size)
	}

	// Render via the new composer using the bridge adapter.
	var svg string
	if entry.Kind == "character" {
		svg, err = s.composeCharacter(entry, q, text, style)
	} else {
		svg, err = s.composeCard(entry, q, text, frameIndex, style)
	}
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}

	c.Set("Content-Type", "image/svg+xml")
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
