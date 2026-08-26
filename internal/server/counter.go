package server

import (
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v3"

	"github.com/miaoledor/lolicount/internal/imgcore"
	"github.com/miaoledor/lolicount/internal/imgcore/fdrawer"
	"github.com/miaoledor/lolicount/internal/imgcore/renderer"
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

	entry, err := renderer.ResolveTheme(s.themes, q.Theme)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}
	fst, err := renderer.ResolveFTheme(s.fthemes, q.FTheme)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}
	fs := fdrawer.FontStyle{Family: fst.Family, Color: fst.Color, Weight: fst.Weight}
	pos := fdrawer.TextPos{X: q.X, Y: q.Y, RX: q.RX, RY: q.RY}
	mode := renderer.ModeForTheme(entry.Kind, q.Mode)

	// Resolve the final text string and background params.
	var text string
	var frameIndex int
	switch {
	case name == "demo":
		// Reserved: never count, long cache (Iron Rule 1). Shows
		// 0123456789 unless ?number= overrides the preview value.
		if q.Number > 0 {
			text = strconv.FormatInt(q.Number, 10)
		} else {
			text = "0123456789"
		}
		frameIndex = 0
	case q.Number > 0:
		// Preview mode: show the given number, no increment, frame 0.
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
		// Frame size depends on theme kind: card themes have frames;
		// character themes are always random (size irrelevant here).
		size := 0
		if th, ok := s.themes.GetCard(entry.Name); ok {
			size = th.Size()
		}
		frameIndex = renderer.FrameIndexForCount(count, size)
	}

	// Build the render params based on theme kind.
	rp := renderer.RenderParams{
		ThemeKind:  entry.Kind,
		Scale:      q.Scale,
		Text:       text,
		FontSize:   q.FSize,
		UnshowFont: q.UnshowF,
		FontStyle:  fs,
		Position:   pos,
	}
	if entry.Kind == imgcore.LegacyKindCharacter {
		ch, ok := s.themes.GetCharacter(entry.Name)
		if !ok {
			return fiber.NewError(fiber.StatusBadRequest, "character theme not found")
		}
		portrait, err := ch.Assemble(nil)
		if err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, err.Error())
		}
		rp.Portrait = portrait
	} else {
		th, ok := s.themes.GetCard(entry.Name)
		if !ok {
			return fiber.NewError(fiber.StatusBadRequest, "card theme not found")
		}
		frame, ok := renderer.PickFrame(th, mode, frameIndex, nil)
		if !ok {
			return fiber.NewError(fiber.StatusInternalServerError, "frame index out of range")
		}
		rp.Frame = frame
	}

	svg, err := renderer.Render(rp)
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
