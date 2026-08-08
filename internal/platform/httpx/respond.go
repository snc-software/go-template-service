// Package httpx holds the transport-level plumbing shared by every feature:
// response writing, request decoding, and middleware.
package httpx

import (
	"bytes"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5/middleware"

	"github.com/snc-software/go-template-service/internal/platform/apperr"
)

// ProblemDetails is the single error shape for the API (RFC 7807).
type ProblemDetails struct {
	Status  int                 `json:"status"`
	Code    apperr.Code         `json:"code"`
	Message string              `json:"message"`
	Errors  []apperr.FieldError `json:"errors,omitempty"`
}

// Pagination reports the effective page and size, after any clamping.
type Pagination struct {
	Page  int `json:"page"`
	Size  int `json:"size"`
	Total int `json:"total"`
}

// Responder writes responses and is the single place a failed request is logged.
type Responder struct {
	logger *slog.Logger
}

func NewResponder(logger *slog.Logger) *Responder {
	return &Responder{logger: logger}
}

func (rs *Responder) OK(w http.ResponseWriter, r *http.Request, data any) {
	rs.json(w, r, http.StatusOK, data)
}

func (rs *Responder) Created(w http.ResponseWriter, r *http.Request, data any) {
	rs.json(w, r, http.StatusCreated, data)
}

func (rs *Responder) NoContent(w http.ResponseWriter) {
	w.WriteHeader(http.StatusNoContent)
}

func (rs *Responder) Error(w http.ResponseWriter, r *http.Request, err error) {
	appError := asAppError(err)

	if appError.Status >= http.StatusInternalServerError {
		rs.logger.ErrorContext(r.Context(), "request failed",
			slog.String("method", r.Method),
			slog.String("path", r.URL.Path),
			slog.String("code", string(appError.Code)),
			slog.String("request_id", middleware.GetReqID(r.Context())),
			slog.Any("err", err),
		)
	}

	WriteProblem(w, appError)
}

func (rs *Responder) json(w http.ResponseWriter, r *http.Request, status int, data any) {
	body, err := encode(data)
	if err != nil {
		rs.logger.ErrorContext(r.Context(), "encode response body",
			slog.String("method", r.Method),
			slog.String("path", r.URL.Path),
			slog.Any("err", err),
		)
		WriteProblem(w, apperr.Internal(err))

		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_, _ = w.Write(body)
}

// WriteProblem serialises an application error as problem+json. It always
// produces a body, so it is safe to call from a panic handler.
func WriteProblem(w http.ResponseWriter, appError *apperr.Error) {
	status := appError.Status
	if status == 0 {
		status = http.StatusInternalServerError
	}

	body, err := encode(ProblemDetails{
		Status:  status,
		Code:    appError.Code,
		Message: appError.Message,
		Errors:  appError.Fields,
	})
	if err != nil {
		body = []byte(`{"status":500,"code":"INTERNAL","message":"Something went wrong"}`)
		status = http.StatusInternalServerError
	}

	w.Header().Set("Content-Type", "application/problem+json")
	w.WriteHeader(status)
	_, _ = w.Write(body)
}

// encode buffers the body so an encoding failure happens before the status line
// is committed.
func encode(data any) ([]byte, error) {
	var buf bytes.Buffer

	encoder := json.NewEncoder(&buf)
	if err := encoder.Encode(data); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func asAppError(err error) *apperr.Error {
	var appError *apperr.Error
	if errors.As(err, &appError) {
		return appError
	}

	return apperr.Internal(err)
}
