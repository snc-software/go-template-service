// Package health serves the liveness and readiness probes.
package health

import (
	"context"
	"net/http"
	"time"

	"github.com/snc-software/go-template-service/internal/platform/apperr"
	"github.com/snc-software/go-template-service/internal/platform/httpx"
)

const readyTimeout = 2 * time.Second

// Pinger is the readiness dependency: anything that can be asked whether it is
// still reachable.
type Pinger interface {
	PingContext(ctx context.Context) error
}

// Status is the probe response body.
type Status struct {
	Status string `json:"status"`
}

// Endpoints serves the probes.
type Endpoints struct {
	pinger    Pinger
	responder *httpx.Responder
}

// NewEndpoints returns Endpoints that check pinger for readiness.
func NewEndpoints(pinger Pinger, responder *httpx.Responder) *Endpoints {
	return &Endpoints{pinger: pinger, responder: responder}
}

// Live is the liveness probe.
//
// @Summary      Liveness probe
// @Description  Returns 200 while the process is running. Never touches dependencies.
// @Tags         Health
// @Produce      json
// @Success      200  {object}  health.Status
// @Router       /health [get]
func (endpoints *Endpoints) Live(w http.ResponseWriter, r *http.Request) {
	endpoints.responder.OK(w, r, Status{Status: "ok"})
}

// Ready is the readiness probe.
//
// @Summary      Readiness probe
// @Description  Returns 200 when the service can reach its dependencies, 503 otherwise.
// @Tags         Health
// @Produce      json
// @Success      200  {object}  health.Status
// @Failure      503  {object}  httpx.ProblemDetails
// @Router       /ready [get]
func (endpoints *Endpoints) Ready(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), readyTimeout)
	defer cancel()

	if err := endpoints.pinger.PingContext(ctx); err != nil {
		endpoints.responder.Error(w, r, apperr.Unavailable("database is not reachable", err))

		return
	}

	endpoints.responder.OK(w, r, Status{Status: "ready"})
}
