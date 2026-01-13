// Package middleware provides HTTP middleware components for the anytype client
package middleware

import (
	"net/http"
)

// HTTPDoer represents any client capable of executing HTTP requests
type HTTPDoer interface {
	Do(*http.Request) (*http.Response, error)
}

// Chain represents a middleware chain
type Chain struct {
	middlewares []func(HTTPDoer) HTTPDoer
	client      HTTPDoer
}

// NewChain creates a new middleware chain with the given client
func NewChain(client HTTPDoer) *Chain {
	return &Chain{
		client: client,
	}
}

// Use adds a middleware to the chain
func (c *Chain) Use(middleware func(HTTPDoer) HTTPDoer) *Chain {
	c.middlewares = append(c.middlewares, middleware)
	return c
}

// Build constructs the final HTTP client with all middlewares applied
func (c *Chain) Build() HTTPDoer {
	result := c.client
	// Apply middlewares in reverse order so the first middleware is the outermost
	for i := len(c.middlewares) - 1; i >= 0; i-- {
		result = c.middlewares[i](result)
	}
	return result
}
