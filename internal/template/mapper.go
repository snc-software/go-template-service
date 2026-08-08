package template

import (
	"github.com/snc-software/go-template-service/internal/platform/httpx"
)

func (request CreateRequest) toDomain() CreateTemplate {
	return CreateTemplate{
		Name:  request.Name,
		Email: request.Email,
	}
}

func toResponse(template Template) Response {
	return Response{
		ID:        template.ID,
		Name:      template.Name,
		Email:     template.Email,
		CreatedAt: template.CreatedAt,
		UpdatedAt: template.UpdatedAt,
	}
}

func toPagedResponse(templates []Template, page, size, total int) PagedResponse {
	items := make([]Response, len(templates))
	for i, template := range templates {
		items[i] = toResponse(template)
	}

	return PagedResponse{
		Items: items,
		Pagination: httpx.Pagination{
			Page:  page,
			Size:  size,
			Total: total,
		},
	}
}
