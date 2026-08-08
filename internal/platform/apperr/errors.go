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

// The codes the API can return. They are part of the public contract: clients
// may branch on them, so they change only with the API version.
const (
	CodeNotFound        Code = "NOT_FOUND"
	CodeInvalidArgument Code = "INVALID_ARGUMENT"
	CodeValidation      Code = "VALIDATION_FAILED"
	CodeConflict        Code = "CONFLICT"
	CodeTooLarge        Code = "REQUEST_TOO_LARGE"
	CodeUnavailable     Code = "UNAVAILABLE"
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

// Error implements the error interface.
func (e *Error) Error() string {
	if e.cause != nil {
		return fmt.Sprintf("%s: %s: %v", e.Code, e.Message, e.cause)
	}

	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

// Unwrap returns the wrapped cause, if any.
func (e *Error) Unwrap() error {
	return e.cause
}

// NotFound returns a 404 error.
func NotFound(message string) *Error {
	return &Error{Code: CodeNotFound, Message: message, Status: http.StatusNotFound}
}

// InvalidArgument returns a 400 error for a malformed or unparseable input.
func InvalidArgument(message string) *Error {
	return &Error{Code: CodeInvalidArgument, Message: message, Status: http.StatusBadRequest}
}

// Validation returns a 400 error carrying one entry per failed field constraint.
func Validation(message string, fields []FieldError) *Error {
	return &Error{Code: CodeValidation, Message: message, Status: http.StatusBadRequest, Fields: fields}
}

// Conflict returns a 409 error.
func Conflict(message string) *Error {
	return &Error{Code: CodeConflict, Message: message, Status: http.StatusConflict}
}

// TooLarge returns a 413 error.
func TooLarge(message string) *Error {
	return &Error{Code: CodeTooLarge, Message: message, Status: http.StatusRequestEntityTooLarge}
}

// Unavailable returns a 503 error wrapping the dependency failure that caused it.
func Unavailable(message string, cause error) *Error {
	return &Error{
		Code:    CodeUnavailable,
		Message: message,
		Status:  http.StatusServiceUnavailable,
		cause:   cause,
	}
}

// Internal returns a 500 error whose cause is logged and never serialised.
func Internal(cause error) *Error {
	return &Error{
		Code:    CodeInternal,
		Message: "Something went wrong",
		Status:  http.StatusInternalServerError,
		cause:   cause,
	}
}
