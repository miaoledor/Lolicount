package server

import (
	"io/fs"
	"strings"

	"github.com/gofiber/fiber/v3"

	"github.com/miaoledor/lolicount/assets"
)

// registerFrontend serves the embedded Nuxt SSG dist as static files and
// adds a SPA fallback so unknown paths (handled client-side by Vue Router)
// return index.html instead of 404. It must be registered AFTER all API
// and counter routes so they take precedence.
func (s *Server) registerFrontend() {
	dist, err := fs.Sub(assets.DistFS, "dist")
	if err != nil {
		s.logger.Error().Err(err).Msg("frontend dist unavailable")
		return
	}

	// Serve static assets (_nuxt/*, favicon.ico, etc.) directly.
	s.app.Use("*", func(c fiber.Ctx) error {
		// Skip API and counter paths — they are handled by their own routes.
		p := c.Path()
		if isDynamicPath(p) {
			return c.Next()
		}

		// Try the exact file first.
		if f, err := dist.Open(strings.TrimPrefix(p, "/")); err == nil {
			f.Close()
			return c.SendFile(p, fiber.SendFile{
				FS: dist,
			})
		}

		// SPA fallback: unknown non-asset paths serve index.html so Vue
		// Router can handle client-side routing.
		return c.SendFile("index.html", fiber.SendFile{
			FS: dist,
		})
	})
}

// isDynamicPath returns true for paths handled by API/counter routes that
// must NOT be intercepted by the frontend static handler.
func isDynamicPath(p string) bool {
	return strings.HasPrefix(p, "/api/") ||
		strings.HasPrefix(p, "/@") ||
		strings.HasPrefix(p, "/get/@") ||
		strings.HasPrefix(p, "/record/@") ||
		p == "/heart-beat"
}
