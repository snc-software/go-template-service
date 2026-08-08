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

type Handler struct {
	service   *Service
	responder *httpx.Responder
}

func NewHandler(service *Service, responder *httpx.Responder) *Handler {
	return &Handler{service: service, responder: responder}
}

func (handler *Handler) Routes() http.Handler {
	router := chi.NewRouter()
	router.Get("/", handler.GetPage)
	router.Get("/{id}", handler.GetByID)
	router.Post("/", handler.Create)
	router.Delete("/{id}", handler.Delete)

	return router
}

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
func (handler *Handler) GetByID(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r)
	if err != nil {
		handler.responder.Error(w, r, err)
		return
	}

	found, err := handler.service.GetByID(r.Context(), id)
	if err != nil {
		handler.responder.Error(w, r, err)
		return
	}

	handler.responder.OK(w, r, toResponse(found))
}

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
func (handler *Handler) GetPage(w http.ResponseWriter, r *http.Request) {
	page, err := positiveIntQuery(r, "page", defaultPage)
	if err != nil {
		handler.responder.Error(w, r, err)
		return
	}

	size, err := positiveIntQuery(r, "size", defaultPageSize)
	if err != nil {
		handler.responder.Error(w, r, err)
		return
	}

	if size > maxPageSize {
		size = maxPageSize
	}

	templates, total, err := handler.service.GetPage(r.Context(), page, size)
	if err != nil {
		handler.responder.Error(w, r, err)
		return
	}

	handler.responder.OK(w, r, toPagedResponse(templates, page, size, total))
}

// @Summary      Create template
// @Description  Creates a new template
// @Tags         Templates
// @Accept       json
// @Produce      json
// @Param        template  body      template.CreateRequest  true  "Template"
// @Success      201   {object}  template.Response
// @Failure      400   {object}  httpx.ProblemDetails
// @Failure      409   {object}  httpx.ProblemDetails
// @Failure      413   {object}  httpx.ProblemDetails
// @Failure      500   {object}  httpx.ProblemDetails
// @Router       /templates [post]
func (handler *Handler) Create(w http.ResponseWriter, r *http.Request) {
	request, err := httpx.Decode[CreateRequest](w, r)
	if err != nil {
		handler.responder.Error(w, r, err)
		return
	}

	created, err := handler.service.Create(r.Context(), request.toDomain())
	if err != nil {
		handler.responder.Error(w, r, err)
		return
	}

	handler.responder.Created(w, r, toResponse(created))
}

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
func (handler *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r)
	if err != nil {
		handler.responder.Error(w, r, err)
		return
	}

	if err := handler.service.Delete(r.Context(), id); err != nil {
		handler.responder.Error(w, r, err)
		return
	}

	handler.responder.NoContent(w)
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
