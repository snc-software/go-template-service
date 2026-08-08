package httpx

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"reflect"
	"strings"

	"github.com/go-playground/validator/v10"

	"github.com/snc-software/go-template-service/internal/platform/apperr"
)

// MaxBodyBytes caps the request body a handler will read.
const MaxBodyBytes = 1 << 20

var validate = newValidator()

func newValidator() *validator.Validate {
	v := validator.New(validator.WithRequiredStructEnabled())

	v.RegisterTagNameFunc(func(field reflect.StructField) string {
		name := strings.Split(field.Tag.Get("json"), ",")[0]
		if name == "-" {
			return ""
		}

		return name
	})

	return v
}

// Decode reads a JSON request body into T and validates it, rejecting bodies
// that are oversized, malformed, contain unknown fields, or fail a constraint.
func Decode[T any](w http.ResponseWriter, r *http.Request) (T, error) {
	var target T

	r.Body = http.MaxBytesReader(w, r.Body, MaxBodyBytes)

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(&target); err != nil {
		return target, decodeError(err)
	}

	if decoder.More() {
		return target, apperr.InvalidArgument("request body must contain a single JSON object")
	}

	if err := validate.Struct(target); err != nil {
		return target, validationError(err)
	}

	return target, nil
}

func decodeError(err error) error {
	var maxBytesError *http.MaxBytesError
	var syntaxError *json.SyntaxError
	var typeError *json.UnmarshalTypeError

	switch {
	case errors.As(err, &maxBytesError):
		return apperr.TooLarge(fmt.Sprintf("request body must not exceed %d bytes", maxBytesError.Limit))

	case errors.As(err, &syntaxError):
		return apperr.InvalidArgument(fmt.Sprintf("request body contains malformed JSON at position %d", syntaxError.Offset))

	case errors.As(err, &typeError):
		return apperr.Validation("request body contains an invalid value", []apperr.FieldError{{
			Field:   typeError.Field,
			Message: fmt.Sprintf("must be of type %s", typeError.Type),
		}})

	case errors.Is(err, io.EOF):
		return apperr.InvalidArgument("request body must not be empty")

	case strings.HasPrefix(err.Error(), "json: unknown field "):
		field := strings.TrimPrefix(err.Error(), "json: unknown field ")
		return apperr.InvalidArgument(fmt.Sprintf("request body contains unknown field %s", field))

	default:
		return apperr.InvalidArgument("request body could not be read")
	}
}

func validationError(err error) error {
	var validationErrors validator.ValidationErrors
	if !errors.As(err, &validationErrors) {
		return apperr.Internal(fmt.Errorf("validate request body: %w", err))
	}

	fields := make([]apperr.FieldError, len(validationErrors))
	for i, fieldError := range validationErrors {
		fields[i] = apperr.FieldError{
			Field:   fieldError.Field(),
			Message: constraintMessage(fieldError),
		}
	}

	return apperr.Validation("request body failed validation", fields)
}

func constraintMessage(fieldError validator.FieldError) string {
	switch fieldError.Tag() {
	case "required":
		return "is required"
	case "email":
		return "must be a valid email address"
	case "min":
		return fmt.Sprintf("must be at least %s characters", fieldError.Param())
	case "max":
		return fmt.Sprintf("must be at most %s characters", fieldError.Param())
	default:
		return fmt.Sprintf("failed the %q constraint", fieldError.Tag())
	}
}
