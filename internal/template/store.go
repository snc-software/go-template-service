package template

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
)

// Storage failures the service translates into API errors.
var (
	ErrNotFound       = errors.New("template not found")
	ErrDuplicateEmail = errors.New("template email already exists")
)

const uniqueViolation = pq.ErrorCode("23505")

const columns = `"Id", "Name", "Email", "CreatedAt", "UpdatedAt"`

type PostgresStore struct {
	db *sqlx.DB
}

func NewPostgresStore(db *sqlx.DB) *PostgresStore {
	return &PostgresStore{db: db}
}

func (store *PostgresStore) GetByID(ctx context.Context, id uuid.UUID) (Template, error) {
	var template Template

	err := store.db.GetContext(ctx, &template,
		`SELECT `+columns+` FROM "Templates" WHERE "Id" = $1`, id)

	switch {
	case errors.Is(err, sql.ErrNoRows):
		return Template{}, fmt.Errorf("select template %s: %w", id, ErrNotFound)
	case err != nil:
		return Template{}, fmt.Errorf("select template %s: %w", id, err)
	}

	return template, nil
}

func (store *PostgresStore) GetPage(ctx context.Context, page, size int) ([]Template, int, error) {
	var total int
	if err := store.db.GetContext(ctx, &total, `SELECT COUNT(*) FROM "Templates"`); err != nil {
		return nil, 0, fmt.Errorf("count templates: %w", err)
	}

	templates := []Template{}
	offset := (page - 1) * size

	err := store.db.SelectContext(ctx, &templates,
		`SELECT `+columns+` FROM "Templates" ORDER BY "CreatedAt" DESC, "Id" DESC LIMIT $1 OFFSET $2`,
		size, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("select templates page %d size %d: %w", page, size, err)
	}

	return templates, total, nil
}

func (store *PostgresStore) Create(ctx context.Context, create CreateTemplate) (Template, error) {
	var created Template

	err := store.db.GetContext(ctx, &created,
		`INSERT INTO "Templates" ("Id", "Name", "Email") VALUES ($1, $2, $3) RETURNING `+columns,
		uuid.New(), create.Name, create.Email)

	switch {
	case isUniqueViolation(err):
		return Template{}, fmt.Errorf("insert template: %w", ErrDuplicateEmail)
	case err != nil:
		return Template{}, fmt.Errorf("insert template: %w", err)
	}

	return created, nil
}

func (store *PostgresStore) Delete(ctx context.Context, id uuid.UUID) error {
	result, err := store.db.ExecContext(ctx, `DELETE FROM "Templates" WHERE "Id" = $1`, id)
	if err != nil {
		return fmt.Errorf("delete template %s: %w", id, err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("delete template %s: %w", id, err)
	}

	if rows == 0 {
		return fmt.Errorf("delete template %s: %w", id, ErrNotFound)
	}

	return nil
}

func isUniqueViolation(err error) bool {
	var pqError *pq.Error

	return errors.As(err, &pqError) && pqError.Code == uniqueViolation
}
