package vcardimport

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/rubiojr/any-vcard/cmd/any-vcard/util"
	"github.com/rubiojr/any-vcard/internal/vcard"
	"github.com/rubiojr/anytype-go"
	"github.com/rubiojr/anytype-go/options"
	"github.com/urfave/cli/v3"
)

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
			Name:  "merge-duplicates",
			Usage: "Merge missing fields into existing duplicates (default: true)",
			Value: true,
		},
		&cli.BoolFlag{
			Name:  "skip-duplicates",
			Usage: "Skip duplicates without merging (overrides --merge-duplicates)",
			Value: false,
		},
		&cli.BoolFlag{
			Name:  "dry-run",
			Usage: "Parse vCard files without importing",
		},
		&cli.StringFlag{
			Name:    "template",
			Aliases: []string{"t"},
			Usage:   "Template ID to use when creating new contacts",
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
	dryRun := cmd.Bool("dry-run")
	skipDuplicates := cmd.Bool("skip-duplicates")
	mergeDuplicates := cmd.Bool("merge-duplicates") && !skipDuplicates // skip overrides merge
	templateID := cmd.String("template")

	allContacts, err := parseAllFiles(cmd)
	if err != nil {
		return err
	}

	if dryRun {
		printDryRun(allContacts)
		return nil
	}

	typeKey, err := ensureContactType(ctx, client, spaceID, cmd.Bool("create-type"))
	if err != nil {
		return err
	}

	phoneKeys, emailKeys, err := util.EnsureContactProperties(ctx, client, spaceID)
	if err != nil {
		return fmt.Errorf("failed to ensure properties: %w", err)
	}

	var dedupIndex *vcard.DedupIndex
	if skipDuplicates || mergeDuplicates {
		dedupIndex = fetchExistingContacts(ctx, client, spaceID, typeKey)
	} else {
		dedupIndex = vcard.NewDedupIndex(nil)
	}

	return importContacts(ctx, client, spaceID, typeKey, phoneKeys, emailKeys, allContacts, dedupIndex, mergeDuplicates, templateID)
}

func parseAllFiles(cmd *cli.Command) ([]vcard.Contact, error) {
	var allContacts []vcard.Contact
	for i := 0; i < cmd.Args().Len(); i++ {
		filePath := cmd.Args().Get(i)
		contacts, err := vcard.ParseFile(filePath)
		if err != nil {
			log.Printf("Error parsing %s: %v", filePath, err)
			continue
		}
		allContacts = append(allContacts, contacts...)
		fmt.Printf("✓ Parsed %d contact(s) from %s\n", len(contacts), filePath)
	}

	if len(allContacts) == 0 {
		return nil, fmt.Errorf("no contacts found in provided files")
	}
	return allContacts, nil
}

func printDryRun(contacts []vcard.Contact) {
	fmt.Printf("\nDry run mode - would import %d contact(s):\n", len(contacts))
	for i, contact := range contacts {
		fmt.Printf("\n%d. %s\n", i+1, contact.DisplayName())
		if len(contact.Emails) > 0 {
			fmt.Printf("   Email: %s\n", strings.Join(contact.Emails, ", "))
		}
		if len(contact.Phones) > 0 {
			fmt.Printf("   Phone: %s\n", strings.Join(contact.Phones, ", "))
		}
	}
}

func ensureContactType(ctx context.Context, client anytype.Client, spaceID string, createType bool) (string, error) {
	types, err := client.Space(spaceID).Types().List(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to list types: %w", err)
	}

	for _, t := range types {
		if strings.EqualFold(t.Key, util.ContactTypeKey) || strings.EqualFold(t.Name, "contact") {
			fmt.Printf("✓ Found existing Contact type with key: %s\n", t.Key)
			return t.Key, nil
		}
	}

	if !createType {
		return "", fmt.Errorf("Contact type not found and --create-type=false")
	}

	fmt.Printf("Creating Contact object type...\n")
	typeResp, err := util.CreateContactType(ctx, client, spaceID)
	if err != nil {
		return "", fmt.Errorf("failed to create Contact type: %w", err)
	}
	fmt.Printf("✓ Created Contact type with key: %s\n", typeResp.Type.Key)
	return typeResp.Type.Key, nil
}

func fetchExistingContacts(ctx context.Context, client anytype.Client, spaceID, typeKey string) *vcard.DedupIndex {
	fmt.Printf("Checking for existing contacts...\n")

	// Fetch all contacts with pagination using Search
	var allObjects []anytype.Object
	const pageSize = 100
	offset := 0

	searchReq := anytype.SearchRequest{
		Types: []string{typeKey},
	}

	for {
		searchResp, err := client.Space(spaceID).Search(ctx, searchReq,
			options.WithLimit(pageSize),
			options.WithOffset(offset),
		)
		if err != nil {
			log.Printf("Warning: could not search contacts: %v", err)
			return vcard.NewDedupIndex(nil)
		}

		allObjects = append(allObjects, searchResp.Data...)

		if len(searchResp.Data) < pageSize {
			break // No more pages
		}
		offset += pageSize
	}

	fmt.Printf("✓ Found %d existing contacts\n", len(allObjects))

	// Convert Anytype objects to contacts for indexing
	contacts := make([]*vcard.Contact, 0, len(allObjects))
	for _, obj := range allObjects {
		contacts = append(contacts, anytypeObjectToContact(obj))
	}

	return vcard.NewDedupIndex(contacts)
}

// anytypeObjectToContact converts an Anytype object to a Contact for dedup
func anytypeObjectToContact(obj anytype.Object) *vcard.Contact {
	c := &vcard.Contact{
		FormattedName: obj.Name,
		ObjectID:      obj.ID,
	}

	for _, prop := range obj.Properties {
		switch prop.Key {
		case "email", "email_2", "email_3":
			if prop.Email != "" {
				c.Emails = append(c.Emails, prop.Email)
			}
		case "phone", "phone_2", "phone_3":
			if prop.Phone != "" {
				c.Phones = append(c.Phones, prop.Phone)
			}
		case "organization":
			c.Organization = prop.Text
		case "title":
			c.Title = prop.Text
		case "birthday":
			c.Birthday = prop.Date
		case "given_name":
			c.GivenName = prop.Text
		case "family_name":
			c.FamilyName = prop.Text
		case "middle_name":
			c.MiddleName = prop.Text
		case "prefix":
			c.Prefix = prop.Text
		case "suffix":
			c.Suffix = prop.Text
		case "notes":
			c.Note = prop.Text
		case "url":
			if prop.URL != "" {
				c.URLs = append(c.URLs, prop.URL)
			}
		case "address":
			if prop.Text != "" && len(c.Addresses) == 0 {
				c.Addresses = append(c.Addresses, vcard.Address{Street: prop.Text})
			} else if len(c.Addresses) > 0 {
				c.Addresses[0].Street = prop.Text
			}
		case "city":
			if len(c.Addresses) == 0 {
				c.Addresses = append(c.Addresses, vcard.Address{})
			}
			c.Addresses[0].City = prop.Text
		case "region":
			if len(c.Addresses) == 0 {
				c.Addresses = append(c.Addresses, vcard.Address{})
			}
			c.Addresses[0].Region = prop.Text
		case "postal_code":
			if len(c.Addresses) == 0 {
				c.Addresses = append(c.Addresses, vcard.Address{})
			}
			c.Addresses[0].PostalCode = prop.Text
		case "country":
			if len(c.Addresses) == 0 {
				c.Addresses = append(c.Addresses, vcard.Address{})
			}
			c.Addresses[0].Country = prop.Text
		}
	}

	return c
}

func importContacts(ctx context.Context, client anytype.Client, spaceID, typeKey string, phoneKeys, emailKeys []string, contacts []vcard.Contact, dedupIndex *vcard.DedupIndex, mergeDuplicates bool, templateID string) error {
	fmt.Printf("\nImporting %d contact(s)...\n", len(contacts))

	var successCount, skippedCount, mergedCount int
	for i := range contacts {
		contact := &contacts[i]

		duplicates := dedupIndex.FindDuplicates(contact)
		if len(duplicates) > 0 {
			if mergeDuplicates {
				// Merge into the first duplicate found
				existing := duplicates[0]
				if vcard.MergeContacts(existing, contact) {
					// Update the existing contact in Anytype
					if err := updateContact(ctx, client, spaceID, phoneKeys, emailKeys, existing); err != nil {
						log.Printf("Error merging contact %d (%s): %v", i+1, contact.DisplayName(), err)
						continue
					}
					mergedCount++
					fmt.Printf("⊕ Merged: %s → %s\n", contact.DisplayName(), existing.DisplayName())
				} else {
					log.Printf("Skipping %s (nothing new to merge)", contact.DisplayName())
					skippedCount++
				}
			} else {
				log.Printf("Skipping duplicate contact %d (%s)", i+1, contact.DisplayName())
				skippedCount++
			}
			continue
		}

		if err := importContact(ctx, client, spaceID, typeKey, phoneKeys, emailKeys, *contact, templateID); err != nil {
			log.Printf("Error importing contact %d (%s): %v", i+1, contact.DisplayName(), err)
			continue
		}

		// Add to index to catch duplicates within the import batch
		dedupIndex.Add(contact)

		successCount++
		fmt.Printf("✓ Imported: %s\n", contact.DisplayName())
	}

	fmt.Printf("\n✓ Successfully imported %d/%d contacts", successCount, len(contacts))
	if mergedCount > 0 {
		fmt.Printf(" (merged %d)", mergedCount)
	}
	if skippedCount > 0 {
		fmt.Printf(" (skipped %d duplicates)", skippedCount)
	}
	fmt.Printf("\n")
	return nil
}

func importContact(ctx context.Context, client anytype.Client, spaceID, typeKey string, phoneKeys, emailKeys []string, contact vcard.Contact, templateID string) error {
	return vcard.Import(ctx, client, spaceID, typeKey, phoneKeys, emailKeys, contact, templateID)
}

// updateContact updates an existing contact with merged data
func updateContact(ctx context.Context, client anytype.Client, spaceID string, phoneKeys, emailKeys []string, contact *vcard.Contact) error {
	return vcard.Update(ctx, client, spaceID, phoneKeys, emailKeys, contact)
}
