package routes

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/snc-software/go-template-service/domain/services"
	"github.com/snc-software/go-template-service/exceptions"
	"github.com/snc-software/go-template-service/mapping"
	"github.com/snc-software/go-template-service/routes/contracts"
	"github.com/snc-software/go-template-service/utils"
)

type TemplateRouteHandler struct {
	templateService services.TemplateService
}

func TemplateRoutes(templateService services.TemplateService) http.Handler {
	handler := TemplateRouteHandler{templateService: templateService}

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
// @Success      200  {object}  contracts.TemplateResponse
// @Failure      400  {string}  string  "invalid id"
// @Failure      404  {string}  string  "not found"
// @Router       /templates/{id} [get]
func (handler TemplateRouteHandler) GetByID(responseWriter http.ResponseWriter, request *http.Request) {
	idStr := chi.URLParam(request, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		utils.HandleError(responseWriter, exceptions.InvalidArgument("invalid id format"))
		return
	}

	template, err := handler.templateService.GetByID(id)
	if err != nil {
		utils.HandleError(responseWriter, err)
		return
	}

	utils.OkResponse(responseWriter, mapping.MapToResponse(template))
}

// @Summary      Get paged templates
// @Description  Returns a paged list of templates
// @Tags         Templates
// @Produce      json
// @Param        page   query  int  false  "Page number"   default(1)
// @Param        size   query  int  false  "Page size"     default(10)
// @Success      200    {object}  contracts.PagedResponse[contracts.TemplateResponse]
// @Failure      500    {string}  string  "internal error"
// @Router       /templates [get]
func (handler TemplateRouteHandler) GetPage(responseWriter http.ResponseWriter, request *http.Request) {
	page, err := strconv.Atoi(request.URL.Query().Get("page"))
	if err != nil || page < 1 {
		page = 1
	}

	size, err := strconv.Atoi(request.URL.Query().Get("size"))
	if err != nil || size < 1 {
		size = 10
	}

	templates, total, err := handler.templateService.GetPage(page, size)
	if err != nil {
		utils.HandleError(responseWriter, exceptions.Internal())
		return
	}

	pagedResponse := mapping.MapToPagedResponse(templates, page, size, total)

	utils.OkResponse(responseWriter, pagedResponse)
}

// @Summary      Create template
// @Description  Creates a new template
// @Tags         Templates
// @Accept       json
// @Produce      json
// @Param        template  body      contracts.CreateTemplateRequest  true  "Template"
// @Success      201   {object}  contracts.TemplateResponse
// @Failure      400   {string}  string  "invalid request body"
// @Failure      500   {string}  string  "internal error"
// @Router       /templates [post]
func (handler TemplateRouteHandler) Create(responseWriter http.ResponseWriter, request *http.Request) {
	var createRequest contracts.CreateTemplateRequest
	err := json.NewDecoder(request.Body).Decode(&createRequest)
	if err != nil {
		utils.HandleError(responseWriter, exceptions.InvalidArgument("invalid request body"))
		return
	}

	domainCreateRequest := mapping.MapToDomain(createRequest)
	created, err := handler.templateService.Create(domainCreateRequest)
	if err != nil {
		utils.HandleError(responseWriter, exceptions.Internal())
		return
	}

	templateResponse := mapping.MapToResponse(created)

	utils.CreatedResponse(responseWriter, templateResponse)
}

// @Summary      Delete template by ID
// @Description  Deletes a single template by ID
// @Tags         Templates
// @Produce      json
// @Param        id   path      string  true  "Template ID"
// @Success      204
// @Failure      400  {string}  string  "invalid id"
// @Failure      404  {string}  string  "not found"
// @Router       /templates/{id} [delete]
func (handler TemplateRouteHandler) Delete(responseWriter http.ResponseWriter, request *http.Request) {
	idStr := chi.URLParam(request, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		utils.HandleError(responseWriter, exceptions.InvalidArgument("invalid id format"))
		return
	}

	deleteErr := handler.templateService.Delete(id)
	if deleteErr != nil {
		utils.HandleError(responseWriter, deleteErr)
		return
	}

	utils.NoContentResponse(responseWriter)
}
