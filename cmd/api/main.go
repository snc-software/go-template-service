// Command api is the service entry point and composition root.
package main

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/joho/godotenv"

	"github.com/snc-software/go-template-service/internal/app"
	"github.com/snc-software/go-template-service/internal/platform/config"
	"github.com/snc-software/go-template-service/internal/platform/database"
)

const (
	readHeaderTimeout = 5 * time.Second
	readTimeout       = 15 * time.Second
	writeTimeout      = 30 * time.Second
	idleTimeout       = 60 * time.Second

	// Must exceed app.RequestTimeout, or the drain truncates in-flight work.
	shutdownTimeout = 20 * time.Second
)

// version is set at build time with -ldflags "-X main.version=...".
var version = "dev"

// @title                Template API
// @version              1.0
// @description          A basic template management API
// @servers.url          http://localhost:8080
// @servers.description  Local
// @tag.name             Templates
// @tag.description      Operations for managing templates
// @tag.name             Health
// @tag.description      Liveness and readiness probes
func main() {
	if err := run(); err != nil {
		slog.Error("startup failed", slog.Any("err", err))
		os.Exit(1)
	}
}

func run() error {
	if err := loadEnvFiles(); err != nil {
		return err
	}

	cfg, err := config.Load()
	if err != nil {
		return err
	}

	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: cfg.LogLevel}))
	slog.SetDefault(logger)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	db, err := database.New(ctx, cfg.Database)
	if err != nil {
		return err
	}
	defer func() { _ = db.Close() }()

	router, err := app.NewRouter(db, logger)
	if err != nil {
		return err
	}

	server := &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           router,
		ReadHeaderTimeout: readHeaderTimeout,
		ReadTimeout:       readTimeout,
		WriteTimeout:      writeTimeout,
		IdleTimeout:       idleTimeout,
	}

	serverFailed := make(chan error, 1)

	go func() {
		logger.Info("server listening",
			slog.String("addr", server.Addr),
			slog.String("version", version),
			slog.Any("database", cfg.Database),
		)

		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			serverFailed <- err
		}
	}()

	select {
	case err := <-serverFailed:
		return fmt.Errorf("serve: %w", err)
	case <-ctx.Done():
		logger.Info("shutdown signal received, draining")
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("graceful shutdown: %w", err)
	}

	logger.Info("shutdown complete")

	return nil
}

func loadEnvFiles() error {
	files := []struct {
		name string
		load func(...string) error
	}{
		{name: ".env", load: godotenv.Load},
		{name: ".env.local", load: godotenv.Overload},
	}

	for _, file := range files {
		if err := file.load(file.name); err != nil && !errors.Is(err, fs.ErrNotExist) {
			return fmt.Errorf("load %s: %w", file.name, err)
		}
	}

	return nil
}
