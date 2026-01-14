package vcardimport

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/rubiojr/any-vcard/cmd/any-vcard/util"
	"github.com/rubiojr/any-vcard/internal/vcard"
	"github.com/rubiojr/anytype-go"
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
	dryRun := cmd.Bool("dry-run")

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
	if cmd.Bool("skip-duplicates") {
		dedupIndex = fetchExistingContacts(ctx, client, spaceID, typeKey)
	} else {
		dedupIndex = vcard.NewDedupIndex(nil)
	}

	return importContacts(ctx, client, spaceID, typeKey, phoneKeys, emailKeys, allContacts, dedupIndex)
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
	searchResp, err := client.Space(spaceID).Search(ctx, anytype.SearchRequest{
		Types: []string{typeKey},
	})
	if err != nil {
		log.Printf("Warning: could not search for existing contacts: %v", err)
		return vcard.NewDedupIndex(nil)
	}
	fmt.Printf("✓ Found %d existing contacts\n", len(searchResp.Data))

	// Convert Anytype objects to contacts for indexing
	contacts := make([]*vcard.Contact, 0, len(searchResp.Data))
	for _, obj := range searchResp.Data {
		contacts = append(contacts, anytypeObjectToContact(obj))
	}

	return vcard.NewDedupIndex(contacts)
}

// anytypeObjectToContact converts an Anytype object to a Contact for dedup
func anytypeObjectToContact(obj anytype.Object) *vcard.Contact {
	c := &vcard.Contact{
		FormattedName: obj.Name,
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
		case "birthday":
			c.Birthday = prop.Date
		}
	}

	return c
}

func importContacts(ctx context.Context, client anytype.Client, spaceID, typeKey string, phoneKeys, emailKeys []string, contacts []vcard.Contact, dedupIndex *vcard.DedupIndex) error {
	fmt.Printf("\nImporting %d contact(s)...\n", len(contacts))

	var successCount, skippedCount int
	for i := range contacts {
		contact := &contacts[i]

		if dedupIndex.IsDuplicate(contact) {
			log.Printf("Skipping duplicate contact %d (%s)", i+1, contact.DisplayName())
			skippedCount++
			continue
		}

		if err := importContact(ctx, client, spaceID, typeKey, phoneKeys, emailKeys, *contact); err != nil {
			log.Printf("Error importing contact %d (%s): %v", i+1, contact.DisplayName(), err)
			continue
		}

		// Add to index to catch duplicates within the import batch
		dedupIndex.Add(contact)

		successCount++
		fmt.Printf("✓ Imported: %s\n", contact.DisplayName())
	}

	fmt.Printf("\n✓ Successfully imported %d/%d contacts", successCount, len(contacts))
	if skippedCount > 0 {
		fmt.Printf(" (skipped %d duplicates)", skippedCount)
	}
	fmt.Printf("\n")
	return nil
}

func importContact(ctx context.Context, client anytype.Client, spaceID, typeKey string, phoneKeys, emailKeys []string, contact vcard.Contact) error {
	return vcard.Import(ctx, client, spaceID, typeKey, phoneKeys, emailKeys, contact)
}
