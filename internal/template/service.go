package template

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"

	"github.com/snc-software/go-template-service/internal/platform/apperr"
)

// Repository is the persistence dependency of Service.
type Repository interface {
	GetByID(ctx context.Context, id uuid.UUID) (Template, error)
	GetPage(ctx context.Context, page, size int) ([]Template, int, error)
	Create(ctx context.Context, create CreateTemplate) (Template, error)
	Delete(ctx context.Context, id uuid.UUID) error
}

// Service holds the business rules for the template resource.
type Service struct {
	repository Repository
}

// NewService returns a Service backed by repository.
func NewService(repository Repository) *Service {
	return &Service{repository: repository}
}

// GetByID returns one template, or a NOT_FOUND application error.
func (service *Service) GetByID(ctx context.Context, id uuid.UUID) (Template, error) {
	template, err := service.repository.GetByID(ctx, id)

	switch {
	case errors.Is(err, ErrNotFound):
		return Template{}, apperr.NotFound(fmt.Sprintf("template %s not found", id))
	case err != nil:
		return Template{}, apperr.Internal(err)
	}

	return template, nil
}

// GetPage returns one page of templates and the total number of them.
func (service *Service) GetPage(ctx context.Context, page, size int) ([]Template, int, error) {
	templates, total, err := service.repository.GetPage(ctx, page, size)
	if err != nil {
		return nil, 0, apperr.Internal(err)
	}

	return templates, total, nil
}

// Create stores a new template, or returns a CONFLICT application error if the
// email is already taken.
func (service *Service) Create(ctx context.Context, create CreateTemplate) (Template, error) {
	created, err := service.repository.Create(ctx, create)

	switch {
	case errors.Is(err, ErrDuplicateEmail):
		return Template{}, apperr.Conflict(fmt.Sprintf("a template with email %q already exists", create.Email))
	case err != nil:
		return Template{}, apperr.Internal(err)
	}

	return created, nil
}

// Delete removes one template, or returns a NOT_FOUND application error.
func (service *Service) Delete(ctx context.Context, id uuid.UUID) error {
	err := service.repository.Delete(ctx, id)

	switch {
	case errors.Is(err, ErrNotFound):
		return apperr.NotFound(fmt.Sprintf("template %s not found", id))
	case err != nil:
		return apperr.Internal(err)
	}

	return nil
}
