// Package domain defines the core business types and errors for the Nucleus Security CLI.
// This package has zero external dependencies.
package domain

import "errors"

// Sentinel errors for domain-level error handling.
var (
	// ErrNotFound indicates the requested resource does not exist.
	ErrNotFound = errors.New("resource not found")

	// ErrUnauthorized indicates invalid or missing authentication credentials.
	ErrUnauthorized = errors.New("unauthorized: invalid or missing API key")

	// ErrForbidden indicates the authenticated user lacks permission.
	ErrForbidden = errors.New("forbidden: insufficient permissions")

	// ErrRateLimited indicates the API rate limit has been exceeded.
	ErrRateLimited = errors.New("rate limited: too many requests")

	// ErrCircuitOpen indicates the circuit breaker is open due to repeated failures.
	ErrCircuitOpen = errors.New("circuit breaker open: API is unavailable")

	// ErrRetryExhausted indicates all retry attempts have been exhausted.
	ErrRetryExhausted = errors.New("all retry attempts exhausted")

	// ErrValidation indicates invalid input data.
	ErrValidation = errors.New("validation error: invalid input")
)

// APIError represents a structured error from the Nucleus API.
type APIError struct {
	StatusCode int
	Message    string
	RequestID  string
}

func (e *APIError) Error() string {
	if e.RequestID != "" {
		return e.Message + " (request_id: " + e.RequestID + ")"
	}
	return e.Message
}

// IsRetryable returns true if the error is transient and the request can be retried.
func IsRetryable(err error) bool {
	if errors.Is(err, ErrRateLimited) {
		return true
	}
	var apiErr *APIError
	if errors.As(err, &apiErr) {
		switch apiErr.StatusCode {
		case 429, 500, 502, 503, 504:
			return true
		}
	}
	return false
}
