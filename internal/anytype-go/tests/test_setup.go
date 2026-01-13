package tests

import (
	"context"
	"testing"
	"time"

	"github.com/epheo/anytype-go"
	"github.com/epheo/anytype-go/tests/mocks"
)

// TestClient represents a test client with authentication details
type TestClient struct {
	Client  anytype.Client
	SpaceID string
	Ctx     context.Context
	Cancel  context.CancelFunc
}

// setupTestClient creates and configures a test client with mock implementations
func setupTestClient(t *testing.T) *TestClient {
	t.Helper()

	// Create a context with a timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)

	// Create a mock client
	client := mocks.NewMockClient()

	// Set a default space ID
	spaceID := "mock-space-id"

	return &TestClient{
		Client:  client,
		SpaceID: spaceID,
		Ctx:     ctx,
		Cancel:  cancel,
	}
}

// skipIfNotAuthenticated is no longer needed with mocks,
// but kept (as a no-op) for backward compatibility
func skipIfNotAuthenticated(t *testing.T, client anytype.Client) {
	t.Helper()
	// With mock client, authentication is always available
}

// findOrCreateTestSpace returns the mock space ID
func findOrCreateTestSpace(t *testing.T, tc *TestClient) string {
	t.Helper()

	// With mock client, we always have a space available
	return tc.SpaceID
}

// cleanupTestClient performs cleanup operations
func cleanupTestClient(tc *TestClient) {
	if tc.Cancel != nil {
		tc.Cancel()
	}

	// No other cleanup needed for mock client
}
