package server

import (
	"github.com/gofiber/fiber/v3"

	"github.com/miaoledor/lolicount/internal/theme"
)

// listThemes answers GET /api/themes with the registered themes and
// their kind (frame/character). The front-end uses the kind to render
// type-specific playground controls (M9). Read-only and stable, so a
// short cache is fine (the list only changes on rebuild).
func (s *Server) listThemes(c fiber.Ctx) error {
	if s.themes == nil {
		return c.Status(fiber.StatusOK).JSON(fiber.Map{"themes": []fiber.Map{}})
	}
	c.Set("Cache-Control", "public, max-age=60")
	infos := s.themes.ListWithKind()
	exposed := make([]fiber.Map, 0, len(infos))
	for _, ti := range infos {
		kind := "frame"
		if ti.Kind == theme.KindCharacter {
			kind = "character"
		}
		exposed = append(exposed, fiber.Map{"name": ti.Name, "kind": kind})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{"themes": exposed})
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
