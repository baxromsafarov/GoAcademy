// Package apierr defines APIError: a transport-agnostic error that carries the
// information needed to render a consistent HTTP error response. Domain services
// return these so the HTTP layer can map them to status codes without importing
// every domain package.
package apierr

import (
	"fmt"
	"net/http"
)

// Stable, machine-readable error codes returned to clients.
const (
	CodeValidation   = "validation_error"
	CodeUnauthorized = "unauthorized"
	CodeForbidden    = "forbidden"
	CodeNotFound     = "not_found"
	CodeConflict     = "conflict"
	CodeRateLimited  = "rate_limited"
	CodeInternal     = "internal"
)

// APIError is an error that maps cleanly to an HTTP response.
type APIError struct {
	Status  int    // HTTP status code
	Code    string // stable machine-readable code (one of the Code* constants)
	Message string // human-readable message, safe to expose to clients
	Details any    // optional structured details (e.g. per-field validation errors)
	cause   error  // optional internal cause, logged but never exposed
}

func (e *APIError) Error() string {
	if e.cause != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.cause)
	}
	return e.Message
}

// Unwrap exposes the internal cause for errors.Is/errors.As chains.
func (e *APIError) Unwrap() error { return e.cause }

// WithCause attaches an internal cause (for logging) and returns the same error.
func (e *APIError) WithCause(err error) *APIError {
	e.cause = err
	return e
}

// WithDetails attaches structured details and returns the same error.
func (e *APIError) WithDetails(details any) *APIError {
	e.Details = details
	return e
}

// New builds an APIError with an explicit status and code.
func New(status int, code, message string) *APIError {
	return &APIError{Status: status, Code: code, Message: message}
}

// Constructors for the common cases.

func Validation(message string) *APIError {
	return New(http.StatusBadRequest, CodeValidation, message)
}

func Unauthorized(message string) *APIError {
	return New(http.StatusUnauthorized, CodeUnauthorized, message)
}

func Forbidden(message string) *APIError {
	return New(http.StatusForbidden, CodeForbidden, message)
}

func NotFound(message string) *APIError {
	return New(http.StatusNotFound, CodeNotFound, message)
}

func Conflict(message string) *APIError {
	return New(http.StatusConflict, CodeConflict, message)
}

func RateLimited(message string) *APIError {
	return New(http.StatusTooManyRequests, CodeRateLimited, message)
}

// Internal returns a generic 500 whose details are never exposed to clients.
func Internal() *APIError {
	return New(http.StatusInternalServerError, CodeInternal, "internal server error")
}
