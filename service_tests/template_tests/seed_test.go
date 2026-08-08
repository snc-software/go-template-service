//go:build integration

package templatetests

import (
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/require"

	"github.com/snc-software/go-template-service/internal/template"
)

const insertTemplate = `INSERT INTO "Templates" ("Id", "Name", "Email", "CreatedAt", "UpdatedAt")
                        VALUES (:Id, :Name, :Email, :CreatedAt, :UpdatedAt)`

// seedTemplate inserts a template straight into the database and returns it.
// The ID and the email are fresh on every call, so no two tests collide on the
// shared database.
func seedTemplate(t *testing.T, db *sqlx.DB) template.Template {
	t.Helper()

	id := uuid.New()

	// Postgres stores timestamptz to microsecond precision, so a nanosecond in
	// the seed would not survive the round trip.
	now := time.Now().UTC().Truncate(time.Microsecond)

	seeded := template.Template{
		ID:        id,
		Name:      "Template " + id.String(),
		Email:     fmt.Sprintf("%s@example.test", id),
		CreatedAt: now,
		UpdatedAt: now,
	}

	_, err := db.NamedExecContext(t.Context(), insertTemplate, seeded)
	require.NoError(t, err, "seed template %s", id)

	return seeded
}
