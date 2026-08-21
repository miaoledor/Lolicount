// Command server is the lolicount HTTP entry point. It loads config,
// initializes the logger, opens the store + counter buffer, builds the
// Fiber server and runs it with graceful shutdown on SIGINT/SIGTERM.
package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/miaoledor/lolicount/internal/config"
	"github.com/miaoledor/lolicount/internal/counter"
	
	"github.com/miaoledor/lolicount/internal/logger"
	"github.com/miaoledor/lolicount/internal/server"
	"github.com/miaoledor/lolicount/internal/store"
	"github.com/miaoledor/lolicount/internal/renderer"
)

func main() {
	cfg, err := config.Load("")
	if err != nil {
		os.Stderr.WriteString("failed to load config: " + err.Error() + "\n")
		os.Exit(1)
	}

	logger.Init(cfg.LogLevel)
	log := logger.L

	log.Info().
		Str("host", cfg.Host).
		Int("port", cfg.Port).
		Str("log_level", cfg.LogLevel).
		Str("db_path", cfg.DBPath).
		Int("db_interval", cfg.DBInterval).
		Msg("config loaded")

	// Open the SQLite store (sole storage path, AGENTS.md Iron Rule 5).
	repo, err := store.NewSQLite(context.Background(), cfg.DBPath)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to open store")
	}
	defer func() {
		if c, ok := repo.(interface{ Close() error }); ok {
			c.Close()
		}
	}()

	// In-memory buffer fronting the store: increments hit memory, a
	// ticker flushes to SQLite every DB_INTERVAL seconds.
	buf := counter.New(repo, log, cfg.DBInterval)
	if err := buf.Start(context.Background()); err != nil {
		log.Fatal().Err(err).Msg("failed to start counter buffer")
	}
	defer buf.Stop()

	// Load built-in themes (card + character) from the embedded assets.
	themes, loadErrs := renderer.NewThemeRegistry()
	for _, e := range loadErrs {
		log.Warn().Err(e).Msg("theme load skipped")
	}
	if entries := themes.List(); len(entries) > 0 {
		names := make([]string, len(entries))
		for i, e := range entries {
			names[i] = e.Name
		}
		log.Info().Strs("themes", names).Msg("themes loaded")
	} else {
		log.Warn().Msg("no built-in themes loaded; /@:name will return 400 until a theme is added")
	}

	// Load built-in font-style themes from the embedded assets/f-theme
	// tree (?ftheme= selects a counter text style).
	fthemes, ftErrs := renderer.NewFThemeRegistry()
	for _, e := range ftErrs {
		log.Warn().Err(e).Msg("f-theme load skipped")
	}
	if names := fthemes.List(); len(names) > 0 {
		log.Info().Strs("f-themes", names).Msg("f-themes loaded")
	} else {
		log.Info().Msg("no built-in f-themes loaded; ?ftheme= unavailable until one is added")
	}

	srv := server.New(cfg, log, themes, fthemes, buf)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	errCh := make(chan error, 1)
	go func() {
		if err := srv.Listen(); err != nil {
			errCh <- err
		}
	}()

	select {
	case err := <-errCh:
		log.Fatal().Err(err).Msg("server failed to start")
	case sig := <-quit:
		log.Info().Str("signal", sig.String()).Msg("shutdown signal received")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Error().Err(err).Msg("graceful shutdown failed")
		os.Exit(1)
	}
	log.Info().Msg("server stopped")
}
