package template

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"

	"github.com/snc-software/go-template-service/internal/platform/apperr"
)

// Store is the persistence dependency of Service.
type Store interface {
	GetByID(ctx context.Context, id uuid.UUID) (Template, error)
	GetPage(ctx context.Context, page, size int) ([]Template, int, error)
	Create(ctx context.Context, create CreateTemplate) (Template, error)
	Delete(ctx context.Context, id uuid.UUID) error
}

type Service struct {
	store Store
}

func NewService(store Store) *Service {
	return &Service{store: store}
}

func (service *Service) GetByID(ctx context.Context, id uuid.UUID) (Template, error) {
	template, err := service.store.GetByID(ctx, id)

	switch {
	case errors.Is(err, ErrNotFound):
		return Template{}, apperr.NotFound(fmt.Sprintf("template %s not found", id))
	case err != nil:
		return Template{}, apperr.Internal(err)
	}

	return template, nil
}

func (service *Service) GetPage(ctx context.Context, page, size int) ([]Template, int, error) {
	templates, total, err := service.store.GetPage(ctx, page, size)
	if err != nil {
		return nil, 0, apperr.Internal(err)
	}

	return templates, total, nil
}

func (service *Service) Create(ctx context.Context, create CreateTemplate) (Template, error) {
	created, err := service.store.Create(ctx, create)

	switch {
	case errors.Is(err, ErrDuplicateEmail):
		return Template{}, apperr.Conflict(fmt.Sprintf("a template with email %q already exists", create.Email))
	case err != nil:
		return Template{}, apperr.Internal(err)
	}

	return created, nil
}

func (service *Service) Delete(ctx context.Context, id uuid.UUID) error {
	err := service.store.Delete(ctx, id)

	switch {
	case errors.Is(err, ErrNotFound):
		return apperr.NotFound(fmt.Sprintf("template %s not found", id))
	case err != nil:
		return apperr.Internal(err)
	}

	return nil
}
