package template

import (
	"time"

	"github.com/google/uuid"

	"github.com/snc-software/go-template-service/internal/platform/httpx"
)

type CreateRequest struct {
	Name  string `json:"name"  validate:"required,min=1,max=255"`
	Email string `json:"email" validate:"required,email,max=255"`
}

type Response struct {
	ID        uuid.UUID `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

type PagedResponse struct {
	Items      []Response       `json:"items"`
	Pagination httpx.Pagination `json:"pagination"`
}
