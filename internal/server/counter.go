// Package server counter.go handles GET /@:name (and /get/@:name).
//
// The theme is the background (layer 0); the count is shown by the
// overlaid <text> layer. All themes go through the same unified compose
// path — no card/character branching. Name-level rate limiting degrades
// to read-only (Iron Rule 3); real counters use realCountCacheControl
// (Iron Rule 1); only deterministic demo output is long-cached.
package server

import (
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v3"

	"github.com/miaoledor/lolicount/internal/imgcore/composer"
	"github.com/miaoledor/lolicount/internal/imgcore/theme"
)

// DemoText is the fixed number rendered for the reserved name=demo.
// It is part of the public contract (Iron Rule 1: demo is fixed
// 0123456789 and never stored) and is also the text used by
// cmd/gen-theme-thumbs when pre-rendering gallery thumbnails.
const DemoText = "0123456789"

// realCountCacheControl is the Cache-Control value for every response
// carrying a real, per-view count (Iron Rule 1). Bare no-store is NOT
// enough: GitHub proxies README images through camo + Fastly, and the
// Fastly edge caches the object anyway (observed x-cache MISS → HIT
// within one second) while passing no-store through to the browser —
// so the count freezes until the edge TTL expires. The edge only
// revalidates per request when the response carries no-cache /
// max-age=0; must-revalidate and no-store keep browsers from serving
// or persisting stale bodies. This combo mirrors komarev/ghpvc, which
// is field-proven on the exact same GitHub pipeline.
const realCountCacheControl = "no-store, no-cache, max-age=0, must-revalidate"

// counterHandler renders GET /@:name (and the /get/@:name alias).
func (s *Server) counterHandler(c fiber.Ctx) error {
	// Fiber/fasthttp route params can reference a per-request buffer that
	// the runtime reuses across requests. The name is later stored as a
	// map key inside counter.Buffer, so it MUST outlive the request.
	name := strings.Clone(c.Params("name"))
	if name == "" {
		return fiber.NewError(fiber.StatusBadRequest, "missing name")
	}

	q, err := parseParams(c)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}

	entry, err := composer.ResolveTheme(s.themes, q.Theme)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}
	fst, err := composer.ResolveFTheme(s.fthemes, q.FTheme)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}
	style := theme.TextStyle{Family: fst.Family, Color: fst.Color, Weight: fst.Weight}

	// Resolve the final text string.
	var text string
	switch {
	case name == "demo":
		if q.Number > 0 {
			text = strconv.FormatInt(q.Number, 10)
		} else {
			text = DemoText
		}
	case q.Number > 0:
		text = strconv.FormatInt(q.Number, 10)
	default:
		if s.counter == nil {
			return fiber.NewError(fiber.StatusServiceUnavailable, "counter not configured")
		}
		count, err := s.incrementOrDegrade(c, name)
		if err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, err.Error())
		}
		text = strconv.FormatInt(count, 10)
		// Theme usage tracking feeds the hot-themes list (startup + hourly
		// refresh): every real render counts, degraded name-limit views
		// included — they are real visitors. Demo/number renders are not
		// usage. Failure to track is non-fatal.
		if s.themeUsage != nil {
			if _, err := s.themeUsage.Incr(c.Context(), entry.Name); err != nil {
				s.logger.Warn().Err(err).Str("theme", entry.Name).Msg("theme usage incr failed")
			}
		}
	}

	// Apply the custom text template when provided. The literal "{n}"
	// in q.Text is replaced with the resolved count string; any other
	// characters are preserved verbatim. When a text template is set,
	// unshowf is forced false so the text layer is always rendered.
	if q.Text != "" {
		text = strings.ReplaceAll(q.Text, "{n}", text)
		q.UnshowF = false
	}

	// Unified compose path for all theme types.
	svg, err := s.compose(entry, q, text, style)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}

	c.Set("Content-Type", "image/svg+xml")
	// CORS open: counters are public images, and the frontend's
	// download-current-frame feature rasterizes the cross-origin preview
	// (Nuxt dev serves the site on :3721 while the API runs on :9721) onto
	// a canvas — without this header the canvas is tainted and export
	// throws a SecurityError.
	c.Set("Access-Control-Allow-Origin", "*")
	// Iron Rule 1: only a single-frame demo is deterministic enough for
	// long caching. number>0 fixes the text, but a multi-frame background
	// remains random and must still use the no-cache combination.
	// Real counters are always no-cache.
	if name == "demo" && !s.themeIsMultiFrame(entry.Name) {
		c.Set("Cache-Control", "public, max-age=31536000")
	} else {
		c.Set("Cache-Control", realCountCacheControl)
	}
	return c.Status(fiber.StatusOK).SendString(svg)
}

// incrementOrDegrade applies the name-level rate limit. Within quota it
// increments; over quota it returns the current count WITHOUT +1
// (AGENTS.md Iron Rule 3: degrade read-only, not 429).
func (s *Server) incrementOrDegrade(c fiber.Ctx, name string) (int64, error) {
	if s.nameLimiter != nil && !s.nameLimiter.Allow(name, time.Now()) {
		return s.counter.Get(c.Context(), name)
	}
	return s.counter.Incr(c.Context(), name)
}
