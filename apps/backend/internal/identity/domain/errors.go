// Package domain holds the identity entities and the rules that need no
// database and no HTTP.
package domain

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// Error is a domain error that carries the stable API code, its HTTP status,
// and optional machine-readable detail (platform/apierror.APIError). Errors are
// values: callers compare with errors.Is/errors.As, never on message text.
type Error struct {
	// code is the stable UPPER_SNAKE code the frontend maps to Indonesian text.
	// It is unexported so the exported accessor Code() can satisfy
	// platform/apierror.APIError without a field/method name clash.
	code string
	// Status is the HTTP status the handler returns.
	Status int
	// Developer is the English message for developers; never user-facing text.
	Developer string
	// detail is optional machine-readable detail, e.g. the offending field.
	detail map[string]any
	// Err is the wrapped cause, kept for logs and errors.Is.
	Err error
}

func (e *Error) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %s: %v", e.code, e.Developer, e.Err)
	}
	return fmt.Sprintf("%s: %s", e.code, e.Developer)
}

// Unwrap exposes the cause so errors.Is/errors.As keep working.
func (e *Error) Unwrap() error { return e.Err }

// Code satisfies platform/apierror.APIError.
func (e *Error) Code() string { return e.code }

// StatusCode satisfies platform/apierror.APIError.
func (e *Error) StatusCode() int { return e.Status }

// DeveloperMessage satisfies platform/apierror.APIError.
func (e *Error) DeveloperMessage() string { return e.Developer }

// Detail returns the machine-readable detail, or nil.
func (e *Error) Detail() map[string]any { return e.detail }

// WithDetail returns a copy carrying detail, so the shared sentinel values are
// never mutated.
func (e *Error) WithDetail(key string, value any) *Error {
	detail := make(map[string]any, len(e.detail)+1)
	for k, v := range e.detail {
		detail[k] = v
	}
	detail[key] = value
	copyOfValue := *e
	copyOfValue.detail = detail
	return &copyOfValue
}

// JSONDetail renders the detail for the API's details field, or nil.
func (e *Error) JSONDetail() json.RawMessage {
	if len(e.detail) == 0 {
		return nil
	}
	raw, err := json.Marshal(e.detail)
	if err != nil {
		return nil
	}
	return raw
}

// newError builds an Error from a code, status, message and cause.
func newError(code string, status int, developer string, cause error) *Error {
	return &Error{code: code, Status: status, Developer: developer, Err: cause}
}

// ErrInvalidCredentials is returned for a wrong username or password. It is
// deliberately the same error for both, so the response cannot be used to probe
// which usernames exist.
var ErrInvalidCredentials = newError("INVALID_CREDENTIALS", http.StatusUnauthorized, "invalid username or password", nil)

// ErrSessionExpired is returned when the session cookie is missing, unknown,
// revoked or past its expiry.
var ErrSessionExpired = newError("SESSION_EXPIRED", http.StatusUnauthorized, "session is missing or expired", nil)

// ErrPermissionDenied is returned when a signed-in user lacks the permission
// the operation requires (x-permission in api/openapi.yaml).
var ErrPermissionDenied = newError("PERMISSION_DENIED", http.StatusForbidden, "permission denied", nil)

// ErrUserNotFound is returned when the user does not exist.
var ErrUserNotFound = newError("USER_NOT_FOUND", http.StatusNotFound, "user not found", nil)

// ErrAccessGroupNotFound is returned when the access group does not exist.
var ErrAccessGroupNotFound = newError("ACCESS_GROUP_NOT_FOUND", http.StatusNotFound, "access group not found", nil)

// ErrUsernameTaken is returned when the username already exists. Usernames are
// unique across the installation.
var ErrUsernameTaken = newError("USERNAME_TAKEN", http.StatusConflict, "username already exists", nil)

// ErrGroupNameTaken is returned when the access group name already exists.
var ErrGroupNameTaken = newError("ACCESS_GROUP_NAME_TAKEN", http.StatusConflict, "access group name already exists", nil)

// ErrUserInactive is returned when a deactivated user tries to sign in.
var ErrUserInactive = newError("USER_INACTIVE", http.StatusForbidden, "user is inactive", nil)

// ErrBuiltinGroupImmutable is returned when a request tries to rename,
// deactivate, or strip permissions from the builtin Administrator group.
var ErrBuiltinGroupImmutable = newError("BUILTIN_GROUP_IMMUTABLE", http.StatusConflict, "the builtin Administrator group cannot be changed", nil)

// ErrUnknownPermission is returned when a request grants a permission that is
// not in the registry (api/openapi.yaml).
var ErrUnknownPermission = newError("UNKNOWN_PERMISSION", http.StatusUnprocessableEntity, "permission is not in the registry", nil)

// ErrInternal wraps an unexpected failure. The handler maps it to
// INTERNAL_ERROR without exposing the cause.
var ErrInternal = newError("INTERNAL_ERROR", http.StatusInternalServerError, "internal error", nil)

// ErrValidation returns a validation failure for one field.
func ErrValidation(field, reason string) *Error {
	e := newError("VALIDATION_FAILED", http.StatusUnprocessableEntity, "validation failed", nil)
	e.detail = map[string]any{"field": field, "reason": reason}
	return e
}
