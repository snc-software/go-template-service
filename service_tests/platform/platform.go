//go:build integration

// Package platform boots what every service test needs: a PostgreSQL container,
// the schema goose builds inside it, and the real router serving over a local
// listener. Nothing about the application is substituted.
package platform

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/jmoiron/sqlx"
	"github.com/pressly/goose/v3"
	"github.com/pressly/goose/v3/lock"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"

	"github.com/snc-software/go-template-service/internal/app"
	"github.com/snc-software/go-template-service/internal/platform/config"
	"github.com/snc-software/go-template-service/internal/platform/database"
	"github.com/snc-software/go-template-service/migrations"
)

const (
	image         = "postgres:17-alpine"
	containerName = "go-template-service-service-tests"
	databaseName  = "template_service"
	databaseUser  = "postgres"
	databasePass  = "postgres"
)

// API is the running service under test.
type API struct {
	// BaseURL is the root the service is listening on.
	BaseURL string
	// DB is a pool on the same database the service is using, for seeding rows
	// and for asserting on what an endpoint left behind.
	DB *sqlx.DB
	// Client is an HTTP client configured for the running server.
	Client *http.Client
}

var running *API

// Run boots the service and then runs m. Every service test package calls it
// from TestMain.
func Run(m *testing.M) int {
	api, shutdown, err := start(context.Background())
	defer shutdown()

	if err != nil {
		fmt.Fprintf(os.Stderr, "service test setup: %v\n", err)

		return 1
	}

	running = api

	return m.Run()
}

// Current returns the service booted by Run.
func Current(t *testing.T) *API {
	t.Helper()

	if running == nil {
		t.Fatal("platform.Run was not called from TestMain")
	}

	return running
}

func start(ctx context.Context) (*API, func(), error) {
	var closers []func()

	shutdown := func() {
		for i := len(closers) - 1; i >= 0; i-- {
			closers[i]()
		}
	}

	// Every test package that boots the platform attaches to this one container
	// rather than starting its own. It is deliberately not terminated here: the
	// package that finishes first would take the database away from the ones
	// still running. Testcontainers' reaper owns it instead, and tears it down
	// when the whole "go test" session ends.
	container, err := postgres.Run(ctx, image,
		postgres.WithDatabase(databaseName),
		postgres.WithUsername(databaseUser),
		postgres.WithPassword(databasePass),
		postgres.BasicWaitStrategies(),
		testcontainers.WithReuseByName(containerName),
	)
	if err != nil {
		return nil, shutdown, fmt.Errorf("start postgres: %w", err)
	}

	cfg, err := databaseConfig(ctx, container)
	if err != nil {
		return nil, shutdown, err
	}

	db, err := database.New(ctx, cfg)
	if err != nil {
		return nil, shutdown, err
	}

	closers = append(closers, func() { _ = db.Close() })

	err = migrate(ctx, db)
	if err != nil {
		return nil, shutdown, err
	}

	router, err := app.NewRouter(db, logger())
	if err != nil {
		return nil, shutdown, err
	}

	server := httptest.NewServer(router)
	closers = append(closers, server.Close)

	return &API{BaseURL: server.URL, DB: db, Client: server.Client()}, shutdown, nil
}

func databaseConfig(ctx context.Context, container *postgres.PostgresContainer) (config.Database, error) {
	host, err := container.Host(ctx)
	if err != nil {
		return config.Database{}, fmt.Errorf("container host: %w", err)
	}

	port, err := container.MappedPort(ctx, "5432/tcp")
	if err != nil {
		return config.Database{}, fmt.Errorf("container port: %w", err)
	}

	return config.Database{
		Host:     host,
		Port:     port.Port(),
		Name:     databaseName,
		User:     databaseUser,
		Password: databasePass,
		SSLMode:  "disable",
	}, nil
}

// migrate brings the shared database up to date. Test packages run as separate
// processes against the same container, so the advisory lock is what stops two
// of them migrating at once.
func migrate(ctx context.Context, db *sqlx.DB) error {
	locker, err := lock.NewPostgresSessionLocker()
	if err != nil {
		return fmt.Errorf("create migration locker: %w", err)
	}

	provider, err := goose.NewProvider(goose.DialectPostgres, db.DB, migrations.FS,
		goose.WithSessionLocker(locker))
	if err != nil {
		return fmt.Errorf("create migration provider: %w", err)
	}

	if _, err := provider.Up(ctx); err != nil {
		return fmt.Errorf("apply migrations: %w", err)
	}

	return nil
}

// logger keeps request logging out of test output while still surfacing the
// server-side detail behind a 5xx.
func logger() *slog.Logger {
	return slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))
}
