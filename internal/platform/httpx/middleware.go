package httpx

import (
	"fmt"
	"log/slog"
	"net/http"
	"runtime/debug"
	"time"

	"github.com/go-chi/chi/v5/middleware"

	"github.com/snc-software/go-template-service/internal/platform/apperr"
)

// Recoverer logs a panic with its stack and converts it into a problem+json 500.
func Recoverer(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				recovered := recover()
				if recovered == nil {
					return
				}

				// http.ErrAbortHandler is a deliberate abort, not a failure.
				if recovered == http.ErrAbortHandler {
					panic(recovered)
				}

				logger.ErrorContext(r.Context(), "panic recovered",
					slog.String("method", r.Method),
					slog.String("path", r.URL.Path),
					slog.String("request_id", middleware.GetReqID(r.Context())),
					slog.Any("panic", recovered),
					slog.String("stack", string(debug.Stack())),
				)

				WriteProblem(w, apperr.Internal(fmt.Errorf("panic: %v", recovered)))
			}()

			next.ServeHTTP(w, r)
		})
	}
}

// RequestLogger emits one structured line per completed request.
func RequestLogger(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			wrapped := middleware.NewWrapResponseWriter(w, r.ProtoMajor)
			start := time.Now()

			defer func() {
				logger.InfoContext(r.Context(), "request",
					slog.String("method", r.Method),
					slog.String("path", r.URL.Path),
					slog.Int("status", wrapped.Status()),
					slog.Int("bytes", wrapped.BytesWritten()),
					slog.Int64("duration_ms", time.Since(start).Milliseconds()),
					slog.String("request_id", middleware.GetReqID(r.Context())),
				)
			}()

			next.ServeHTTP(wrapped, r)
		})
	}
}
