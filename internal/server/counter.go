package server

import (
	"github.com/gofiber/fiber/v3"

	"github.com/miaoledor/lolicount/internal/theme"
)

// counterHandler renders GET /@:name (and the /get/@:name alias).
//
// M3 scope: real counting. "demo" is a reserved name that never counts
// (renders frame 0, text "0", no DB write). For other names, the counter
// buffer is incremented (+1) and the new value is rendered per M2.5:
// frame = (count+1) % size, text = count. The `number` query param
// overrides into preview mode (show that number, no increment, frame =
// number % size). Cache-Control is no-store for real counters (Iron
// Rule 1); demo long-cache arrives in M4.
func (s *Server) counterHandler(c fiber.Ctx) error {
	name := c.Params("name")
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
		// Reserved: never count. number>0 previews that frame; else 0.
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
		count, err := s.counter.Incr(c.Context(), name)
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
	c.Set("Cache-Control", "no-store")
	return c.Status(fiber.StatusOK).SendString(svg)
}

// frameOf returns (v) % size, guarding against size==0.
func frameOf(v int64, size int) int {
	if size <= 0 {
		return 0
	}
	return int(v % int64(size))
}
