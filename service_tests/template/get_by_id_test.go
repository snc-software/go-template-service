//go:build integration

package templatetests

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/snc-software/go-template-service/internal/platform/apperr"
	"github.com/snc-software/go-template-service/internal/platform/httpx"
	"github.com/snc-software/go-template-service/internal/template"
	"github.com/snc-software/go-template-service/service_tests/platform"
)

func Test_GetById_Should_Return_Template(t *testing.T) {
	t.Parallel()

	api := platform.Current(t)
	seeded := seedTemplate(t, api.DB)

	response := api.Get(t, "/templates/"+seeded.ID.String())

	require.Equal(t, http.StatusOK, response.StatusCode)
	require.Equal(t, "application/json", response.ContentType)
	assert.Equal(t, []string{"createdAt", "email", "id", "name", "updatedAt"}, response.JSONFields(t))

	var body template.Response
	response.DecodeJSON(t, &body)

	body.CreatedAt = body.CreatedAt.UTC()
	body.UpdatedAt = body.UpdatedAt.UTC()

	assert.Equal(t, template.Response{
		ID:        seeded.ID,
		Name:      seeded.Name,
		Email:     seeded.Email,
		CreatedAt: seeded.CreatedAt,
		UpdatedAt: seeded.UpdatedAt,
	}, body)
}

func Test_GetById_Should_Return_NotFound_When_Template_Does_Not_Exist(t *testing.T) {
	t.Parallel()

	api := platform.Current(t)
	id := uuid.New()

	response := api.Get(t, "/templates/"+id.String())

	require.Equal(t, http.StatusNotFound, response.StatusCode)
	require.Equal(t, "application/problem+json", response.ContentType)
	assert.Equal(t, []string{"code", "message", "status"}, response.JSONFields(t))

	var problem httpx.ProblemDetails
	response.DecodeJSON(t, &problem)

	assert.Equal(t, httpx.ProblemDetails{
		Status:  http.StatusNotFound,
		Code:    apperr.CodeNotFound,
		Message: fmt.Sprintf("template %s not found", id),
	}, problem)
}
