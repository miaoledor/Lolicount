package server

import (
	"time"

	"github.com/gofiber/fiber/v3"
)

// ipRateLimit is Fiber middleware that applies the IP-level token
// buckets. Over-limit requests get 429 (AGENTS.md Iron Rule 3: IP limit
// is the 429 path, as opposed to name-level degradation).
//
// c.IP() relies on the Fiber TrustedProxies config (set in server.New):
// behind a reverse proxy it reads X-Forwarded-For from trusted hop IPs
// only, so a spoofed header from an untrusted source can't bypass the
// limiter by pretending to be a different client.
func (s *Server) ipRateLimit(c fiber.Ctx) error {
	if s.ipLimiter == nil {
		return c.Next()
	}
	ip := c.IP()
	if !s.ipLimiter.Allow(ip, time.Now()) {
		c.Set("Retry-After", "1")
		return fiber.NewError(fiber.StatusTooManyRequests,
			"rate limit exceeded for IP "+ip)
	}
	return c.Next()
}

// cors applies permissive CORS only to the /api/* upload channel.
// Counter SVG paths (/@:name) are embedded in README/HTML and must NOT
// carry CORS headers (AGENTS.md Key Conventions).
func cors() fiber.Handler {
	allowed := map[string]struct{}{
		"GET":     {},
		"POST":    {},
		"OPTIONS": {},
	}
	return func(c fiber.Ctx) error {
		origin := c.Get("Origin")
		if origin != "" {
			c.Set("Access-Control-Allow-Origin", origin)
			c.Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
			c.Set("Access-Control-Allow-Headers", "Content-Type")
			c.Set("Access-Control-Max-Age", "86400")
		}
		if c.Method() == fiber.MethodOptions {
			return c.SendStatus(fiber.StatusNoContent)
		}
		if _, ok := allowed[c.Method()]; !ok {
			return fiber.NewError(fiber.StatusMethodNotAllowed, "method not allowed")
		}
		return c.Next()
	}
}
