// Package template is the example resource, wired end to end.
package template

import (
	"time"

	"github.com/google/uuid"
)

// Template is the domain type. It is never serialised directly.
type Template struct {
	ID        uuid.UUID `db:"Id"`
	Name      string    `db:"Name"`
	Email     string    `db:"Email"`
	CreatedAt time.Time `db:"CreatedAt"`
	UpdatedAt time.Time `db:"UpdatedAt"`
}

// CreateTemplate is everything needed to store a new template.
type CreateTemplate struct {
	Name  string
	Email string
}
