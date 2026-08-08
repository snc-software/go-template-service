package exceptions

import "fmt"

type Exception struct {
	Code    string
	Message string
}

func (e Exception) Error() string {
	return e.Message
}

var (
	NotFound        = func(message string) Exception { return Exception{Code: "NOT_FOUND", Message: message} }
	InvalidArgument = func(message string) Exception { return Exception{Code: "INVALID_ARGUMENT", Message: message} }
	Internal        = func() Exception { return Exception{Code: "INTERNAL", Message: "Something went wrong"} }
)

func (e Exception) GetMessage() string {
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}
