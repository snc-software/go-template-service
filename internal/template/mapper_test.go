package template

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	"github.com/snc-software/go-template-service/internal/platform/httpx"
)

func Test_ToDomain_Should_Return_CreateTemplate(t *testing.T) {
	t.Parallel()

	request := CreateRequest{Name: "Ada", Email: "ada@example.test"}

	assert.Equal(t, CreateTemplate{Name: "Ada", Email: "ada@example.test"}, request.toDomain())
}

func Test_ToResponse_Should_Return_Response(t *testing.T) {
	t.Parallel()

	id := uuid.New()
	created := time.Date(2026, time.August, 8, 10, 30, 0, 0, time.UTC)
	updated := created.Add(time.Hour)

	response := toResponse(Template{
		ID:        id,
		Name:      "Ada",
		Email:     "ada@example.test",
		CreatedAt: created,
		UpdatedAt: updated,
	})

	assert.Equal(t, Response{
		ID:        id,
		Name:      "Ada",
		Email:     "ada@example.test",
		CreatedAt: created,
		UpdatedAt: updated,
	}, response)
}

func Test_ToPagedResponse_Should_Return_Items_With_Pagination(t *testing.T) {
	t.Parallel()

	first := Template{ID: uuid.New(), Name: "Ada", Email: "ada@example.test"}
	second := Template{ID: uuid.New(), Name: "Grace", Email: "grace@example.test"}

	paged := toPagedResponse([]Template{first, second}, 2, 10, 25)

	assert.Equal(t, []Response{toResponse(first), toResponse(second)}, paged.Items)
	assert.Equal(t, httpx.Pagination{Page: 2, Size: 10, Total: 25}, paged.Pagination)
}

func Test_ToPagedResponse_Should_Return_Empty_Items_When_There_Are_No_Templates(t *testing.T) {
	t.Parallel()

	paged := toPagedResponse(nil, 1, 10, 0)

	assert.NotNil(t, paged.Items)
	assert.Empty(t, paged.Items)
}
