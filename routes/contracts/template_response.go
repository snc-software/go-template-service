package contracts

import "github.com/google/uuid"

// @name TemplateResponse
type TemplateResponse struct {
	ID    uuid.UUID `json:"id"`
	Name  string    `json:"name"`
	Email string    `json:"email"`
}
