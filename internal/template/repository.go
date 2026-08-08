package template

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"

	"github.com/snc-software/go-template-service/internal/platform/database"
)

// Storage failures the service translates into API errors.
var (
	ErrNotFound       = errors.New("template not found")
	ErrDuplicateEmail = errors.New("template email already exists")
)

const columns = `"Id", "Name", "Email", "CreatedAt", "UpdatedAt"`

// PostgresRepository stores templates in PostgreSQL.
type PostgresRepository struct {
	db *sqlx.DB
}

// NewPostgresRepository returns a PostgresRepository backed by db.
func NewPostgresRepository(db *sqlx.DB) *PostgresRepository {
	return &PostgresRepository{db: db}
}

// GetByID returns the template with the given ID, or ErrNotFound.
func (repository *PostgresRepository) GetByID(ctx context.Context, id uuid.UUID) (Template, error) {
	var template Template

	err := repository.db.GetContext(ctx, &template,
		`SELECT `+columns+` FROM "Templates" WHERE "Id" = $1`, id)

	switch {
	case errors.Is(err, sql.ErrNoRows):
		return Template{}, fmt.Errorf("select template %s: %w", id, ErrNotFound)
	case err != nil:
		return Template{}, fmt.Errorf("select template %s: %w", id, err)
	}

	return template, nil
}

// GetPage returns one page of templates, newest first, and the total count.
func (repository *PostgresRepository) GetPage(ctx context.Context, page, size int) ([]Template, int, error) {
	var total int
	if err := repository.db.GetContext(ctx, &total, `SELECT COUNT(*) FROM "Templates"`); err != nil {
		return nil, 0, fmt.Errorf("count templates: %w", err)
	}

	templates := []Template{}
	offset := (page - 1) * size

	err := repository.db.SelectContext(ctx, &templates,
		`SELECT `+columns+` FROM "Templates" ORDER BY "CreatedAt" DESC, "Id" DESC LIMIT $1 OFFSET $2`,
		size, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("select templates page %d size %d: %w", page, size, err)
	}

	return templates, total, nil
}

// Create inserts a template and returns it as stored, or ErrDuplicateEmail.
func (repository *PostgresRepository) Create(ctx context.Context, create CreateTemplate) (Template, error) {
	var created Template

	err := repository.db.GetContext(ctx, &created,
		`INSERT INTO "Templates" ("Id", "Name", "Email") VALUES ($1, $2, $3) RETURNING `+columns,
		uuid.New(), create.Name, create.Email)

	switch {
	case database.IsUniqueViolation(err):
		return Template{}, fmt.Errorf("insert template: %w", ErrDuplicateEmail)
	case err != nil:
		return Template{}, fmt.Errorf("insert template: %w", err)
	}

	return created, nil
}

// Delete removes the template with the given ID, or returns ErrNotFound.
func (repository *PostgresRepository) Delete(ctx context.Context, id uuid.UUID) error {
	result, err := repository.db.ExecContext(ctx, `DELETE FROM "Templates" WHERE "Id" = $1`, id)
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
