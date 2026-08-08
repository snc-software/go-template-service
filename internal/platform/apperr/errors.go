// Package apperr defines the error type that crosses layer boundaries, carrying
// both a machine-readable code and the HTTP status it maps to.
package apperr

import (
	"fmt"
	"net/http"
)

// Code is the stable, machine-readable classification of an error. Messages are
// for humans and may change without notice.
type Code string

const (
	CodeNotFound        Code = "NOT_FOUND"
	CodeInvalidArgument Code = "INVALID_ARGUMENT"
	CodeValidation      Code = "VALIDATION_FAILED"
	CodeConflict        Code = "CONFLICT"
	CodeTooLarge        Code = "REQUEST_TOO_LARGE"
	CodeInternal        Code = "INTERNAL"
)

// FieldError describes one failed constraint on one request field.
type FieldError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

// Error is the application error type. The wrapped cause is logged server-side
// and never serialised into the response body.
type Error struct {
	Code    Code
	Message string
	Status  int
	Fields  []FieldError

	cause error
}

func (e *Error) Error() string {
	if e.cause != nil {
		return fmt.Sprintf("%s: %s: %v", e.Code, e.Message, e.cause)
	}

	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

func (e *Error) Unwrap() error {
	return e.cause
}

func NotFound(message string) *Error {
	return &Error{Code: CodeNotFound, Message: message, Status: http.StatusNotFound}
}

func InvalidArgument(message string) *Error {
	return &Error{Code: CodeInvalidArgument, Message: message, Status: http.StatusBadRequest}
}

func Validation(message string, fields []FieldError) *Error {
	return &Error{Code: CodeValidation, Message: message, Status: http.StatusBadRequest, Fields: fields}
}

func Conflict(message string) *Error {
	return &Error{Code: CodeConflict, Message: message, Status: http.StatusConflict}
}

func TooLarge(message string) *Error {
	return &Error{Code: CodeTooLarge, Message: message, Status: http.StatusRequestEntityTooLarge}
}

func Internal(cause error) *Error {
	return &Error{
		Code:    CodeInternal,
		Message: "Something went wrong",
		Status:  http.StatusInternalServerError,
		cause:   cause,
	}
}
