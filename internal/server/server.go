// Package server wires the Fiber v3 application: routes, middleware and
// graceful shutdown. Handlers are added incrementally per milestone.
package server

import (
	"context"
	"errors"
	"io/fs"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/compress"
	"github.com/rs/zerolog"

	"github.com/miaoledor/lolicount/assets"
	"github.com/miaoledor/lolicount/internal/config"
	"github.com/miaoledor/lolicount/internal/counter"
	"github.com/miaoledor/lolicount/internal/imgcore/composer"
	"github.com/miaoledor/lolicount/internal/ratelimit"
)

// hotThemeCount is the size of the hot-themes list surfaced by
// GET /api/themes/hot (per spec: exactly 10).
const hotThemeCount = 10

// hotThemeRefresh is how often the hot list is recomputed from the
// usage buffer (per spec: every hour; it also refreshes at startup).
const hotThemeRefresh = time.Hour

// Server holds the Fiber app and its dependencies.
type Server struct {
	app         *fiber.App
	cfg         *config.Config
	logger      zerolog.Logger
	themes      composer.ThemeRegistry
	fthemes     composer.FThemeRegistry
	counter     *counter.Buffer
	themeUsage  *counter.Buffer
	ipLimiter   *ratelimit.IPLimiter
	nameLimiter *ratelimit.NameLimiter
	psbFS       fs.FS
	spineFS     fs.FS
	live2dFS    fs.FS
	unityFS     fs.FS

	// Hot-theme cache: recomputed at startup and every hotThemeRefresh
	// from the theme usage buffer, served by GET /api/themes/hot.
	hotMu     sync.RWMutex
	hotThemes []string
	hotStop   chan struct{}
}

// New constructs the Server with routes and middleware registered.
// themeUsage is the theme-popularity buffer (may be nil in tests: the
// hot list then stays empty and usage is not tracked).
func New(cfg *config.Config, logger zerolog.Logger, themes composer.ThemeRegistry, fthemes composer.FThemeRegistry, buf *counter.Buffer, themeUsage *counter.Buffer) *Server {
	app := fiber.New(fiber.Config{
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  30 * time.Second,
		BodyLimit:    50 * 1024 * 1024,
		AppName:      "lolicount",
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
		themeUsage:  themeUsage,
		ipLimiter:   ratelimit.NewIPLimiter(cfg.RateLimitIPPerSec, cfg.RateLimitIPPerMin),
		nameLimiter: ratelimit.NewNameLimiter(cfg.RateLimitNamePerSec),
		hotStop:     make(chan struct{}),
	}
	// Emote (PSB) models are served from the on-disk PSB_DIR (default
	// assets/psb) on demand — never embedded in the binary and never held
	// in process memory: each request streams the file straight from disk.
	// A missing directory leaves psbFS nil, which the handlers treat as
	// "no models" (empty list, 404 on fetch) instead of failing startup.
	if st, err := os.Stat(cfg.PSBDir); err == nil && st.IsDir() {
		s.psbFS = os.DirFS(cfg.PSBDir)
	} else {
		s.logger.Warn().Str("psb_dir", cfg.PSBDir).Msg("psb dir missing, emote widget endpoints disabled")
	}
	// Spine dynamic-illustration models live under assets/spine/. Same
	// convention: a missing tree leaves spineFS nil => "no models".
	if spineRoot, err := fs.Sub(assets.FS, "spine"); err == nil {
		s.spineFS = spineRoot
	}
	// Live2D (Cubism) dynamic-illustration models live under assets/live2d/.
	// Same convention: a missing tree leaves live2dFS nil => "no models".
	if live2dRoot, err := fs.Sub(assets.FS, "live2d"); err == nil {
		s.live2dFS = live2dRoot
	}
	// Unity UGUI dynamic-illustration models live under assets/unity/. Same
	// convention: a missing tree leaves unityFS nil => "no models".
	if unityRoot, err := fs.Sub(assets.FS, "unity"); err == nil {
		s.unityFS = unityRoot
	}
	// Hot themes: seed the cache at startup (spec: refresh at startup and
	// every hour) and run the hourly refresh loop.
	s.refreshHotThemes()
	go s.hotLoop()

	s.registerRoutes()
	return s
}

// refreshHotThemes recomputes the cached top-10 theme list from the
// usage buffer. Falls back to the first registered themes (alphabetical)
// when no usage has accumulated yet, so the category always exists.
func (s *Server) refreshHotThemes() {
	if s.themeUsage == nil {
		return
	}
	top := s.themeUsage.Top(hotThemeCount)
	names := make([]string, 0, len(top))
	for _, c := range top {
		names = append(names, c.Name)
	}
	if len(names) == 0 && s.themes != nil {
		for _, e := range s.themes.List() {
			names = append(names, e.Name)
			if len(names) >= hotThemeCount {
				break
			}
		}
	}
	s.hotMu.Lock()
	s.hotThemes = names
	s.hotMu.Unlock()
}

// hotLoop recomputes the hot list every hotThemeRefresh until stopped.
func (s *Server) hotLoop() {
	ticker := time.NewTicker(hotThemeRefresh)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			s.refreshHotThemes()
		case <-s.hotStop:
			return
		}
	}
}

// registerRoutes wires all HTTP routes.
func (s *Server) registerRoutes() {
	// Compress text-heavy responses: counter SVGs embed ~50KB-1.6MB of
	// base64 art and gzip fully recovers the base64 inflation. Level
	// BestSpeed bounds the CPU — a fasthttp limitation is that responses
	// are re-compressed on every hit with no compressed-body cache, so
	// heavier levels would tax every README view. Paths serving
	// already-compressed binaries (PSB models set Content-Encoding
	// themselves and are auto-skipped by the middleware; webp/png gallery
	// images and Live2D/Spine model binaries gain nothing from gzip) are
	// skipped here outright.
	s.app.Use(compress.New(compress.Config{
		Level: compress.LevelBestSpeed,
		Next: func(c fiber.Ctx) bool {
			p := c.Path()
			for _, prefix := range []string{
				"/images/", "/emote/", "/psb/", "/api/psb/", "/live2d/models/", "/spine/", "/unity/models/",
			} {
				if strings.HasPrefix(p, prefix) {
					return true
				}
			}
			return false
		},
	}))

	s.app.Get("/heart-beat", s.heartbeat)

	s.app.Get("/@:name", sanitizeBackslashEscape, s.ipRateLimit, s.counterHandler)
	s.app.Get("/get/@:name", sanitizeBackslashEscape, s.ipRateLimit, s.counterHandler)
	s.app.Get("/record/@:name", sanitizeBackslashEscape, s.ipRateLimit, s.recordHandler)

	s.app.Use("/api", cors())

	s.app.Get("/api/themes", s.listThemes)
	s.app.Get("/api/themes/hot", s.listHotThemes)
	s.app.Get("/api/fthemes", s.listFThemes)
	s.app.Get("/api/config", s.getConfig)
	s.app.Get("/api/count/@:name", sanitizeBackslashEscape, s.ipRateLimit, s.countHandler)
	s.app.Get("/api/psb/models", s.listPsbModels)
	s.app.Get("/api/psb/:model/download", s.psbModelDownload)
	s.app.Get("/api/spine/models", s.listSpineModels)
	s.app.Get("/api/live2d/models", s.listLive2DModels)
	s.app.Get("/api/unity/models", s.listUnityModels)
	s.app.Post("/api/editor/preview", s.editorPreviewHandler)
	s.app.Post("/api/editor/export", s.editorExportHandler)

	s.app.Get("/psb/:model", s.psbModelHandler)
	s.app.Get("/spine/models/:name/:file", s.spineModelHandler)
	s.app.Get("/spine/anim/:name/:file", s.spineAnimHandler)
	s.app.Get("/live2d/models/:name/:file", s.live2dModelHandler)
	s.app.Get("/unity/models/:name/:file", s.unityModelHandler)

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
	close(s.hotStop)
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
