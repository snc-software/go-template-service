package template

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/snc-software/go-template-service/internal/platform/apperr"
	"github.com/snc-software/go-template-service/internal/platform/httpx"
)

const (
	defaultPage     = 1
	defaultPageSize = 10
	maxPageSize     = 100
)

// Endpoints is the HTTP surface of the template resource.
type Endpoints struct {
	service   *Service
	responder *httpx.Responder
}

// NewEndpoints returns Endpoints backed by service.
func NewEndpoints(service *Service, responder *httpx.Responder) *Endpoints {
	return &Endpoints{service: service, responder: responder}
}

// Routes returns the router for this resource, ready to be mounted.
func (endpoints *Endpoints) Routes() http.Handler {
	router := chi.NewRouter()
	router.Get("/", endpoints.GetPage)
	router.Get("/{id}", endpoints.GetByID)
	router.Post("/", endpoints.Create)
	router.Delete("/{id}", endpoints.Delete)

	return router
}

// GetByID handles GET /templates/{id}.
//
// @Summary      Get template by ID
// @Description  Returns a single template by ID
// @Tags         Templates
// @Produce      json
// @Param        id   path      string  true  "Template ID"
// @Success      200  {object}  template.Response
// @Failure      400  {object}  httpx.ProblemDetails
// @Failure      404  {object}  httpx.ProblemDetails
// @Failure      500  {object}  httpx.ProblemDetails
// @Router       /templates/{id} [get]
func (endpoints *Endpoints) GetByID(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r)
	if err != nil {
		endpoints.responder.Error(w, r, err)
		return
	}

	found, err := endpoints.service.GetByID(r.Context(), id)
	if err != nil {
		endpoints.responder.Error(w, r, err)
		return
	}

	endpoints.responder.OK(w, r, toResponse(found))
}

// GetPage handles GET /templates.
//
// @Summary      Get paged templates
// @Description  Returns a paged list of templates
// @Tags         Templates
// @Produce      json
// @Param        page   query  int  false  "Page number"                    default(1)
// @Param        size   query  int  false  "Page size, capped at 100"       default(10)
// @Success      200    {object}  template.PagedResponse
// @Failure      400    {object}  httpx.ProblemDetails
// @Failure      500    {object}  httpx.ProblemDetails
// @Router       /templates [get]
func (endpoints *Endpoints) GetPage(w http.ResponseWriter, r *http.Request) {
	page, err := positiveIntQuery(r, "page", defaultPage)
	if err != nil {
		endpoints.responder.Error(w, r, err)
		return
	}

	size, err := positiveIntQuery(r, "size", defaultPageSize)
	if err != nil {
		endpoints.responder.Error(w, r, err)
		return
	}

	if size > maxPageSize {
		size = maxPageSize
	}

	templates, total, err := endpoints.service.GetPage(r.Context(), page, size)
	if err != nil {
		endpoints.responder.Error(w, r, err)
		return
	}

	endpoints.responder.OK(w, r, toPagedResponse(templates, page, size, total))
}

// Create handles POST /templates.
//
// @Summary      Create template
// @Description  Creates a new template
// @Tags         Templates
// @Accept       json
// @Produce      json
// @Param        request   body      template.CreateRequest  true  "Template"
// @Success      201   {object}  template.Response
// @Failure      400   {object}  httpx.ProblemDetails
// @Failure      409   {object}  httpx.ProblemDetails
// @Failure      413   {object}  httpx.ProblemDetails
// @Failure      500   {object}  httpx.ProblemDetails
// @Router       /templates [post]
func (endpoints *Endpoints) Create(w http.ResponseWriter, r *http.Request) {
	request, err := httpx.Decode[CreateRequest](w, r)
	if err != nil {
		endpoints.responder.Error(w, r, err)
		return
	}

	created, err := endpoints.service.Create(r.Context(), request.toDomain())
	if err != nil {
		endpoints.responder.Error(w, r, err)
		return
	}

	endpoints.responder.Created(w, r, toResponse(created))
}

// Delete handles DELETE /templates/{id}.
//
// @Summary      Delete template by ID
// @Description  Deletes a single template by ID
// @Tags         Templates
// @Produce      json
// @Param        id   path      string  true  "Template ID"
// @Success      204
// @Failure      400  {object}  httpx.ProblemDetails
// @Failure      404  {object}  httpx.ProblemDetails
// @Failure      500  {object}  httpx.ProblemDetails
// @Router       /templates/{id} [delete]
func (endpoints *Endpoints) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r)
	if err != nil {
		endpoints.responder.Error(w, r, err)
		return
	}

	if err := endpoints.service.Delete(r.Context(), id); err != nil {
		endpoints.responder.Error(w, r, err)
		return
	}

	endpoints.responder.NoContent(w)
}

func pathID(r *http.Request) (uuid.UUID, error) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		return uuid.Nil, apperr.InvalidArgument("id must be a UUID")
	}

	return id, nil
}

func positiveIntQuery(r *http.Request, name string, fallback int) (int, error) {
	raw := r.URL.Query().Get(name)
	if raw == "" {
		return fallback, nil
	}

	value, err := strconv.Atoi(raw)
	if err != nil || value < 1 {
		return 0, apperr.InvalidArgument(fmt.Sprintf("%s must be a positive integer", name))
	}

	return value, nil
}
