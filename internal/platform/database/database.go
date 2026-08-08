// Package database opens and tunes the process-wide connection pool.
package database

import (
	"context"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"

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
	db, err := sqlx.ConnectContext(ctx, "postgres", cfg.DSN())
	if err != nil {
		return nil, fmt.Errorf("connect to database: %w", err)
	}

	db.SetMaxOpenConns(maxOpenConns)
	db.SetMaxIdleConns(maxIdleConns)
	db.SetConnMaxLifetime(connMaxLifetime)
	db.SetConnMaxIdleTime(connMaxIdleTime)

	return db, nil
}
