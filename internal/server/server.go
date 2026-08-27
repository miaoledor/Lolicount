// Package server wires the Fiber v3 application: routes, middleware and
// graceful shutdown. Handlers are added incrementally per milestone.
package server

import (
	"context"
	"errors"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/rs/zerolog"

	"github.com/miaoledor/lolicount/internal/config"
	"github.com/miaoledor/lolicount/internal/counter"
	"github.com/miaoledor/lolicount/internal/imgcore/composer"
	"github.com/miaoledor/lolicount/internal/ratelimit"
)

// Server holds the Fiber app and its dependencies.
type Server struct {
	app         *fiber.App
	cfg         *config.Config
	logger      zerolog.Logger
	themes      composer.ThemeRegistry
	fthemes     composer.FThemeRegistry
	counter     *counter.Buffer
	ipLimiter   *ratelimit.IPLimiter
	nameLimiter *ratelimit.NameLimiter
}

// New constructs the Server with routes and middleware registered.
func New(cfg *config.Config, logger zerolog.Logger, themes composer.ThemeRegistry, fthemes composer.FThemeRegistry, buf *counter.Buffer) *Server {
	app := fiber.New(fiber.Config{
		ReadTimeout:           30 * time.Second,
		WriteTimeout:          30 * time.Second,
		IdleTimeout:           30 * time.Second,
		BodyLimit:             50 * 1024 * 1024,
		AppName:               "lolicount",
		TrustProxy:   cfg.TrustProxy,
		TrustProxyConfig: fiber.TrustProxyConfig{
			Loopback: true,
			Private:  cfg.TrustProxyPrivate,
		},
		ProxyHeader: "X-Forwarded-For",
	})

	s := &Server{
		app:         app,
		cfg:         cfg,
		logger:      logger,
		themes:      themes,
		fthemes:     fthemes,
		counter:     buf,
		ipLimiter:   ratelimit.NewIPLimiter(cfg.RateLimitIPPerSec, cfg.RateLimitIPPerMin),
		nameLimiter: ratelimit.NewNameLimiter(cfg.RateLimitNamePerSec),
	}
	s.registerRoutes()
	return s
}

// registerRoutes wires all HTTP routes.
func (s *Server) registerRoutes() {
	s.app.Get("/heart-beat", s.heartbeat)

	s.app.Get("/@:name", sanitizeBackslashEscape, s.ipRateLimit, s.counterHandler)
	s.app.Get("/get/@:name", sanitizeBackslashEscape, s.ipRateLimit, s.counterHandler)
	s.app.Get("/record/@:name", sanitizeBackslashEscape, s.ipRateLimit, s.recordHandler)

	s.app.Use("/api", cors())

	s.app.Get("/api/themes", s.listThemes)
	s.app.Get("/api/fthemes", s.listFThemes)
	s.app.Get("/api/config", s.getConfig)
	s.app.Post("/api/editor/preview", s.editorPreviewHandler)
	s.app.Post("/api/editor/export", s.editorExportHandler)

	// Admin routes — all require X-Admin-Key header (ADMIN_KEY env).
	// When ADMIN_KEY is empty, adminAuth returns 404 so the endpoints
	// are invisible, not just forbidden.
	admin := s.app.Group("/api/admin", s.adminAuth)
	admin.Get("/ping", func(c fiber.Ctx) error {
		return c.SendString("ok")
	})

	s.registerFrontend()
}

// Listen starts the HTTP server on the configured address.
func (s *Server) Listen() error {
	s.logger.Info().Str("addr", s.cfg.Addr()).Msg("server starting")
	return s.app.Listen(s.cfg.Addr())
}

// Shutdown gracefully stops the server.
func (s *Server) Shutdown(ctx context.Context) error {
	s.logger.Info().Msg("server shutting down")
	if s.ipLimiter != nil {
		s.ipLimiter.Stop()
	}
	if s.nameLimiter != nil {
		s.nameLimiter.Stop()
	}
	if err := s.app.ShutdownWithContext(ctx); err != nil && !errors.Is(err, context.DeadlineExceeded) {
		return err
	}
	return nil
}
