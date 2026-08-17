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
)

// Server holds the Fiber app and its dependencies.
type Server struct {
	app    *fiber.App
	cfg    *config.Config
	logger zerolog.Logger
}

// New constructs the Server with routes and middleware registered.
func New(cfg *config.Config, logger zerolog.Logger) *Server {
	app := fiber.New(fiber.Config{
		ReadTimeout:           10 * time.Second,
		WriteTimeout:          10 * time.Second,
		IdleTimeout:           30 * time.Second,
		AppName:               "lolicount",
	})

	s := &Server{app: app, cfg: cfg, logger: logger}
	s.registerRoutes()
	return s
}

// registerRoutes wires all HTTP routes. Extended in later milestones.
func (s *Server) registerRoutes() {
	s.app.Get("/heart-beat", s.heartbeat)
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
