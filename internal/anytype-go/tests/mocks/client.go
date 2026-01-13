package mocks

import (
	"context"

	"github.com/epheo/anytype-go"
)

// MockClient implements the anytype.Client interface for testing
type MockClient struct {
	AuthService      *MockAuthService
	SpacesService    *MockSpacesService
	SpaceService     *MockSpaceService
	mockSearchClient *MockSearchClient
}

// NewMockClient creates a new instance of MockClient with default mock services
func NewMockClient() *MockClient {
	spacesService := NewMockSpacesService()
	spaceService := NewMockSpaceService()
	authService := NewMockAuthService()

	return &MockClient{
		AuthService:      authService,
		SpacesService:    spacesService,
		SpaceService:     spaceService,
		mockSearchClient: NewMockSearchClient(),
	}
}

// Auth returns the mock auth service
func (c *MockClient) Auth() anytype.AuthClient {
	return c.AuthService
}

// Spaces returns the mock spaces service
func (c *MockClient) Spaces() anytype.SpaceClient {
	return c.SpacesService
}

// Space returns the mock space service for a specific space
func (c *MockClient) Space(spaceID string) anytype.SpaceContext {
	c.SpaceService.CurrentSpaceID = spaceID
	return c.SpaceService
}

// SearchClient returns the mock search client
func (c *MockClient) SearchClient() anytype.SearchClient {
	return c.mockSearchClient
}

// Version returns a mock version
func (c *MockClient) Version(ctx context.Context) (anytype.VersionInfo, error) {
	return anytype.VersionInfo{
		Version:    "v1.0.0-mock",
		APIVersion: "v1.0",
	}, nil
}

// Search returns a client for performing global search operations
func (c *MockClient) Search() anytype.SearchClient {
	return c.mockSearchClient
}
