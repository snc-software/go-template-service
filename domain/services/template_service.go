package services

import (
	"fmt"

	"github.com/google/uuid"

	"github.com/snc-software/go-template-service/domain/models"
	"github.com/snc-software/go-template-service/exceptions"
	"github.com/snc-software/go-template-service/persistence/entities"
	"github.com/snc-software/go-template-service/persistence/readers"
	"github.com/snc-software/go-template-service/persistence/writers"
)

type TemplateService struct{}

func (TemplateService) GetByID(id uuid.UUID) (models.TemplateModel, error) {
	template, err := readers.TemplateReader{}.GetByID(id)
	if err != nil {

		return models.TemplateModel{}, exceptions.NotFound(fmt.Sprintf("Template with id '%s' not found", id))
	}

	return models.TemplateModel{
		ID:    template.ID,
		Name:  template.Name,
		Email: template.Email,
	}, nil
}

func (TemplateService) Create(createRequest models.CreateTemplateModel) (models.TemplateModel, error) {
	template, err := writers.TemplateWriter{}.Create(entities.Template{
		Name:  createRequest.Name,
		Email: createRequest.Email,
	})
	if err != nil {
		return models.TemplateModel{}, exceptions.Internal()
	}

	return models.TemplateModel{
		ID:    template.ID,
		Name:  template.Name,
		Email: template.Email,
	}, nil
}

func (TemplateService) GetPage(page, size int) ([]models.TemplateModel, int, error) {
	templates, total, err := readers.TemplateReader{}.GetPage(page, size)
	if err != nil {
		return nil, 0, exceptions.Internal()
	}

	items := make([]models.TemplateModel, len(templates))
	for i, template := range templates {
		items[i] = models.TemplateModel{
			ID:    template.ID,
			Name:  template.Name,
			Email: template.Email,
		}
	}

	return items, total, nil
}

func (TemplateService) Delete(id uuid.UUID) error {
	_, err := readers.TemplateReader{}.GetByID(id)
	if err != nil {
		return exceptions.NotFound(fmt.Sprintf("Template with id '%s' not found", id))
	}

	return writers.TemplateWriter{}.Delete(id)
}
