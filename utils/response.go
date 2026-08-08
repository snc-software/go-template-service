package utils

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/snc-software/go-template-service/exceptions"
)

type ProblemDetails struct {
	Status  int    `json:"status"`
	Code    string `json:"code"`
	Message string `json:"message"`
}

func OkResponse(responseWriter http.ResponseWriter, data any) {
	WriteJsonResponse(responseWriter, data, http.StatusOK)
}

func NoContentResponse(responseWriter http.ResponseWriter) {
	WriteJsonResponse(responseWriter, nil, http.StatusNoContent)
}

func CreatedResponse(responseWriter http.ResponseWriter, data any) {
	WriteJsonResponse(responseWriter, data, http.StatusCreated)
}

func WriteJsonResponse(responseWriter http.ResponseWriter, data any, statusCode int) {
	responseWriter.Header().Set("Content-Type", "application/json")
	responseWriter.WriteHeader(statusCode)
	if data != nil {
		json.NewEncoder(responseWriter).Encode(data)
	}
}

func Error(responseWriter http.ResponseWriter, message string, statusCode int) {
	responseWriter.Header().Set("Content-Type", "application/json")
	responseWriter.WriteHeader(statusCode)
	json.NewEncoder(responseWriter).Encode(map[string]string{"error": message})
}

func HandleError(responseWriter http.ResponseWriter, err error) {
	var exception exceptions.Exception
	if errors.As(err, &exception) {
		switch exception.Code {
		case "NOT_FOUND":
			problem(responseWriter, exception.Message, exception.Code, http.StatusNotFound)
		default:
			problem(responseWriter, exception.Message, exception.Code, http.StatusInternalServerError)
		}
		return
	}

	problem(responseWriter, "internal error", "INTERNAL", http.StatusInternalServerError)
}

func problem(responseWriter http.ResponseWriter, message, code string, status int) {
	responseWriter.Header().Set("Content-Type", "application/problem+json")
	responseWriter.WriteHeader(status)
	json.NewEncoder(responseWriter).Encode(ProblemDetails{
		Status:  status,
		Code:    code,
		Message: message,
	})
}
