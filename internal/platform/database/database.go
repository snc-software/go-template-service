// Package database opens and tunes the process-wide connection pool.
package database

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"

	"github.com/snc-software/go-template-service/internal/platform/config"
)

const (
	maxOpenConns    = 25
	maxIdleConns    = 25
	connMaxLifetime = 5 * time.Minute
	connMaxIdleTime = 5 * time.Minute
)

// New opens the connection pool and verifies the database is reachable.
func New(ctx context.Context, cfg config.Database) (*sqlx.DB, error) {
	db, err := sqlx.ConnectContext(ctx, "pgx", cfg.DSN())
	if err != nil {
		return nil, fmt.Errorf("connect to database: %w", err)
	}

	db.SetMaxOpenConns(maxOpenConns)
	db.SetMaxIdleConns(maxIdleConns)
	db.SetConnMaxLifetime(connMaxLifetime)
	db.SetConnMaxIdleTime(connMaxIdleTime)

	return db, nil
}

// IsUniqueViolation reports whether err is a PostgreSQL unique constraint
// failure. It is the one driver-specific check every repository needs, and
// keeping it here is what lets the rest of a repository stay driver-agnostic.
func IsUniqueViolation(err error) bool {
	var pgError *pgconn.PgError

	return errors.As(err, &pgError) && pgError.Code == pgerrcode.UniqueViolation
}
