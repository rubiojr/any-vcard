package e2e

import (
	"context"
	"fmt"
	"math/rand"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/epheo/anytype-go"
	vcardimport "github.com/rubiojr/any-vcard/cmd/any-vcard/import"
	"github.com/rubiojr/any-vcard/cmd/any-vcard/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	defaultAPIURL = "http://localhost:31009"
)

// TestEnv holds the test environment configuration
type TestEnv struct {
	Client  anytype.Client
	SpaceID string
	APIURL  string
	AppKey  string
}

// SetupTestSpace creates a temporary test space with a unique name
func SetupTestSpace(t *testing.T) *TestEnv {
	t.Helper()

	appKey := os.Getenv("ANYTYPE_APP_KEY")
	require.NotEmpty(t, appKey, "ANYTYPE_APP_KEY environment variable must be set")

	apiURL := os.Getenv("ANYTYPE_URL")
	if apiURL == "" {
		apiURL = defaultAPIURL
	}

	client := util.NewClientWithAppKey(apiURL, appKey)
	ctx := context.Background()

	// Generate unique space name
	spaceName := fmt.Sprintf("anyvcard_test_%d", rand.Intn(100000))

	// Create the test space
	resp, err := client.Spaces().Create(ctx, anytype.CreateSpaceRequest{
		Name: spaceName,
	})
	require.NoError(t, err, "Failed to create test space")
	require.NotEmpty(t, resp.Space.ID, "Created space should have an ID")

	t.Logf("Created test space: %s (ID: %s)", spaceName, resp.Space.ID)

	return &TestEnv{
		Client:  client,
		SpaceID: resp.Space.ID,
		APIURL:  apiURL,
		AppKey:  appKey,
	}
}

// TestImportVCard tests the end-to-end import of a vCard file
func TestImportVCard(t *testing.T) {
	env := SetupTestSpace(t)
	ctx := context.Background()

	// Create contact type in the test space
	typeResp, err := util.CreateContactType(ctx, env.Client, env.SpaceID)
	require.NoError(t, err, "Failed to create Contact type")
	t.Logf("Created Contact type with key: %s", typeResp.Type.Key)

	// Ensure properties exist
	phoneKeys, emailKeys, err := util.EnsureContactProperties(ctx, env.Client, env.SpaceID)
	require.NoError(t, err, "Failed to ensure contact properties")
	t.Logf("Phone keys: %v, Email keys: %v", phoneKeys, emailKeys)

	// Parse the sample vCard file
	vcardPath := "../examples/sample-contacts.vcf"
	contacts, err := vcardimport.ParseVCardFile(vcardPath)
	require.NoError(t, err, "Failed to parse vCard file")
	require.Len(t, contacts, 5, "Expected 5 contacts in sample file")

	// Import each contact
	for _, contact := range contacts {
		err := vcardimport.ImportContact(ctx, env.Client, env.SpaceID, typeResp.Type.Key, phoneKeys, emailKeys, contact)
		require.NoError(t, err, "Failed to import contact: %s", contact.FormattedName)
		t.Logf("Imported contact: %s", contact.FormattedName)
	}

	// Wait for indexing
	time.Sleep(2 * time.Second)

	// Verify all contacts were created
	searchResp, err := env.Client.Space(env.SpaceID).Search(ctx, anytype.SearchRequest{
		Types: []string{typeResp.Type.Key},
	})
	require.NoError(t, err, "Failed to search for contacts")
	assert.Len(t, searchResp.Data, 5, "Expected 5 contacts to be created")

	// Verify each contact's details
	verifyContact(t, searchResp.Data, "John Doe", map[string]string{
		"organization": "Acme Corporation",
		"title":        "Senior Developer",
	}, []string{"john.doe@example.com", "jdoe@work.com"}, []string{"+1-555-123-4567", "+1-555-987-6543"})

	verifyContact(t, searchResp.Data, "Jane Smith", map[string]string{
		"organization": "Tech Innovations Inc.",
		"title":        "Product Manager",
	}, []string{"jane.smith@example.com"}, []string{"+1-555-234-5678"})

	verifyContact(t, searchResp.Data, "Bob Johnson", map[string]string{},
		[]string{"bob.johnson@example.com"}, []string{"+1-555-345-6789", "+1-555-111-2222", "+1-555-333-4444"})

	verifyContact(t, searchResp.Data, "Dr. Sarah Williams", map[string]string{
		"organization": "Boston University",
		"title":        "Professor of Computer Science",
	}, []string{"s.williams@university.edu", "sarah@personal.com"}, []string{"+1-555-456-7890", "+1-555-567-8901"})

	verifyContact(t, searchResp.Data, "Carlos Rodriguez", map[string]string{
		"organization": "NextGen Startup",
		"title":        "CTO",
	}, []string{"carlos.rodriguez@startup.io"}, []string{"+1-555-678-9012"})

	t.Logf("All contacts verified successfully in space: %s", env.SpaceID)
}

// verifyContact checks that a contact exists with the expected properties
func verifyContact(t *testing.T, objects []anytype.Object, name string, expectedProps map[string]string, expectedEmails, expectedPhones []string) {
	t.Helper()

	var found *anytype.Object
	for i := range objects {
		if objects[i].Name == name {
			found = &objects[i]
			break
		}
	}
	require.NotNil(t, found, "Contact %q not found", name)

	// Check text properties
	for key, expectedValue := range expectedProps {
		propValue := getPropertyText(found.Properties, key)
		assert.Equal(t, expectedValue, propValue, "Property %q mismatch for contact %q", key, name)
	}

	// Check emails
	foundEmails := getPropertyEmails(found.Properties)
	for _, expectedEmail := range expectedEmails {
		assert.Contains(t, foundEmails, expectedEmail, "Email %q not found for contact %q", expectedEmail, name)
	}

	// Check phones
	foundPhones := getPropertyPhones(found.Properties)
	for _, expectedPhone := range expectedPhones {
		phoneFound := false
		for _, p := range foundPhones {
			if normalizePhone(p) == normalizePhone(expectedPhone) {
				phoneFound = true
				break
			}
		}
		assert.True(t, phoneFound, "Phone %q not found for contact %q (found: %v)", expectedPhone, name, foundPhones)
	}
}

// getPropertyText returns the text value of a property by key
func getPropertyText(props []anytype.Property, key string) string {
	for _, p := range props {
		if p.Key == key {
			return p.Text
		}
	}
	return ""
}

// getPropertyEmails returns all email values from properties
func getPropertyEmails(props []anytype.Property) []string {
	var emails []string
	for _, p := range props {
		if p.Email != "" {
			emails = append(emails, p.Email)
		}
	}
	return emails
}

// getPropertyPhones returns all phone values from properties
func getPropertyPhones(props []anytype.Property) []string {
	var phones []string
	for _, p := range props {
		if p.Phone != "" {
			phones = append(phones, p.Phone)
		}
	}
	return phones
}

// normalizePhone removes common phone formatting characters for comparison
func normalizePhone(phone string) string {
	phone = strings.ReplaceAll(phone, " ", "")
	phone = strings.ReplaceAll(phone, "-", "")
	phone = strings.ReplaceAll(phone, "(", "")
	phone = strings.ReplaceAll(phone, ")", "")
	return phone
}
