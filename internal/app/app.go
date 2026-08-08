// Package app builds the HTTP router from its dependencies. It is the single
// definition of the service's routes and middleware, used by the binary and by
// the service tests alike.
package app

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/jmoiron/sqlx"

	"github.com/snc-software/go-template-service/docs"
	"github.com/snc-software/go-template-service/internal/health"
	"github.com/snc-software/go-template-service/internal/platform/httpx"
	"github.com/snc-software/go-template-service/internal/platform/openapi"
	"github.com/snc-software/go-template-service/internal/template"
)

const apiTitle = "Template API"

// RequestTimeout bounds a single request. Anything that drains the server must
// outlast it.
const RequestTimeout = 15 * time.Second

// NewRouter wires every feature into the router and returns it ready to serve.
func NewRouter(db *sqlx.DB, logger *slog.Logger) (http.Handler, error) {
	reference, err := openapi.NewEndpoints(apiTitle, docs.OpenAPI)
	if err != nil {
		return nil, err
	}

	responder := httpx.NewResponder(logger)
	templates := template.NewEndpoints(template.NewService(template.NewPostgresRepository(db)), responder)
	probes := health.NewEndpoints(db, responder)

	router := chi.NewRouter()
	router.Use(httpx.Recoverer(logger))

	router.Get("/health", probes.Live)
	router.Get("/ready", probes.Ready)

	router.Group(func(r chi.Router) {
		r.Use(httpx.RequestLogger(logger))
		r.Use(middleware.Timeout(RequestTimeout))

		r.Get(openapi.ReferencePath, reference.Reference)
		r.Get(openapi.SpecPath, reference.Spec)
		r.Mount("/templates", templates.Routes())
	})

	return router, nil
}
