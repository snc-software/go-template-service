package middleware

import (
	"net/http"

	"github.com/snc-software/go-template-service/utils"
)

func ErrorHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(responseWriter http.ResponseWriter, request *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				utils.Error(responseWriter, "internal server error", http.StatusInternalServerError)
			}
		}()

		next.ServeHTTP(responseWriter, request)
	})
}
