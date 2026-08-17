package server

import (
	"github.com/gofiber/fiber/v3"

	"github.com/miaoledor/lolicount/internal/theme"
)

// counterHandler renders GET /@:name (and the /get/@:name alias).
//
// M2.5 scope: single frame image + count text overlay. "demo" is a
// reserved name that never counts (returns frame 0, text "0"). For other
// names the count is a placeholder 0 until counter.Buffer lands in M3;
// the frame is selected by the `number` query param (default 0).
// Cache-Control is no-store for now; demo's long-cache arrives in M4
// (Iron Rule 1).
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

	// Count placeholder: demo and non-counting names show 0 until M3.
	var count int64
	if name == "demo" {
		count = 0
	}

	rp := q.toRenderParams(count, th.Size())
	svg, err := theme.Render(th, rp)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}

	c.Set("Content-Type", "image/svg+xml")
	c.Set("Cache-Control", "no-store")
	return c.Status(fiber.StatusOK).SendString(svg)
}
