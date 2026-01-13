package tests

import (
	"testing"

	"github.com/epheo/anytype-go"
)

// TestListsAndViews tests list and view-related operations
func TestListsAndViews(t *testing.T) {
	tc := setupTestClient(t)
	defer cleanupTestClient(tc)

	spaceID := findOrCreateTestSpace(t, tc)

	// Create a test object to add to a list
	testObject, err := tc.Client.Space(spaceID).Objects().Create(tc.Ctx, anytype.CreateObjectRequest{
		TypeKey: "page",
		Name:    "List Test Object",
	})
	if err != nil {
		t.Fatalf("Failed to create test object: %v", err)
	}

	testObjectID := testObject.Object.ID
	t.Logf("Created test object with ID: %s", testObjectID)

	// Try to find a collection/list in the space
	searchResp, err := tc.Client.Space(spaceID).Search(tc.Ctx, anytype.SearchRequest{
		Types: []string{"collection"},
	})
	if err != nil {
		t.Fatalf("Failed to search for collections: %v", err)
	}

	// Find a suitable list object
	var listID string
	for _, obj := range searchResp.Data {
		if obj.Layout == "collection" || obj.Layout == "list" {
			listID = obj.ID
			break
		}
	}

	// If no collection was found, tests are conditional
	if listID == "" {
		t.Log("No suitable list/collection objects found for testing")
		t.Log("Skipping list/view operations tests")

		// Clean up the test object
		_, err = tc.Client.Space(spaceID).Object(testObjectID).Delete(tc.Ctx)
		if err != nil {
			t.Logf("Failed to clean up test object: %v", err)
		}

		return
	}

	t.Logf("Found list with ID: %s", listID)

	// List views for the collection
	viewsResp, err := tc.Client.Space(spaceID).List(listID).Views().List(tc.Ctx)
	if err != nil {
		t.Fatalf("Failed to get list views: %v", err)
	}

	t.Logf("Found %d views for list", len(viewsResp.Data))

	// If there's at least one view, test getting objects from that view
	if len(viewsResp.Data) > 0 {
		viewID := viewsResp.Data[0].ID
		t.Logf("Testing with view: %s", viewsResp.Data[0].Name)

		// Get objects in the view
		_, err := tc.Client.Space(spaceID).List(listID).View(viewID).Objects().List(tc.Ctx)
		if err != nil {
			t.Fatalf("Failed to get objects in list view: %v", err)
		}

		// Test adding object to list
		err = tc.Client.Space(spaceID).List(listID).Objects().Add(tc.Ctx, []string{testObjectID})
		if err != nil {
			t.Logf("Could not add object to list (this might be expected): %v", err)
		} else {
			t.Log("Successfully added object to list")

			// Test removing object from list
			err = tc.Client.Space(spaceID).List(listID).Object(testObjectID).Remove(tc.Ctx)
			if err != nil {
				t.Logf("Could not remove object from list: %v", err)
			} else {
				t.Log("Successfully removed object from list")
			}
		}
	}

	// Clean up: Delete the test object
	_, err = tc.Client.Space(spaceID).Object(testObjectID).Delete(tc.Ctx)
	if err != nil {
		t.Fatalf("Failed to delete test object: %v", err)
	}
}

// TestMembers tests member-related operations
func TestMembers(t *testing.T) {
	tc := setupTestClient(t)
	defer cleanupTestClient(tc)

	spaceID := findOrCreateTestSpace(t, tc)

	// List members
	membersResp, err := tc.Client.Space(spaceID).Members().List(tc.Ctx)
	if err != nil {
		t.Fatalf("Failed to list members: %v", err)
	}

	if len(membersResp.Data) == 0 {
		t.Log("No members found in space, this is unexpected")
		return
	}

	t.Logf("Found %d members in space", len(membersResp.Data))

	// Get details for the first member
	memberID := membersResp.Data[0].ID
	memberDetails, err := tc.Client.Space(spaceID).Member(memberID).Get(tc.Ctx)
	if err != nil {
		t.Fatalf("Failed to get member details: %v", err)
	}

	if memberDetails.Member.ID != memberID {
		t.Errorf("Member ID mismatch: got %s, want %s", memberDetails.Member.ID, memberID)
	}

	if memberDetails.Member.Name == "" {
		t.Error("Member name should not be empty")
	}

	if memberDetails.Member.Role == "" {
		t.Error("Member role should not be empty")
	}

	if memberDetails.Member.Status == "" {
		t.Error("Member status should not be empty")
	}
}
