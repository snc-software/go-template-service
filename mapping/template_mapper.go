package mapping

import (
	"github.com/snc-software/go-template-service/domain/models"
	"github.com/snc-software/go-template-service/routes/contracts"
)

func MapToDomain(createRequest contracts.CreateTemplateRequest) models.CreateTemplateModel {
	return models.CreateTemplateModel{
		Name:  createRequest.Name,
		Email: createRequest.Email,
	}
}

func MapToResponse(template models.TemplateModel) contracts.TemplateResponse {
	return contracts.TemplateResponse{
		ID:    template.ID,
		Name:  template.Name,
		Email: template.Email,
	}
}

func MapToPagedResponse(templates []models.TemplateModel, page, size, total int) contracts.PagedResponse[contracts.TemplateResponse] {
	items := make([]contracts.TemplateResponse, len(templates))
	for i, template := range templates {
		items[i] = MapToResponse(template)
	}

	return contracts.PagedResponse[contracts.TemplateResponse]{
		Items: items,
		Pagination: contracts.Pagination{
			Page:  page,
			Size:  size,
			Total: total,
		},
	}
}
