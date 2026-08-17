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
	"github.com/miaoledor/lolicount/internal/theme"
)

// Server holds the Fiber app and its dependencies.
type Server struct {
	app    *fiber.App
	cfg    *config.Config
	logger zerolog.Logger
	themes theme.Registry
}

// New constructs the Server with routes and middleware registered.
// themes may be nil in M1-only setups; the counter route returns 503
// when no registry is configured.
func New(cfg *config.Config, logger zerolog.Logger, themes theme.Registry) *Server {
	app := fiber.New(fiber.Config{
		ReadTimeout:           10 * time.Second,
		WriteTimeout:          10 * time.Second,
		IdleTimeout:           30 * time.Second,
		AppName:               "lolicount",
	})

	s := &Server{app: app, cfg: cfg, logger: logger, themes: themes}
	s.registerRoutes()
	return s
}

// registerRoutes wires all HTTP routes. Extended in later milestones.
func (s *Server) registerRoutes() {
	s.app.Get("/heart-beat", s.heartbeat)
	// Counter SVG. /get/@:name is a compatibility alias (Moe-Counter).
	s.app.Get("/@:name", s.counterHandler)
	s.app.Get("/get/@:name", s.counterHandler)
}

// Listen starts the HTTP server on the configured address.
func (s *Server) Listen() error {
	s.logger.Info().Str("addr", s.cfg.Addr()).Msg("server starting")
	return s.app.Listen(s.cfg.Addr())
}

// Shutdown gracefully stops the server, draining in-flight requests.
func (s *Server) Shutdown(ctx context.Context) error {
	s.logger.Info().Msg("server shutting down")
	if err := s.app.ShutdownWithContext(ctx); err != nil && !errors.Is(err, context.DeadlineExceeded) {
		return err
	}
	return nil
}
