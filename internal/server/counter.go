package server

import (
	"github.com/gofiber/fiber/v3"

	"github.com/miaoledor/lolicount/internal/theme"
)

// demoNumber is the value rendered for the reserved name "demo": the
// digits 0..9 in order. Stored as 1234567890 and forced to padding>=10
// so the leading zero is preserved (AGENTS.md: demo固定返回0123456789).
const demoNumber int64 = 1234567890

// counterHandler renders GET /@:name (and the /get/@:name alias).
//
// M2 scope: only the render path is implemented. "demo" returns a fixed
// number; num>0 returns that number directly; otherwise the count is a
// placeholder 0 until the counter.Buffer lands in M3. Cache-Control is
// no-store for now; demo's long-cache variant arrives in M4 (Iron Rule 1).
func (s *Server) counterHandler(c fiber.Ctx) error {
	name := c.Params("name")
	if name == "" {
		return fiber.NewError(fiber.StatusBadRequest, "missing name")
	}

	q, err := parseParams(c)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}

	var count int64
	switch {
	case name == "demo":
		count = demoNumber
		// Force a 10-digit field so the leading 0 shows: 0123456789.
		if q.Padding == nil {
			v := 10
			q.Padding = &v
		} else if *q.Padding < 10 {
			v := 10
			q.Padding = &v
		}
	case q.Num > 0:
		count = q.Num
	default:
		// Placeholder until M3 wires counter.Buffer.Incr(name).
		count = 0
	}

	th, err := resolveTheme(s.themes, q.Theme)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}

	svg, err := theme.Render(th, q.toRenderParams(count))
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}

	c.Set("Content-Type", "image/svg+xml")
	c.Set("Cache-Control", "no-store")
	return c.Status(fiber.StatusOK).SendString(svg)
}
