package diff

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/rubiojr/any-vcard/cmd/any-vcard/util"
	"github.com/rubiojr/any-vcard/internal/vcard"
	"github.com/rubiojr/anytype-go"
	"github.com/rubiojr/anytype-go/options"
	"github.com/urfave/cli/v3"
)

// contactWithObjName holds a contact and its Anytype object name
type contactWithObjName struct {
	Contact *vcard.Contact
	ObjName string
}

var Command = &cli.Command{
	Name:  "diff",
	Usage: "Find and diff contacts with the same display name",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:    "name",
			Aliases: []string{"n"},
			Usage:   "Filter by contact name (case-insensitive substring match)",
		},
		&cli.BoolFlag{
			Name:    "verbose",
			Aliases: []string{"v"},
			Usage:   "Show debug output",
		},
	},
	Action: func(ctx context.Context, cmd *cli.Command) error {
		if err := util.RequireFlags(cmd, "app-key", "space"); err != nil {
			return err
		}
		return runDiff(ctx, cmd)
	},
}

func runDiff(ctx context.Context, cmd *cli.Command) error {
	client := util.NewClient(cmd)
	spaceID := cmd.String("space")
	nameFilter := cmd.String("name")
	verbose := cmd.Bool("verbose")

	// Find contact type
	typesResp, err := client.Space(spaceID).Types().List(ctx)
	if err != nil {
		return fmt.Errorf("failed to list types: %w", err)
	}

	var contactTypeKey string
	for _, t := range typesResp {
		if t.Key == "contact" || strings.ToLower(t.Name) == "contact" {
			contactTypeKey = t.Key
			break
		}
	}

	if contactTypeKey == "" {
		return fmt.Errorf("contact type not found in space")
	}

	// Fetch all contacts with pagination using Search
	var allObjects []anytype.Object
	const pageSize = 100
	offset := 0

	searchReq := anytype.SearchRequest{
		Types: []string{contactTypeKey},
	}

	for {
		searchResp, err := client.Space(spaceID).Search(ctx, searchReq,
			options.WithLimit(pageSize),
			options.WithOffset(offset),
		)
		if err != nil {
			return fmt.Errorf("failed to search contacts: %w", err)
		}

		allObjects = append(allObjects, searchResp.Data...)

		if len(searchResp.Data) < pageSize {
			break // No more pages
		}
		offset += pageSize
	}

	if verbose {
		fmt.Printf("Found %d contacts total\n", len(allObjects))
	}

	if len(allObjects) == 0 {
		fmt.Println("No contacts found")
		return nil
	}

	// Normalize the filter the same way we normalize names
	normalizedFilter := ""
	if nameFilter != "" {
		normalizedFilter = vcard.NormalizeNameForDedup(nameFilter)
		if verbose {
			fmt.Printf("Filter: %q -> normalized: %q\n", nameFilter, normalizedFilter)
		}
	}

	// Group contacts by Anytype object name
	byName := make(map[string][]*contactWithObjName)
	for i := range allObjects {
		obj := &allObjects[i]
		contact := objectToContact(obj)
		objName := obj.Name // Use Anytype object name, not contact.DisplayName()
		normalizedName := vcard.NormalizeNameForDedup(objName)

		if verbose && normalizedFilter != "" && strings.Contains(normalizedName, normalizedFilter) {
			fmt.Printf("Match: %q (normalized: %q)\n", objName, normalizedName)
		}

		// Apply name filter if provided
		if normalizedFilter != "" && !strings.Contains(normalizedName, normalizedFilter) {
			continue
		}

		byName[normalizedName] = append(byName[normalizedName], &contactWithObjName{
			Contact: contact,
			ObjName: objName,
		})
	}

	// Find and display duplicates
	var names []string
	for name, contacts := range byName {
		if len(contacts) > 1 {
			names = append(names, name)
		}
	}
	sort.Strings(names)

	if len(names) == 0 {
		fmt.Println("No duplicate contacts found")
		return nil
	}

	for _, name := range names {
		contacts := byName[name]
		fmt.Printf("=== %s (%d contacts) ===\n", contacts[0].ObjName, len(contacts))

		for i, c := range contacts {
			fmt.Printf("\n[%d] ID: %s\n", i+1, c.Contact.ObjectID)
			printContact(c.Contact)
		}

		// Show diff between first and others
		if len(contacts) > 1 {
			fmt.Printf("\n--- Differences ---\n")
			base := contacts[0].Contact
			for i := 1; i < len(contacts); i++ {
				fmt.Printf("\n[1] vs [%d]:\n", i+1)
				printDiff(base, contacts[i].Contact)
			}
		}
		fmt.Println()
	}

	return nil
}

func objectToContact(obj *anytype.Object) *vcard.Contact {
	c := &vcard.Contact{
		ObjectID: obj.ID,
	}

	for _, prop := range obj.Properties {
		switch prop.Key {
		case "name":
			c.FormattedName = prop.Text
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
		case "organization":
			c.Organization = prop.Text
		case "title":
			c.Title = prop.Text
		case "notes":
			c.Note = prop.Text
		case "birthday":
			c.Birthday = prop.Date
		case "email", "email2", "email3":
			if prop.Email != "" {
				c.Emails = append(c.Emails, prop.Email)
			}
		case "phone", "phone2", "phone3":
			if prop.Phone != "" {
				c.Phones = append(c.Phones, prop.Phone)
			}
		case "url":
			if prop.URL != "" {
				c.URLs = append(c.URLs, prop.URL)
			}
		case "address":
			if prop.Text != "" {
				if len(c.Addresses) == 0 {
					c.Addresses = append(c.Addresses, vcard.Address{})
				}
				c.Addresses[0].Street = prop.Text
			}
		case "city":
			if prop.Text != "" {
				if len(c.Addresses) == 0 {
					c.Addresses = append(c.Addresses, vcard.Address{})
				}
				c.Addresses[0].City = prop.Text
			}
		case "region":
			if prop.Text != "" {
				if len(c.Addresses) == 0 {
					c.Addresses = append(c.Addresses, vcard.Address{})
				}
				c.Addresses[0].Region = prop.Text
			}
		case "postal_code":
			if prop.Text != "" {
				if len(c.Addresses) == 0 {
					c.Addresses = append(c.Addresses, vcard.Address{})
				}
				c.Addresses[0].PostalCode = prop.Text
			}
		case "country":
			if prop.Text != "" {
				if len(c.Addresses) == 0 {
					c.Addresses = append(c.Addresses, vcard.Address{})
				}
				c.Addresses[0].Country = prop.Text
			}
		}
	}

	return c
}

func printContact(c *vcard.Contact) {
	if c.GivenName != "" || c.FamilyName != "" {
		fmt.Printf("  Name: %s %s\n", c.GivenName, c.FamilyName)
	}
	if c.Organization != "" {
		fmt.Printf("  Organization: %s\n", c.Organization)
	}
	if c.Title != "" {
		fmt.Printf("  Title: %s\n", c.Title)
	}
	for i, phone := range c.Phones {
		fmt.Printf("  Phone %d: %s\n", i+1, phone)
	}
	for i, email := range c.Emails {
		fmt.Printf("  Email %d: %s\n", i+1, email)
	}
	if len(c.Addresses) > 0 {
		addr := c.Addresses[0]
		parts := filterEmpty(addr.Street, addr.City, addr.Region, addr.PostalCode, addr.Country)
		if len(parts) > 0 {
			fmt.Printf("  Address: %s\n", strings.Join(parts, ", "))
		}
	}
	for i, url := range c.URLs {
		fmt.Printf("  URL %d: %s\n", i+1, url)
	}
	if c.Birthday != "" {
		fmt.Printf("  Birthday: %s\n", c.Birthday)
	}
	if c.Note != "" {
		note := c.Note
		if len(note) > 50 {
			note = note[:50] + "..."
		}
		fmt.Printf("  Note: %s\n", note)
	}
}

func printDiff(a, b *vcard.Contact) {
	diffField("GivenName", a.GivenName, b.GivenName)
	diffField("FamilyName", a.FamilyName, b.FamilyName)
	diffField("MiddleName", a.MiddleName, b.MiddleName)
	diffField("Prefix", a.Prefix, b.Prefix)
	diffField("Suffix", a.Suffix, b.Suffix)
	diffField("Organization", a.Organization, b.Organization)
	diffField("Title", a.Title, b.Title)
	diffField("Birthday", a.Birthday, b.Birthday)
	diffSlice("Phones", a.Phones, b.Phones)
	diffSlice("Emails", a.Emails, b.Emails)
	diffSlice("URLs", a.URLs, b.URLs)

	// Address diff
	var addrA, addrB string
	if len(a.Addresses) > 0 {
		addr := a.Addresses[0]
		addrA = strings.Join(filterEmpty(addr.Street, addr.City, addr.Region, addr.PostalCode, addr.Country), ", ")
	}
	if len(b.Addresses) > 0 {
		addr := b.Addresses[0]
		addrB = strings.Join(filterEmpty(addr.Street, addr.City, addr.Region, addr.PostalCode, addr.Country), ", ")
	}
	diffField("Address", addrA, addrB)

	// Note diff (truncated)
	noteA, noteB := a.Note, b.Note
	if len(noteA) > 30 {
		noteA = noteA[:30] + "..."
	}
	if len(noteB) > 30 {
		noteB = noteB[:30] + "..."
	}
	diffField("Note", noteA, noteB)
}

func diffField(name, a, b string) {
	if a != b {
		if a == "" {
			fmt.Printf("  %s: (empty) → %q\n", name, b)
		} else if b == "" {
			fmt.Printf("  %s: %q → (empty)\n", name, a)
		} else {
			fmt.Printf("  %s: %q → %q\n", name, a, b)
		}
	}
}

func diffSlice(name string, a, b []string) {
	// Find items only in a
	bSet := make(map[string]bool)
	for _, v := range b {
		bSet[v] = true
	}
	for _, v := range a {
		if !bSet[v] {
			fmt.Printf("  %s: -%s\n", name, v)
		}
	}

	// Find items only in b
	aSet := make(map[string]bool)
	for _, v := range a {
		aSet[v] = true
	}
	for _, v := range b {
		if !aSet[v] {
			fmt.Printf("  %s: +%s\n", name, v)
		}
	}
}

func filterEmpty(parts ...string) []string {
	var result []string
	for _, p := range parts {
		if p != "" {
			result = append(result, p)
		}
	}
	return result
}
