// Package template is the example resource.
package template

import (
	"time"

	"github.com/google/uuid"
)

type Template struct {
	ID        uuid.UUID `db:"Id"`
	Name      string    `db:"Name"`
	Email     string    `db:"Email"`
	CreatedAt time.Time `db:"CreatedAt"`
	UpdatedAt time.Time `db:"UpdatedAt"`
}

type CreateTemplate struct {
	Name  string
	Email string
}
