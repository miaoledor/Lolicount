package server

import (
	"github.com/gofiber/fiber/v3"

	"github.com/miaoledor/lolicount/internal/imgcore"
)

// listThemes answers GET /api/themes with the registered themes and
// their kind (frame/character). Read-only and stable, so a short cache
// is fine.
func (s *Server) listThemes(c fiber.Ctx) error {
	if s.themes == nil {
		return c.Status(fiber.StatusOK).JSON(fiber.Map{"themes": []fiber.Map{}})
	}
	c.Set("Cache-Control", "public, max-age=60")
	entries := s.themes.List()
	exposed := make([]fiber.Map, 0, len(entries))
	for _, e := range entries {
		kind := "frame"
		if e.Kind == imgcore.LegacyKindCharacter {
			kind = "character"
		}
		exposed = append(exposed, fiber.Map{"name": e.Name, "kind": kind})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{"themes": exposed})
}

// listFThemes answers GET /api/fthemes with the registered font-style
// theme names.
func (s *Server) listFThemes(c fiber.Ctx) error {
	if s.fthemes == nil {
		return c.Status(fiber.StatusOK).JSON(fiber.Map{"fthemes": []string{}})
	}
	c.Set("Cache-Control", "public, max-age=60")
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"fthemes": s.fthemes.List(),
	})
}

// getConfig answers GET /api/config with the public-facing configuration
// the front-end needs to build embed links.
func (s *Server) getConfig(c fiber.Ctx) error {
	c.Set("Cache-Control", "public, max-age=60")
	baseURL := ""
	if s.cfg != nil {
		baseURL = s.cfg.BaseURL
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"baseUrl": baseURL,
	})
}
