package vcard

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	govcard "github.com/emersion/go-vcard"
	"github.com/rubiojr/anytype-go"
)

// Contact represents a parsed vCard contact
type Contact struct {
	FormattedName string
	GivenName     string
	FamilyName    string
	MiddleName    string
	Prefix        string
	Suffix        string
	Emails        []string
	Phones        []string
	Addresses     []Address
	Organization  string
	Title         string
	URLs          []string
	Note          string
	Birthday      string
	Photo         string
	ObjectID      string // Anytype object ID (used for merge operations)
}

// DisplayName returns the best available name for the contact
func (c Contact) DisplayName() string {
	if c.FormattedName != "" {
		return c.FormattedName
	}
	parts := filterEmpty(c.Prefix, c.GivenName, c.MiddleName, c.FamilyName, c.Suffix)
	if len(parts) > 0 {
		return strings.Join(parts, " ")
	}
	if c.Organization != "" {
		return c.Organization
	}
	return "Unnamed Contact"
}

// Address represents a physical address
type Address struct {
	Street     string
	City       string
	Region     string
	PostalCode string
	Country    string
	Full       string
}

// filterEmpty returns only non-empty strings
func filterEmpty(strs ...string) []string {
	result := make([]string, 0, len(strs))
	for _, s := range strs {
		if s != "" {
			result = append(result, s)
		}
	}
	return result
}

// NormalizePhone removes common formatting characters for comparison
func NormalizePhone(phone string) string {
	phone = strings.ToLower(strings.TrimSpace(phone))
	return strings.NewReplacer(" ", "", "-", "", "(", "", ")", "").Replace(phone)
}

// ParseFile parses a vCard file and returns the contacts
func ParseFile(filePath string) ([]Contact, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	decoder := govcard.NewDecoder(file)
	var contacts []Contact

	for {
		card, err := decoder.Decode()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return contacts, fmt.Errorf("failed to decode vCard: %w", err)
		}
		contacts = append(contacts, parseCard(card))
	}

	return contacts, nil
}

func parseCard(card govcard.Card) Contact {
	contact := Contact{
		FormattedName: card.PreferredValue(govcard.FieldFormattedName),
		Organization:  card.PreferredValue(govcard.FieldOrganization),
		Title:         card.PreferredValue(govcard.FieldTitle),
		Note:          card.PreferredValue(govcard.FieldNote),
		Birthday:      card.PreferredValue(govcard.FieldBirthday),
		Photo:         card.PreferredValue(govcard.FieldPhoto),
	}

	if names := card.Name(); names != nil {
		contact.FamilyName = names.FamilyName
		contact.GivenName = names.GivenName
		contact.MiddleName = names.AdditionalName
		contact.Prefix = names.HonorificPrefix
		contact.Suffix = names.HonorificSuffix
	}

	contact.Emails = parseFieldValues(card, govcard.FieldEmail, "mailto:")
	contact.Phones = parseFieldValues(card, govcard.FieldTelephone, "tel:")
	contact.URLs = parseFieldValues(card, govcard.FieldURL, "")

	if addr := card.Address(); addr != nil {
		street := addr.StreetAddress
		if street == "" {
			street = addr.ExtendedAddress
		}
		contact.Addresses = append(contact.Addresses, Address{
			Street:     street,
			City:       addr.Locality,
			Region:     addr.Region,
			PostalCode: addr.PostalCode,
			Country:    addr.Country,
			Full:       street,
		})
	}

	return contact
}

// parseFieldValues extracts and cleans values from a vCard field
func parseFieldValues(card govcard.Card, field, trimPrefix string) []string {
	var result []string
	for _, val := range card.Values(field) {
		val = strings.TrimSpace(val)
		if val == "" {
			continue
		}
		if trimPrefix != "" {
			val = strings.TrimPrefix(val, trimPrefix)
		}
		result = append(result, val)
	}
	return result
}

// ParseBirthday attempts to parse birthday in common formats
func ParseBirthday(bday string) string {
	formats := []string{"20060102", "2006-01-02"}
	for _, format := range formats {
		if t, err := time.Parse(format, bday); err == nil {
			return t.Format(time.RFC3339)
		}
	}
	return bday
}

// BuildNotes constructs the notes field including overflow data
func BuildNotes(contact Contact) string {
	var notes []string
	if contact.Note != "" {
		notes = append(notes, contact.Note)
	}
	if len(contact.Emails) > 3 {
		notes = append(notes, "Additional emails: "+strings.Join(contact.Emails[3:], ", "))
	}
	if len(contact.URLs) > 1 {
		notes = append(notes, "Additional URLs: "+strings.Join(contact.URLs[1:], ", "))
	}
	return strings.Join(notes, "\n\n")
}

// Import creates an Anytype object from a Contact
func Import(ctx context.Context, client anytype.Client, spaceID, typeKey string, phoneKeys, emailKeys []string, contact Contact, templateID string) error {
	name := contact.DisplayName()
	props := BuildProperties(contact, phoneKeys, emailKeys)

	req := anytype.CreateObjectRequest{
		TypeKey:    typeKey,
		Name:       name,
		Properties: props,
		Icon: &anytype.Icon{
			Format: anytype.IconFormatEmoji,
			Emoji:  "👤",
		},
	}

	if templateID != "" {
		req.TemplateID = templateID
	}

	_, err := client.Space(spaceID).Objects().Create(ctx, req)
	return err
}

// Update updates an existing Anytype object with contact data
func Update(ctx context.Context, client anytype.Client, spaceID string, phoneKeys, emailKeys []string, contact *Contact) error {
	if contact.ObjectID == "" {
		return fmt.Errorf("contact has no ObjectID")
	}

	props := BuildProperties(*contact, phoneKeys, emailKeys)

	req := anytype.UpdateObjectRequest{
		Properties: props,
	}

	return client.Space(spaceID).Object(contact.ObjectID).Update(ctx, req)
}

// BuildProperties constructs the properties slice for a contact
func BuildProperties(contact Contact, phoneKeys, emailKeys []string) []map[string]any {
	var props []map[string]any

	addProp := func(key string, value map[string]any) {
		value["key"] = key
		props = append(props, value)
	}

	addTextProp := func(key, text string) {
		if text != "" {
			addProp(key, map[string]any{"text": text})
		}
	}

	name := contact.DisplayName()
	if name != "Unnamed Contact" {
		addTextProp("name", name)
	}

	addTextProp("given_name", contact.GivenName)
	addTextProp("family_name", contact.FamilyName)
	addTextProp("middle_name", contact.MiddleName)
	addTextProp("prefix", contact.Prefix)
	addTextProp("suffix", contact.Suffix)

	for i, email := range contact.Emails {
		if i >= len(emailKeys) {
			break
		}
		addProp(emailKeys[i], map[string]any{"email": email})
	}

	for i, phone := range contact.Phones {
		if i >= len(phoneKeys) {
			break
		}
		addProp(phoneKeys[i], map[string]any{"phone": phone})
	}

	if len(contact.Addresses) > 0 {
		addr := contact.Addresses[0]
		addTextProp("address", addr.Street)
		addTextProp("city", addr.City)
		addTextProp("region", addr.Region)
		addTextProp("postal_code", addr.PostalCode)
		addTextProp("country", addr.Country)
	}

	addTextProp("organization", contact.Organization)
	addTextProp("title", contact.Title)

	if len(contact.URLs) > 0 {
		addProp("url", map[string]any{"url": contact.URLs[0]})
	}

	notes := BuildNotes(contact)
	if notes != "" {
		addTextProp("notes", notes)
	}

	if contact.Birthday != "" {
		addProp("birthday", map[string]any{"date": ParseBirthday(contact.Birthday)})
	}

	return props
}
