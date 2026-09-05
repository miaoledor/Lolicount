package server

import (
	"io/fs"
	"path"
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

	// Cache the raw embedded index.html; it never changes at runtime. The
	// baseUrl payload value is rewritten per-request when BASE_URL is set,
	// so a runtime env change takes effect without a rebuild (build once,
	// configure per env — same model kun-galgame-forum uses for SSR).
	rawIndex, err := fs.ReadFile(dist, "index.html")
	if err != nil {
		s.logger.Warn().Err(err).Msg("index.html not found in dist")
		rawIndex = nil
	}

	s.app.Use("*", func(c fiber.Ctx) error {
		p := c.Path()
		if isDynamicPath(p) {
			return c.Next()
		}

		// Try the exact file first. Prerendered Nuxt pages are directories
		// (e.g. emote/index.html): resolve a directory to its index.html
		// instead of letting the request fall through to the SPA fallback,
		// which serves the home page and navigates the client back to "/".
		name := strings.TrimSuffix(strings.TrimPrefix(p, "/"), "/")
		if name == "" {
			name = "index.html"
		} else if info, err := fs.Stat(dist, name); err == nil && info.IsDir() {
			name = path.Join(name, "index.html")
		}
		if _, err := fs.Stat(dist, name); err == nil {
			setFrontendCache(c, name)
			return c.SendFile("/"+name, fiber.SendFile{
				FS: dist,
			})
		}

		// SPA fallback: serve index.html. When BASE_URL is configured, the
		// baked baseUrl payload is rewritten so embed links use the runtime
		// domain instead of the build-time value.
		if rawIndex != nil {
			body := rawIndex
			if s.cfg != nil && s.cfg.BaseURL != "" {
				if rewritten := rewriteBaseUrl(body, s.cfg.BaseURL); rewritten != nil {
					body = rewritten
				}
			}
			setFrontendCache(c, "index.html")
			return c.Type("html").Send(body)
		}
		return c.SendFile("index.html", fiber.SendFile{
			FS: dist,
		})
	})
}

// setFrontendCache applies the cache policy for embedded frontend files.
// Nuxt build assets under _nuxt/ are content-hashed, so they are safely
// immutable. Everything else — HTML entry points, widget scripts, public
// images — must be no-store: embedded files have a zero modification
// time, so revalidation (no-cache + conditional GET) would always return
// 304 and keep serving stale bodies after a redeploy.
func setFrontendCache(c fiber.Ctx, name string) {
	if strings.HasPrefix(name, "_nuxt/") {
		c.Set("Cache-Control", "public, max-age=31536000, immutable")
	} else {
		c.Set("Cache-Control", "no-store")
	}
}

// rewriteBaseUrl replaces the baseUrl:"..." value in the __NUXT__ payload
// so a single image can be re-pointed at any domain without a rebuild.
// Returns nil if the marker is absent (caller serves the original bytes).
func rewriteBaseUrl(html []byte, baseURL string) []byte {
	const marker = `baseUrl:"`
	h := string(html)
	idx := strings.Index(h, marker)
	if idx < 0 {
		return nil
	}
	start := idx + len(marker)
	end := strings.Index(h[start:], `"`)
	if end < 0 {
		return nil
	}
	var b strings.Builder
	b.WriteString(h[:start])
	b.WriteString(baseURL)
	b.WriteString(h[start+end:])
	return []byte(b.String())
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
