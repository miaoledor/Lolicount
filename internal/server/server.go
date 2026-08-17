// Package server wires the Fiber v3 application: routes, middleware and
// graceful shutdown. Handlers are added incrementally per milestone.
package server

import (
	"context"
	"errors"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/rs/zerolog"

	"github.com/miaoledor/lolicount/internal/bg"
	"github.com/miaoledor/lolicount/internal/config"
	"github.com/miaoledor/lolicount/internal/counter"
	"github.com/miaoledor/lolicount/internal/ratelimit"
	"github.com/miaoledor/lolicount/internal/theme"
)

// Server holds the Fiber app and its dependencies.
type Server struct {
	app         *fiber.App
	cfg         *config.Config
	logger      zerolog.Logger
	themes      theme.Registry
	counter     *counter.Buffer
	backgrounds bg.Registry
	ipLimiter   *ratelimit.IPLimiter
	nameLimiter *ratelimit.NameLimiter
}

// New constructs the Server with routes and middleware registered.
// themes may be nil in M1-only setups; counter may be nil before M3.
func New(cfg *config.Config, logger zerolog.Logger, themes theme.Registry, buf *counter.Buffer, bgs bg.Registry) *Server {
	app := fiber.New(fiber.Config{
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  30 * time.Second,
		AppName:      "lolicount",
		// TrustProxy makes c.IP() read X-Forwarded-For from trusted hop
		// IPs (loopback/private) so rate limiting works correctly behind
		// a reverse proxy. Without it, every proxied request looks like
		// it comes from 127.0.0.1 and shares one IP quota.
		TrustProxy: cfg.TrustProxy,
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
		counter:     buf,
		backgrounds: bgs,
		ipLimiter:   ratelimit.NewIPLimiter(cfg.RateLimitIPPerSec, cfg.RateLimitIPPerMin),
		nameLimiter: ratelimit.NewNameLimiter(cfg.RateLimitNamePerSec),
	}
	s.registerRoutes()
	return s
}

// registerRoutes wires all HTTP routes. Extended in later milestones.
func (s *Server) registerRoutes() {
	s.app.Get("/heart-beat", s.heartbeat)

	// Counter SVG paths: IP rate limit applies (429 on over-limit).
	// /get/@:name is a compatibility alias (Moe-Counter). The limiter is
	// mounted per-route (not on "/") so 404/405 paths are unaffected.
	s.app.Get("/@:name", s.ipRateLimit, s.counterHandler)
	s.app.Get("/get/@:name", s.ipRateLimit, s.counterHandler)
	s.app.Get("/record/@:name", s.ipRateLimit, s.recordHandler)

	// Upload channel (M6): CORS only here, NOT on counter SVG paths
	// (AGENTS.md Key Conventions).
	s.app.Use("/api", cors())
}

// Listen starts the HTTP server on the configured address.
func (s *Server) Listen() error {
	s.logger.Info().Str("addr", s.cfg.Addr()).Msg("server starting")
	return s.app.Listen(s.cfg.Addr())
}

// Shutdown gracefully stops the server, draining in-flight requests and
// halting the rate-limiter reaper goroutines.
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
