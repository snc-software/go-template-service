package models

import "github.com/google/uuid"

type TemplateModel struct {
	ID    uuid.UUID
	Name  string
	Email string
}
