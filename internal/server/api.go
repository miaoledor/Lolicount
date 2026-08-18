package server

import "github.com/gofiber/fiber/v3"

// listThemes answers GET /api/themes with the registered theme names.
// Front-end theme gallery consumes this. Read-only and stable, so a
// short cache is fine (the list only changes on rebuild).
func (s *Server) listThemes(c fiber.Ctx) error {
	if s.themes == nil {
		return c.Status(fiber.StatusOK).JSON(fiber.Map{"themes": []string{}})
	}
	c.Set("Cache-Control", "public, max-age=60")
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"themes": s.themes.List(),
	})
}

// listFThemes answers GET /api/fthemes with the registered font-style
// theme names. Front-end style picker consumes this.
func (s *Server) listFThemes(c fiber.Ctx) error {
	if s.fthemes == nil {
		return c.Status(fiber.StatusOK).JSON(fiber.Map{"fthemes": []string{}})
	}
	c.Set("Cache-Control", "public, max-age=60")
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"fthemes": s.fthemes.List(),
	})
}
