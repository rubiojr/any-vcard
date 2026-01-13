package vcardimport

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/emersion/go-vcard"
	"github.com/epheo/anytype-go"
	"github.com/rubiojr/any-vcard/cmd/any-vcard/util"
	"github.com/urfave/cli/v3"
)

// VCardContact represents a parsed vCard contact
type VCardContact struct {
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

var Command = &cli.Command{
	Name:      "import",
	Usage:     "Import vCard file(s) into Anytype",
	ArgsUsage: "<vcard-file> [vcard-file...]",
	Flags: []cli.Flag{
		&cli.BoolFlag{
			Name:  "create-type",
			Usage: "Create Contact object type if it doesn't exist",
			Value: true,
		},
		&cli.BoolFlag{
			Name:  "skip-duplicates",
			Usage: "Skip importing contacts that already exist (based on name+email or name+phone)",
			Value: true,
		},
		&cli.BoolFlag{
			Name:  "dry-run",
			Usage: "Parse vCard files without importing",
		},
	},
	Action: func(ctx context.Context, cmd *cli.Command) error {
		if err := util.RequireFlags(cmd, "app-key", "space"); err != nil {
			return err
		}
		if cmd.Args().Len() == 0 {
			return fmt.Errorf("at least one vCard file is required")
		}
		return importVCards(ctx, cmd)
	},
}

func importVCards(ctx context.Context, cmd *cli.Command) error {
	client := util.NewClient(cmd)
	spaceID := cmd.String("space")
	createType := cmd.Bool("create-type")
	skipDuplicates := cmd.Bool("skip-duplicates")
	dryRun := cmd.Bool("dry-run")

	var allContacts []VCardContact
	for i := 0; i < cmd.Args().Len(); i++ {
		filePath := cmd.Args().Get(i)
		contacts, err := ParseVCardFile(filePath)
		if err != nil {
			log.Printf("Error parsing %s: %v", filePath, err)
			continue
		}
		allContacts = append(allContacts, contacts...)
		fmt.Printf("✓ Parsed %d contact(s) from %s\n", len(contacts), filePath)
	}

	if len(allContacts) == 0 {
		return fmt.Errorf("no contacts found in provided files")
	}

	if dryRun {
		fmt.Printf("\nDry run mode - would import %d contact(s):\n", len(allContacts))
		for i, contact := range allContacts {
			fmt.Printf("\n%d. %s\n", i+1, contact.FormattedName)
			if len(contact.Emails) > 0 {
				fmt.Printf("   Email: %s\n", strings.Join(contact.Emails, ", "))
			}
			if len(contact.Phones) > 0 {
				fmt.Printf("   Phone: %s\n", strings.Join(contact.Phones, ", "))
			}
		}
		return nil
	}

	typeKey := util.ContactTypeKey
	types, err := client.Space(spaceID).Types().List(ctx)
	if err != nil {
		return fmt.Errorf("failed to list types: %w", err)
	}

	contactTypeExists := false
	for _, t := range types {
		if strings.EqualFold(t.Key, util.ContactTypeKey) || strings.EqualFold(t.Name, "contact") {
			contactTypeExists = true
			typeKey = t.Key
			fmt.Printf("✓ Found existing Contact type with key: %s\n", typeKey)
			break
		}
	}

	if !contactTypeExists {
		if !createType {
			return fmt.Errorf("Contact type not found and --create-type=false")
		}

		fmt.Printf("Creating Contact object type...\n")
		typeResp, err := util.CreateContactType(ctx, client, spaceID)
		if err != nil {
			return fmt.Errorf("failed to create Contact type: %w", err)
		}
		typeKey = typeResp.Type.Key
		fmt.Printf("✓ Created Contact type with key: %s\n", typeKey)
	}

	phoneKeys, emailKeys, err := util.EnsureContactProperties(ctx, client, spaceID)
	if err != nil {
		return fmt.Errorf("failed to ensure properties: %w", err)
	}

	var existingContacts []anytype.Object
	if skipDuplicates {
		fmt.Printf("Checking for existing contacts...\n")
		searchResp, err := client.Space(spaceID).Search(ctx, anytype.SearchRequest{
			Types: []string{typeKey},
		})
		if err != nil {
			log.Printf("Warning: could not search for existing contacts: %v", err)
		} else {
			existingContacts = searchResp.Data
			fmt.Printf("✓ Found %d existing contacts\n", len(existingContacts))
		}
	}

	fmt.Printf("\nImporting %d contact(s)...\n", len(allContacts))
	successCount := 0
	skippedCount := 0
	for i, contact := range allContacts {
		if skipDuplicates && isDuplicate(contact, existingContacts) {
			log.Printf("Skipping duplicate contact %d (%s)", i+1, contact.FormattedName)
			skippedCount++
			continue
		}

		if err := ImportContact(ctx, client, spaceID, typeKey, phoneKeys, emailKeys, contact); err != nil {
			log.Printf("Error importing contact %d (%s): %v", i+1, contact.FormattedName, err)
			continue
		}
		successCount++
		fmt.Printf("✓ Imported: %s\n", contact.FormattedName)
	}

	fmt.Printf("\n✓ Successfully imported %d/%d contacts", successCount, len(allContacts))
	if skippedCount > 0 {
		fmt.Printf(" (skipped %d duplicates)", skippedCount)
	}
	fmt.Printf("\n")
	return nil
}

// ParseVCardFile parses a vCard file and returns the contacts
func ParseVCardFile(filePath string) ([]VCardContact, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	decoder := vcard.NewDecoder(file)
	var contacts []VCardContact

	for {
		card, err := decoder.Decode()
		if err != nil {
			break
		}
		contact := parseVCard(card)
		contacts = append(contacts, contact)
	}

	return contacts, nil
}

func parseVCard(card vcard.Card) VCardContact {
	contact := VCardContact{}

	if fn := card.PreferredValue(vcard.FieldFormattedName); fn != "" {
		contact.FormattedName = fn
	}

	if names := card.Name(); names != nil {
		contact.FamilyName = names.FamilyName
		contact.GivenName = names.GivenName
		contact.MiddleName = names.AdditionalName
		contact.Prefix = names.HonorificPrefix
		contact.Suffix = names.HonorificSuffix
	}

	for _, email := range card.Values(vcard.FieldEmail) {
		if email != "" {
			email = strings.TrimPrefix(email, "mailto:")
			contact.Emails = append(contact.Emails, email)
		}
	}

	for _, phone := range card.Values(vcard.FieldTelephone) {
		if phone != "" {
			phone = strings.TrimPrefix(phone, "tel:")
			contact.Phones = append(contact.Phones, phone)
		}
	}

	if addr := card.Address(); addr != nil {
		addrStr := addr.StreetAddress
		if addrStr == "" {
			addrStr = addr.ExtendedAddress
		}
		contact.Addresses = append(contact.Addresses, Address{
			Street:     addrStr,
			City:       addr.Locality,
			Region:     addr.Region,
			PostalCode: addr.PostalCode,
			Country:    addr.Country,
			Full:       addrStr,
		})
	}

	if org := card.PreferredValue(vcard.FieldOrganization); org != "" {
		contact.Organization = org
	}

	if title := card.PreferredValue(vcard.FieldTitle); title != "" {
		contact.Title = title
	}

	for _, url := range card.Values(vcard.FieldURL) {
		if url != "" {
			url = strings.TrimSpace(url)
			contact.URLs = append(contact.URLs, url)
		}
	}

	if note := card.PreferredValue(vcard.FieldNote); note != "" {
		contact.Note = note
	}

	if bday := card.PreferredValue(vcard.FieldBirthday); bday != "" {
		contact.Birthday = bday
	}

	if photo := card.PreferredValue(vcard.FieldPhoto); photo != "" {
		contact.Photo = photo
	}

	return contact
}

func isDuplicate(contact VCardContact, existingContacts []anytype.Object) bool {
	contactName := strings.ToLower(strings.TrimSpace(contact.FormattedName))
	if contactName == "" {
		parts := []string{}
		if contact.GivenName != "" {
			parts = append(parts, contact.GivenName)
		}
		if contact.FamilyName != "" {
			parts = append(parts, contact.FamilyName)
		}
		contactName = strings.ToLower(strings.TrimSpace(strings.Join(parts, " ")))
	}

	contactEmail := ""
	if len(contact.Emails) > 0 {
		contactEmail = strings.ToLower(strings.TrimSpace(contact.Emails[0]))
	}

	contactPhone := ""
	if len(contact.Phones) > 0 {
		contactPhone = strings.ToLower(strings.TrimSpace(contact.Phones[0]))
		contactPhone = strings.ReplaceAll(contactPhone, " ", "")
		contactPhone = strings.ReplaceAll(contactPhone, "-", "")
		contactPhone = strings.ReplaceAll(contactPhone, "(", "")
		contactPhone = strings.ReplaceAll(contactPhone, ")", "")
	}

	for _, existing := range existingContacts {
		existingName := strings.ToLower(strings.TrimSpace(existing.Name))

		if contactName != "" && existingName != "" && contactName == existingName {
			if contactEmail != "" && len(existing.Properties) > 0 {
				for _, prop := range existing.Properties {
					if prop.Key == "email" && prop.Email != "" {
						if strings.ToLower(strings.TrimSpace(prop.Email)) == contactEmail {
							return true
						}
					}
				}
			}

			if contactPhone != "" && len(existing.Properties) > 0 {
				for _, prop := range existing.Properties {
					if prop.Key == "phone" && prop.Phone != "" {
						existingPhone := strings.ToLower(strings.TrimSpace(prop.Phone))
						existingPhone = strings.ReplaceAll(existingPhone, " ", "")
						existingPhone = strings.ReplaceAll(existingPhone, "-", "")
						existingPhone = strings.ReplaceAll(existingPhone, "(", "")
						existingPhone = strings.ReplaceAll(existingPhone, ")", "")
						if existingPhone == contactPhone {
							return true
						}
					}
				}
			}
		}
	}

	return false
}

// ImportContact imports a single contact into Anytype
func ImportContact(ctx context.Context, client anytype.Client, spaceID, typeKey string, phoneKeys, emailKeys []string, contact VCardContact) error {
	name := contact.FormattedName
	if name == "" {
		parts := []string{}
		if contact.Prefix != "" {
			parts = append(parts, contact.Prefix)
		}
		if contact.GivenName != "" {
			parts = append(parts, contact.GivenName)
		}
		if contact.MiddleName != "" {
			parts = append(parts, contact.MiddleName)
		}
		if contact.FamilyName != "" {
			parts = append(parts, contact.FamilyName)
		}
		if contact.Suffix != "" {
			parts = append(parts, contact.Suffix)
		}
		if len(parts) > 0 {
			name = strings.Join(parts, " ")
		}
	}

	if name == "" {
		name = "Unnamed Contact"
	}

	var propsSlice []map[string]any

	addProp := func(key string, value map[string]any) {
		value["key"] = key
		propsSlice = append(propsSlice, value)
	}

	if name != "" && name != "Unnamed Contact" {
		addProp("name", map[string]any{"text": name})
	}

	if contact.GivenName != "" {
		addProp("given_name", map[string]any{"text": contact.GivenName})
	}
	if contact.FamilyName != "" {
		addProp("family_name", map[string]any{"text": contact.FamilyName})
	}
	if contact.MiddleName != "" {
		addProp("middle_name", map[string]any{"text": contact.MiddleName})
	}
	if contact.Prefix != "" {
		addProp("prefix", map[string]any{"text": contact.Prefix})
	}
	if contact.Suffix != "" {
		addProp("suffix", map[string]any{"text": contact.Suffix})
	}

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
		if addr.Street != "" {
			addProp("address", map[string]any{"text": addr.Street})
		}
		if addr.City != "" {
			addProp("city", map[string]any{"text": addr.City})
		}
		if addr.Region != "" {
			addProp("region", map[string]any{"text": addr.Region})
		}
		if addr.PostalCode != "" {
			addProp("postal_code", map[string]any{"text": addr.PostalCode})
		}
		if addr.Country != "" {
			addProp("country", map[string]any{"text": addr.Country})
		}
	}

	if contact.Organization != "" {
		addProp("organization", map[string]any{"text": contact.Organization})
	}

	if contact.Title != "" {
		addProp("title", map[string]any{"text": contact.Title})
	}

	if len(contact.URLs) > 0 {
		addProp("url", map[string]any{"url": contact.URLs[0]})
	}

	notes := []string{}
	if contact.Note != "" {
		notes = append(notes, contact.Note)
	}
	if len(contact.Emails) > 3 {
		notes = append(notes, "Additional emails: "+strings.Join(contact.Emails[3:], ", "))
	}
	if len(contact.URLs) > 1 {
		notes = append(notes, "Additional URLs: "+strings.Join(contact.URLs[1:], ", "))
	}

	if len(notes) > 0 {
		addProp("notes", map[string]any{"text": strings.Join(notes, "\n\n")})
	}

	if contact.Birthday != "" {
		var birthdayFormatted string
		if t, err := time.Parse("20060102", contact.Birthday); err == nil {
			birthdayFormatted = t.Format(time.RFC3339)
		} else if t, err := time.Parse("2006-01-02", contact.Birthday); err == nil {
			birthdayFormatted = t.Format(time.RFC3339)
		} else {
			birthdayFormatted = contact.Birthday
		}
		addProp("birthday", map[string]any{"date": birthdayFormatted})
	}

	req := anytype.CreateObjectRequest{
		TypeKey:    typeKey,
		Name:       name,
		Properties: propsSlice,
		Icon: &anytype.Icon{
			Format: anytype.IconFormatEmoji,
			Emoji:  "👤",
		},
	}

	_, err := client.Space(spaceID).Objects().Create(ctx, req)
	return err
}
