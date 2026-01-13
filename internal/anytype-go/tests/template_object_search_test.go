package tests

import (
	"testing"

	"github.com/epheo/anytype-go"
)

// TestTemplates tests template-related operations
func TestTemplates(t *testing.T) {
	tc := setupTestClient(t)
	defer cleanupTestClient(tc)

	spaceID := findOrCreateTestSpace(t, tc)

	// Find a type that we can use for template tests
	types, err := tc.Client.Space(spaceID).Types().List(tc.Ctx)
	if err != nil {
		t.Fatalf("Failed to list object types: %v", err)
	}

	// Try to find a suitable type (prefer 'page')
	var typeKey string
	for _, objType := range types {
		if objType.Key == "page" {
			typeKey = objType.Key
			break
		}
	}

	if typeKey == "" && len(types) > 0 {
		typeKey = types[0].Key // Fallback to first available type
	}

	if typeKey == "" {
		t.Skip("No types available for template testing")
	}

	// List templates for the type
	templates, err := tc.Client.Space(spaceID).Type(typeKey).Templates().List(tc.Ctx)
	if err != nil {
		t.Fatalf("Failed to list templates: %v", err)
	}

	// Template tests are conditional on having templates available
	if len(templates) == 0 {
		t.Logf("No templates found for type %s, skipping detailed template tests", typeKey)
		return
	}

	// Get template details
	templateID := templates[0].ID
	templateDetails, err := tc.Client.Space(spaceID).Type(typeKey).Template(templateID).Get(tc.Ctx)
	if err != nil {
		t.Fatalf("Failed to get template details: %v", err)
	}

	if templateDetails.Template.ID != templateID {
		t.Errorf("Template ID mismatch: got %s, want %s", templateDetails.Template.ID, templateID)
	}

	if templateDetails.Template.Name == "" {
		t.Error("Template name should not be empty")
	}
}

// TestObjects tests object-related operations
func TestObjects(t *testing.T) {
	tc := setupTestClient(t)
	defer cleanupTestClient(tc)

	spaceID := findOrCreateTestSpace(t, tc)

	// List existing objects
	objects, err := tc.Client.Space(spaceID).Objects().List(tc.Ctx)
	if err != nil {
		t.Fatalf("Failed to list objects: %v", err)
	}

	initialObjectCount := len(objects)
	t.Logf("Initial object count: %d", initialObjectCount)

	// Create a test object
	testObjectName := "Go SDK Test Object"
	createReq := anytype.CreateObjectRequest{
		TypeKey: "page",
		Name:    testObjectName,
		Body:    "# Test Object\n\nThis is a test object created by the Go SDK tests.",
	}

	// Set an emoji icon
	createReq.Icon = &anytype.Icon{
		Format: anytype.IconFormatEmoji,
		Emoji:  "🧪",
	}

	newObject, err := tc.Client.Space(spaceID).Objects().Create(tc.Ctx, createReq)
	if err != nil {
		t.Fatalf("Failed to create object: %v", err)
	}

	objectID := newObject.Object.ID
	t.Logf("Created test object with ID: %s", objectID)

	// Verify the object was created correctly
	if newObject.Object.Name != testObjectName {
		t.Errorf("Object name mismatch: got %s, want %s", newObject.Object.Name, testObjectName)
	}

	// Get object details
	objectDetails, err := tc.Client.Space(spaceID).Object(objectID).Get(tc.Ctx)
	if err != nil {
		t.Fatalf("Failed to get object details: %v", err)
	}

	if objectDetails.Object.ID != objectID {
		t.Errorf("Object ID mismatch: got %s, want %s", objectDetails.Object.ID, objectID)
	}

	// Try to export the object
	exportResp, err := tc.Client.Space(spaceID).Object(objectID).Export(tc.Ctx, "markdown")
	if err != nil {
		t.Logf("Export not available: %v", err)
	} else {
		if len(exportResp.Markdown) == 0 {
			t.Error("Exported markdown should not be empty")
		}
	}

	// Clean up: Delete the test object
	deletedObject, err := tc.Client.Space(spaceID).Object(objectID).Delete(tc.Ctx)
	if err != nil {
		t.Fatalf("Failed to delete test object: %v", err)
	}

	if !deletedObject.Object.Archived {
		t.Error("Deleted object should be marked as archived")
	}
}

// TestSearching tests search functionality
func TestSearching(t *testing.T) {
	tc := setupTestClient(t)
	defer cleanupTestClient(tc)

	spaceID := findOrCreateTestSpace(t, tc)

	// Create a test object with a unique searchable name
	uniqueSearchTerm := "UniqueTestSearchTerm2025"
	createReq := anytype.CreateObjectRequest{
		TypeKey: "page",
		Name:    uniqueSearchTerm,
		Body:    "This is a test object with a unique search term for testing search functionality",
	}

	newObject, err := tc.Client.Space(spaceID).Objects().Create(tc.Ctx, createReq)
	if err != nil {
		t.Fatalf("Failed to create test object for search: %v", err)
	}

	objectID := newObject.Object.ID
	t.Logf("Created search test object with ID: %s", objectID)

	// Perform basic search
	searchResp, err := tc.Client.Space(spaceID).Search(tc.Ctx, anytype.SearchRequest{
		Query: uniqueSearchTerm,
	})
	if err != nil {
		t.Fatalf("Failed to perform basic search: %v", err)
	}

	// Verify search results
	found := false
	for _, result := range searchResp.Data {
		if result.ID == objectID {
			found = true
			break
		}
	}

	if !found {
		t.Errorf("Search did not find the created test object with term: %s", uniqueSearchTerm)
	}

	// Advanced search with sorting and filtering
	advancedSearchResp, err := tc.Client.Space(spaceID).Search(tc.Ctx, anytype.SearchRequest{
		Query: uniqueSearchTerm,
		Sort: &anytype.SortOptions{
			Property:  anytype.SortPropertyLastModifiedDate,
			Direction: anytype.SortDirectionDesc,
		},
		Types: []string{"page"},
	})
	if err != nil {
		t.Fatalf("Failed to perform advanced search: %v", err)
	}

	// Verify advanced search results
	found = false
	for _, result := range advancedSearchResp.Data {
		if result.ID == objectID {
			found = true
			break
		}
	}

	if !found {
		t.Errorf("Advanced search did not find the created test object")
	}

	// Clean up: Delete the test object
	_, err = tc.Client.Space(spaceID).Object(objectID).Delete(tc.Ctx)
	if err != nil {
		t.Fatalf("Failed to delete search test object: %v", err)
	}
}
