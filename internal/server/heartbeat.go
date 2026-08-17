package server

import (
	"time"

	"github.com/gofiber/fiber/v3"
)

// heartbeat answers GET /heart-beat with an alive signal.
// Health checks must never be cached by proxies, so Cache-Control: no-store.
func (s *Server) heartbeat(c fiber.Ctx) error {
	c.Set("Cache-Control", "no-store")
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"status":    "alive",
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	})
}
