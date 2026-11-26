// Package apierror maps domain errors to the stable API error shape.
package apierror

import (
	"errors"
	"fmt"
	"net/http"
)

// APIError is implemented by domain errors that carry a stable API code.
// Modules implement it on their ErrXxx values; the platform never imports a
// module in return.
type APIError interface {
	error
	Code() string
	StatusCode() int
	DeveloperMessage() string
}

// Error is the mapping result for a handler. The generated schema type in
// internal/platform/http/gen is what the API actually serialises.
type Error struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Status  int    `json:"-"`
}

func (e Error) Error() string { return fmt.Sprintf("%s: %s", e.Code, e.Message) }

// Map returns the stable code, English developer message and HTTP status for
// err. Unknown errors become INTERNAL_ERROR/500 without leaking detail.
func Map(err error) Error {
	var apiErr APIError
	if errors.As(err, &apiErr) {
		return Error{
			Code:    apiErr.Code(),
			Message: apiErr.DeveloperMessage(),
			Status:  apiErr.StatusCode(),
		}
	}
	return Error{
		Code:    "INTERNAL_ERROR",
		Message: "internal error",
		Status:  http.StatusInternalServerError,
	}
}
