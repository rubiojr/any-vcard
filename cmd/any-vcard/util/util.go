package util

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/epheo/anytype-go"
	_ "github.com/epheo/anytype-go/client"
	"github.com/urfave/cli/v3"
)

const (
	ContactTypeKey = "contact"
	AppName        = "any-vcard"
	Version        = "0.1.0"
)

// RequireFlags checks that the specified flags are set
func RequireFlags(cmd *cli.Command, flags ...string) error {
	var missing []string
	for _, flag := range flags {
		if cmd.String(flag) == "" {
			missing = append(missing, flag)
		}
	}
	if len(missing) > 0 {
		return fmt.Errorf("required flags %q not set", strings.Join(missing, ", "))
	}
	return nil
}

// NewClient creates a new Anytype client from CLI flags
func NewClient(cmd *cli.Command) anytype.Client {
	return anytype.NewClient(
		anytype.WithBaseURL(cmd.String("url")),
		anytype.WithAppKey(cmd.String("app-key")),
	)
}

// NewClientWithURL creates a new Anytype client with just a URL (for auth)
func NewClientWithURL(baseURL string) anytype.Client {
	return anytype.NewClient(
		anytype.WithBaseURL(baseURL),
	)
}

// NewClientWithAppKey creates a new Anytype client with URL and app key (no space required)
func NewClientWithAppKey(baseURL, appKey string) anytype.Client {
	return anytype.NewClient(
		anytype.WithBaseURL(baseURL),
		anytype.WithAppKey(appKey),
	)
}

// EnsureContactProperties creates required properties if they don't exist
// Returns phoneKeys and emailKeys for all available phone/email properties
func EnsureContactProperties(ctx context.Context, client anytype.Client, spaceID string) ([]string, []string, error) {
	existingProps, err := client.Space(spaceID).Properties().List(ctx)
	if err != nil {
		log.Printf("Warning: could not list properties: %v", err)
		existingProps = []anytype.Property{}
	}

	var existingPhoneKeys []string
	var existingEmailKeys []string
	existingPhoneByName := make(map[string]string)
	existingEmailByName := make(map[string]string)

	for _, prop := range existingProps {
		if prop.Format == "phone" {
			existingPhoneKeys = append(existingPhoneKeys, prop.Key)
			existingPhoneByName[prop.Name] = prop.Key
		} else if prop.Format == "email" {
			existingEmailKeys = append(existingEmailKeys, prop.Key)
			existingEmailByName[prop.Name] = prop.Key
		}
	}

	phoneProps := []struct {
		Name string
		Key  string
	}{
		{"Phone", "phone"},
		{"Phone 2", "phone2"},
		{"Phone 3", "phone3"},
	}

	emailProps := []struct {
		Name string
		Key  string
	}{
		{"Email", "email"},
		{"Email 2", "email2"},
		{"Email 3", "email3"},
	}

	var phoneKeys []string
	var emailKeys []string
	var createdKeys []string

	for _, phoneProp := range phoneProps {
		if existingKey, exists := existingPhoneByName[phoneProp.Name]; exists {
			phoneKeys = append(phoneKeys, existingKey)
		} else {
			resp, err := client.Space(spaceID).Properties().Create(ctx, anytype.CreatePropertyRequest{
				Key:    phoneProp.Key,
				Name:   phoneProp.Name,
				Format: "phone",
			})
			if err != nil {
				log.Printf("Warning: could not create property %s: %v", phoneProp.Name, err)
				continue
			}
			phoneKeys = append(phoneKeys, resp.Property.Key)
			createdKeys = append(createdKeys, resp.Property.Key)
			fmt.Printf("  Created property: %s (key: %s)\n", phoneProp.Name, resp.Property.Key)
		}
	}

	for _, emailProp := range emailProps {
		if existingKey, exists := existingEmailByName[emailProp.Name]; exists {
			emailKeys = append(emailKeys, existingKey)
		} else {
			resp, err := client.Space(spaceID).Properties().Create(ctx, anytype.CreatePropertyRequest{
				Key:    emailProp.Key,
				Name:   emailProp.Name,
				Format: "email",
			})
			if err != nil {
				log.Printf("Warning: could not create property %s: %v", emailProp.Name, err)
				continue
			}
			emailKeys = append(emailKeys, resp.Property.Key)
			createdKeys = append(createdKeys, resp.Property.Key)
			fmt.Printf("  Created property: %s (key: %s)\n", emailProp.Name, resp.Property.Key)
		}
	}

	if len(createdKeys) > 0 {
		allKeys := append(phoneKeys, emailKeys...)
		if err := WaitForProperties(ctx, client, spaceID, allKeys); err != nil {
			log.Printf("Warning: %v", err)
		}
	}

	if len(phoneKeys) == 0 {
		return nil, nil, fmt.Errorf("no phone properties available")
	}
	if len(emailKeys) == 0 {
		return nil, nil, fmt.Errorf("no email properties available")
	}

	return phoneKeys, emailKeys, nil
}

// WaitForProperties polls the server until all specified property keys are available
func WaitForProperties(ctx context.Context, client anytype.Client, spaceID string, keys []string) error {
	fmt.Printf("  Waiting for properties to be available...\n")
	for i := 0; i < 20; i++ {
		props, err := client.Space(spaceID).Properties().List(ctx)
		if err != nil {
			time.Sleep(500 * time.Millisecond)
			continue
		}
		allFound := true
		for _, key := range keys {
			found := false
			for _, prop := range props {
				if prop.Key == key {
					found = true
					break
				}
			}
			if !found {
				allFound = false
				break
			}
		}
		if allFound {
			time.Sleep(2 * time.Second)
			fmt.Printf("  Properties ready.\n")
			return nil
		}
		time.Sleep(500 * time.Millisecond)
	}
	return fmt.Errorf("timeout waiting for properties to be available")
}

// CreateContactType creates the Contact object type in a space
func CreateContactType(ctx context.Context, client anytype.Client, spaceID string) (*anytype.TypeResponse, error) {
	properties := []anytype.PropertyDefinition{
		{Key: "name", Name: "Name", Format: "text"},
		{Key: "given_name", Name: "Given Name", Format: "text"},
		{Key: "family_name", Name: "Family Name", Format: "text"},
		{Key: "middle_name", Name: "Middle Name", Format: "text"},
		{Key: "prefix", Name: "Prefix", Format: "text"},
		{Key: "suffix", Name: "Suffix", Format: "text"},
		{Key: "email", Name: "Email", Format: "email"},
		{Key: "phone", Name: "Phone", Format: "phone"},
		{Key: "address", Name: "Address", Format: "text"},
		{Key: "city", Name: "City", Format: "text"},
		{Key: "region", Name: "Region", Format: "text"},
		{Key: "postal_code", Name: "Postal Code", Format: "text"},
		{Key: "country", Name: "Country", Format: "text"},
		{Key: "organization", Name: "Organization", Format: "text"},
		{Key: "title", Name: "Title", Format: "text"},
		{Key: "url", Name: "URL", Format: "url"},
		{Key: "birthday", Name: "Birthday", Format: "date"},
		{Key: "notes", Name: "Notes", Format: "text"},
	}

	req := anytype.CreateTypeRequest{
		Key:        "contact",
		Name:       "Contact",
		Layout:     "basic",
		PluralName: "Contacts",
		Icon: &anytype.Icon{
			Format: anytype.IconFormatEmoji,
			Emoji:  "👤",
		},
		Properties: properties,
	}

	return client.Space(spaceID).Types().Create(ctx, req)
}

// GlobalFlags returns the common flags used by most commands
func GlobalFlags() []cli.Flag {
	return []cli.Flag{
		&cli.StringFlag{
			Name:    "url",
			Aliases: []string{"u"},
			Value:   "http://localhost:31009",
			Usage:   "Anytype API URL",
		},
		&cli.StringFlag{
			Name:    "app-key",
			Aliases: []string{"k"},
			Usage:   "Anytype App Key",
			Sources: cli.EnvVars("ANYTYPE_APP_KEY"),
		},
		&cli.StringFlag{
			Name:    "space",
			Aliases: []string{"s"},
			Usage:   "Space ID to import contacts into",
			Sources: cli.EnvVars("ANYTYPE_SPACE_ID"),
		},
	}
}
