package mocks

import (
	"context"
)

// MockListClient implements the anytype.ListClient interface for testing
type MockListClient struct {
	AddFunc func(ctx context.Context, objectIDs []string) error
}

// NewMockListClient creates a new instance of MockListClient with default implementations
func NewMockListClient() *MockListClient {
	return &MockListClient{
		AddFunc: func(ctx context.Context, objectIDs []string) error {
			return nil
		},
	}
}

// Add calls the mock implementation
func (c *MockListClient) Add(ctx context.Context, objectIDs []string) error {
	return c.AddFunc(ctx, objectIDs)
}
