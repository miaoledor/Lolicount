// Package server count_api.go handles GET /api/count/@:name — the JSON
// counterpart of counterHandler used by the emote widget (docs/emote-widget.md).
package server

import (
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v3"
)

// countHandler answers GET /api/count/@:name with the count as JSON,
// mirroring counterHandler semantics: real names increment (name-level
// rate limiting degrades read-only, Iron Rule 3), while demo / number>0
// return fixed values without incrementing. The route is mounted under
// /api so it inherits the CORS middleware — the widget fetches it
// cross-origin from third-party pages. Real counts are always no-store
// (Iron Rule 1).
func (s *Server) countHandler(c fiber.Ctx) error {
	// See counterHandler: clone the param so it never aliases a reused
	// fasthttp buffer when it reaches the buffer/store layer.
	name := strings.Clone(c.Params("name"))
	if name == "" {
		return fiber.NewError(fiber.StatusBadRequest, "missing name")
	}

	// Same fixed-number contract as the SVG path: values within
	// [0, 999999] are displayed as-is without counting.
	number := int64(0)
	if raw := c.Query("number"); raw != "" {
		parsed, err := strconv.ParseInt(raw, 10, 64)
		if err != nil || parsed < 0 || parsed > 999999 {
			return fiber.NewError(fiber.StatusBadRequest, "invalid number")
		}
		number = parsed
	}

	var count int64
	switch {
	case name == "demo":
		if number > 0 {
			count = number
		} else {
			// The SVG demo renders "0123456789"; as a JSON number the
			// leading zero is dropped, so the digit sequence is used.
			count = 123456789
		}
	case number > 0:
		count = number
	default:
		if s.counter == nil {
			return fiber.NewError(fiber.StatusServiceUnavailable, "counter not configured")
		}
		var err error
		count, err = s.incrementOrDegrade(c, name)
		if err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, err.Error())
		}
	}

	c.Set("Cache-Control", "no-store")
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"name": name,
		"num":  count,
	})
}
