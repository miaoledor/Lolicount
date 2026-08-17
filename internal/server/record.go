package server

import (
	"github.com/gofiber/fiber/v3"
	"strings"
)

// recordHandler answers GET /record/@:name with the current count as JSON.
// It does NOT increment (read-only). The value comes from the buffer if
// cached, otherwise from the store. Cache-Control: no-store so callers
// always see the latest value (Iron Rule 1).
func (s *Server) recordHandler(c fiber.Ctx) error {
	// See counterHandler: clone the param so it never aliases a reused
	// fasthttp buffer when it reaches the buffer/store layer.
	name := strings.Clone(c.Params("name"))
	if name == "" {
		return fiber.NewError(fiber.StatusBadRequest, "missing name")
	}
	if s.counter == nil {
		return fiber.NewError(fiber.StatusServiceUnavailable, "counter not configured")
	}

	count, err := s.counter.Get(c.Context(), name)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}

	c.Set("Cache-Control", "no-store")
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"name": name,
		"num":  count,
	})
}
