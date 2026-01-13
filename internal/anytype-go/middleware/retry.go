package middleware

import (
	"bytes"
	"io"
	"net/http"
	"time"
)

// RetryConfig configures the retry middleware
type RetryConfig struct {
	// MaxRetries is the maximum number of retries
	MaxRetries int
	// RetryDelay is the base delay between retries
	RetryDelay time.Duration
	// MaxRetryDelay is the maximum delay between retries
	MaxRetryDelay time.Duration
	// RetryableStatusCodes defines HTTP status codes that should trigger a retry
	RetryableStatusCodes []int
	// ShouldRetry is a custom function to determine if a request should be retried
	ShouldRetry func(*http.Response, error) bool
}

// DefaultRetryConfig provides sensible defaults for retry configuration
func DefaultRetryConfig() RetryConfig {
	return RetryConfig{
		MaxRetries:    5,
		RetryDelay:    200 * time.Millisecond,
		MaxRetryDelay: 30 * time.Second,
		ShouldRetry:   defaultShouldRetry,
		RetryableStatusCodes: []int{
			http.StatusRequestTimeout,
			http.StatusTooManyRequests,
			http.StatusInternalServerError,
			http.StatusBadGateway,
			http.StatusServiceUnavailable,
			http.StatusGatewayTimeout,
		},
	}
}

// RetryMiddleware implements a retry strategy for HTTP requests
type RetryMiddleware struct {
	Next   HTTPDoer
	Config RetryConfig
}

// NewRetryMiddleware creates a new retry middleware with the given configuration
func NewRetryMiddleware(next HTTPDoer, config RetryConfig) *RetryMiddleware {
	return &RetryMiddleware{
		Next:   next,
		Config: config,
	}
}

// Do executes an HTTP request with retries according to the retry policy
func (m *RetryMiddleware) Do(req *http.Request) (*http.Response, error) {
	var resp *http.Response
	var err error

	// Keep a reference to the original request body
	var bodyCloner bodyCloner
	if req.Body != nil {
		bodyCloner, err = newBodyCloner(req)
		if err != nil {
			return nil, err
		}
	}

	// Try the initial request
	resp, err = m.Next.Do(req)

	// Retry loop
	retries := 0
	for retries < m.Config.MaxRetries && m.shouldRetry(resp, err) {
		// Delay before retry with exponential backoff
		delay := exponentialBackoff(m.Config.RetryDelay, retries, m.Config.MaxRetryDelay)
		select {
		case <-req.Context().Done():
			// Context was canceled
			return resp, req.Context().Err()
		case <-time.After(delay):
			// Continue with retry
		}

		// Clone the request to ensure it's fresh for retry
		retryReq, cloneErr := cloneRequest(req, bodyCloner)
		if cloneErr != nil {
			return resp, cloneErr
		}

		// Execute the retry
		resp, err = m.Next.Do(retryReq)
		retries++
	}

	return resp, err
}

// shouldRetry determines if a request should be retried based on the response and error
func (m *RetryMiddleware) shouldRetry(resp *http.Response, err error) bool {
	if m.Config.ShouldRetry != nil {
		return m.Config.ShouldRetry(resp, err)
	}
	return defaultShouldRetry(resp, err)
}

// defaultShouldRetry provides default retry logic
func defaultShouldRetry(resp *http.Response, err error) bool {
	// Retry on connection errors
	if err != nil {
		return true
	}

	// Retry on specific status codes
	if resp != nil {
		switch resp.StatusCode {
		case http.StatusRequestTimeout,
			http.StatusTooManyRequests,
			http.StatusInternalServerError,
			http.StatusBadGateway,
			http.StatusServiceUnavailable,
			http.StatusGatewayTimeout:
			return true
		}
	}

	return false
}

// WithRetry returns a middleware function that applies retry logic with default configuration
func WithRetry() func(HTTPDoer) HTTPDoer {
	return WithCustomRetry(DefaultRetryConfig())
}

// WithCustomRetry returns a middleware function that applies retry logic with custom configuration
func WithCustomRetry(config RetryConfig) func(HTTPDoer) HTTPDoer {
	return func(next HTTPDoer) HTTPDoer {
		return NewRetryMiddleware(next, config)
	}
}

// exponentialBackoff calculates exponential backoff delay
func exponentialBackoff(baseDelay time.Duration, retry int, maxDelay time.Duration) time.Duration {
	delay := baseDelay * (1 << uint(retry))
	if delay > maxDelay {
		delay = maxDelay
	}
	return delay
}

// Helper types and functions for request cloning and body preservation

type bodyCloner interface {
	cloneBody() (io.ReadCloser, error)
}

type readCloserCloner struct {
	buf *bytes.Buffer
}

func newBodyCloner(req *http.Request) (bodyCloner, error) {
	if req.Body == nil {
		return nil, nil
	}

	buf := new(bytes.Buffer)
	if _, err := buf.ReadFrom(req.Body); err != nil {
		return nil, err
	}
	if err := req.Body.Close(); err != nil {
		return nil, err
	}

	// Replace the original body with a new ReadCloser
	req.Body = io.NopCloser(bytes.NewReader(buf.Bytes()))

	return &readCloserCloner{buf: buf}, nil
}

func (c *readCloserCloner) cloneBody() (io.ReadCloser, error) {
	return io.NopCloser(bytes.NewReader(c.buf.Bytes())), nil
}

// cloneRequest creates a clone of the request with a fresh body
func cloneRequest(req *http.Request, bodyCloner bodyCloner) (*http.Request, error) {
	clone := req.Clone(req.Context())

	if bodyCloner != nil {
		var err error
		clone.Body, err = bodyCloner.cloneBody()
		if err != nil {
			return nil, err
		}
	}

	return clone, nil
}
