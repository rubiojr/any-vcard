// Package middleware provides HTTP middleware components for the anytype client
package middleware

import (
	"context"
	"net/http"
)

// Validator is an interface for types that can validate themselves
type Validator interface {
	Validate() error
}

// ValidationMiddleware validates requests before they are sent
type ValidationMiddleware struct {
	Next HTTPDoer
}

// NewValidationMiddleware creates a new validation middleware
func NewValidationMiddleware(next HTTPDoer) *ValidationMiddleware {
	return &ValidationMiddleware{
		Next: next,
	}
}

// Do executes an HTTP request after validating any validator in the context
func (m *ValidationMiddleware) Do(req *http.Request) (*http.Response, error) {
	// Get the validator from the context
	if validator, ok := GetValidator(req.Context()); ok {
		if err := validator.Validate(); err != nil {
			return nil, err
		}
	}

	// Forward the request to the next handler
	return m.Next.Do(req)
}

type contextKey string

const validatorKey contextKey = "validator"

// WithValidator adds a validator to the context
func WithValidator(ctx context.Context, validator Validator) context.Context {
	return context.WithValue(ctx, validatorKey, validator)
}

// GetValidator retrieves a validator from the context
func GetValidator(ctx context.Context) (Validator, bool) {
	validator, ok := ctx.Value(validatorKey).(Validator)
	return validator, ok
}
