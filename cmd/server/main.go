// Command server is the lolicount HTTP entry point. It loads config,
// initializes the logger, builds the Fiber server and runs it with
// graceful shutdown on SIGINT/SIGTERM.
package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/miaoledor/lolicount/internal/config"
	"github.com/miaoledor/lolicount/internal/logger"
	"github.com/miaoledor/lolicount/internal/server"
)

func main() {
	cfg, err := config.Load("")
	if err != nil {
		// Logger not ready yet; print to stderr and exit.
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

	srv := server.New(cfg, log)

	// Graceful shutdown: wait for interrupt, then give in-flight requests time.
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
