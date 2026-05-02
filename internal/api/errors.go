// Package api provides HTTP handler primitives for the defuse panel.
//
// Handlers in this package return errors instead of writing them directly.
// Errors of type *HTTPError carry an explicit status code; any other error
// is treated as a 500 and logged with full context. This keeps handlers
// short and centralizes the response shape for failures.
package api

import (
	"errors"
	"fmt"
	"net/http"
)

// HTTPError represents a response with an explicit HTTP status code.
// Handlers return *HTTPError when they want to surface a specific status
// to the client (e.g., 404 not found, 400 bad input).
type HTTPError struct {
	Status  int
	Message string
	// Cause is the underlying error, if any. It is logged but never
	// included in the response body.
	Cause error
}

// Error implements the error interface.
func (e *HTTPError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("%d %s: %v", e.Status, e.Message, e.Cause)
	}
	return fmt.Sprintf("%d %s", e.Status, e.Message)
}

// Unwrap allows errors.Is and errors.As to traverse to the underlying cause.
func (e *HTTPError) Unwrap() error {
	return e.Cause
}

// BadRequest returns a 400 error with the given user-facing message.
func BadRequest(msg string) *HTTPError {
	return &HTTPError{Status: http.StatusBadRequest, Message: msg}
}

// Unauthorized returns a 401 error.
func Unauthorized(msg string) *HTTPError {
	return &HTTPError{Status: http.StatusUnauthorized, Message: msg}
}

// Forbidden returns a 403 error.
func Forbidden(msg string) *HTTPError {
	return &HTTPError{Status: http.StatusForbidden, Message: msg}
}

// NotFound returns a 404 error.
func NotFound(msg string) *HTTPError {
	return &HTTPError{Status: http.StatusNotFound, Message: msg}
}

// Conflict returns a 409 error.
func Conflict(msg string) *HTTPError {
	return &HTTPError{Status: http.StatusConflict, Message: msg}
}

// TooManyRequests returns a 429 error.
func TooManyRequests(msg string) *HTTPError {
	return &HTTPError{Status: http.StatusTooManyRequests, Message: msg}
}

// WrapHTTP attaches an HTTP status to an existing error. The original
// error is preserved as the Cause and is reachable via errors.Is /
// errors.As, but is not exposed to the client.
func WrapHTTP(err error, status int, msg string) *HTTPError {
	return &HTTPError{Status: status, Message: msg, Cause: err}
}

// AsHTTPError returns the *HTTPError wrapped inside err, or nil if there
// is none. Useful for callers that want to inspect the status of an error
// returned by a handler-internal function.
func AsHTTPError(err error) *HTTPError {
	var he *HTTPError
	if errors.As(err, &he) {
		return he
	}
	return nil
}
