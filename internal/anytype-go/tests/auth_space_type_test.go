package tests

import (
	"testing"

	"github.com/epheo/anytype-go"
)

// TestAuthentication tests the authentication flow
// This test is a bit special as it requires user interaction, so it's skipped by default
func TestAuthentication(t *testing.T) {
	// Skip this test by default as it requires user interaction
	t.Skip("This test requires user interaction and is skipped by default")

	tc := setupTestClient(t)
	defer cleanupTestClient(tc)

	// Create an unauthenticated client
	client := anytype.NewClient(
		anytype.WithBaseURL("http://localhost:31009"),
	)

	// Initiate authentication flow
	authResponse, err := client.Auth().DisplayCode(tc.Ctx, "GoSDKTest")
	if err != nil {
		t.Fatalf("Failed to initiate authentication: %v", err)
	}

	if authResponse.ChallengeID == "" {
		t.Error("Challenge ID should not be empty")
	}

	// Note: We can't test GetToken without user interaction
	t.Log("Authentication challenge initiated successfully, challenge ID:", authResponse.ChallengeID)
	t.Log("To complete the test, a user would need to enter the verification code from the app")
}

// TestSpaces tests space-related operations
func TestSpaces(t *testing.T) {
	tc := setupTestClient(t)
	defer cleanupTestClient(tc)

	skipIfNotAuthenticated(t, tc.Client)

	// List spaces
	spacesResp, err := tc.Client.Spaces().List(tc.Ctx)
	if err != nil {
		t.Fatalf("Failed to list spaces: %v", err)
	}

	// If no spaces exist, create one
	var spaceID string
	if len(spacesResp.Data) == 0 {
		newSpace, err := tc.Client.Spaces().Create(tc.Ctx, anytype.CreateSpaceRequest{
			Name:        "Test Space",
			Description: "A space created for testing",
		})
		if err != nil {
			t.Fatalf("Failed to create space: %v", err)
		}
		spaceID = newSpace.Space.ID
		t.Logf("Created new space: %s (ID: %s)", newSpace.Space.Name, spaceID)
	} else {
		spaceID = spacesResp.Data[0].ID
		t.Logf("Using existing space: ID=%s", spaceID)
	}

	// Get space details
	spaceDetails, err := tc.Client.Space(spaceID).Get(tc.Ctx)
	if err != nil {
		t.Fatalf("Failed to get space details: %v", err)
	}

	if spaceDetails.Space.ID != spaceID {
		t.Errorf("Space ID mismatch: got %s, want %s", spaceDetails.Space.ID, spaceID)
	}

	if spaceDetails.Space.Name == "" {
		t.Error("Space name should not be empty")
	}
}

// TestObjectTypes tests operations with object types
func TestObjectTypes(t *testing.T) {
	tc := setupTestClient(t)
	defer cleanupTestClient(tc)

	spaceID := findOrCreateTestSpace(t, tc)

	// List types
	types, err := tc.Client.Space(spaceID).Types().List(tc.Ctx)
	if err != nil {
		t.Fatalf("Failed to list object types: %v", err)
	}

	// Verify we have some types
	if len(types) == 0 {
		t.Error("Expected to find at least one object type")
	}

	// Try to find the Page type
	var pageTypeKey string
	for _, objType := range types {
		if objType.Key == "page" {
			pageTypeKey = objType.Key
			break
		}
	}

	if pageTypeKey == "" {
		t.Log("Could not find 'page' type, this might be expected in certain configurations")
	} else {
		t.Logf("Found page type with key: %s", pageTypeKey)
	}
}
