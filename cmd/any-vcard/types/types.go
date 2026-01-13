package types

import (
	"context"
	"fmt"
	"strings"

	"github.com/rubiojr/any-vcard/cmd/any-vcard/util"
	"github.com/urfave/cli/v3"
)

var Command = &cli.Command{
	Name:  "types",
	Usage: "Manage object types in the space",
	Commands: []*cli.Command{
		listCommand,
		createCommand,
	},
}

var listCommand = &cli.Command{
	Name:  "list",
	Usage: "List available object types in the space",
	Action: func(ctx context.Context, cmd *cli.Command) error {
		if err := util.RequireFlags(cmd, "app-key", "space"); err != nil {
			return err
		}
		return listTypes(ctx, cmd)
	},
}

var createCommand = &cli.Command{
	Name:  "create",
	Usage: "Create the Contact object type in the space",
	Action: func(ctx context.Context, cmd *cli.Command) error {
		if err := util.RequireFlags(cmd, "app-key", "space"); err != nil {
			return err
		}
		return createContactType(ctx, cmd)
	},
}

func listTypes(ctx context.Context, cmd *cli.Command) error {
	client := util.NewClient(cmd)
	spaceID := cmd.String("space")

	types, err := client.Space(spaceID).Types().List(ctx)
	if err != nil {
		return fmt.Errorf("failed to list types: %w", err)
	}

	fmt.Printf("\nAvailable object types in space %s:\n\n", spaceID)
	for _, t := range types {
		fmt.Printf("- %s (key: %s)\n", t.Name, t.Key)
		if t.Description != "" {
			fmt.Printf("  %s\n", t.Description)
		}
	}
	fmt.Printf("\nTotal: %d types\n", len(types))

	return nil
}

func createContactType(ctx context.Context, cmd *cli.Command) error {
	client := util.NewClient(cmd)
	spaceID := cmd.String("space")

	fmt.Printf("Creating Contact type in space %s...\n", spaceID)

	types, err := client.Space(spaceID).Types().List(ctx)
	if err != nil {
		return fmt.Errorf("failed to list types: %w", err)
	}

	for _, t := range types {
		if strings.EqualFold(t.Key, "contact") || strings.EqualFold(t.Name, "contact") {
			fmt.Printf("✓ Contact type already exists with key: %s\n", t.Key)
			fmt.Printf("  Properties: %d\n", len(t.PropertyDefinitions))
			return nil
		}
	}

	typeResp, err := util.CreateContactType(ctx, client, spaceID)
	if err != nil {
		return fmt.Errorf("failed to create Contact type: %w", err)
	}

	fmt.Printf("✓ Contact type created successfully!\n")
	fmt.Printf("  Type Key: %s\n", typeResp.Type.Key)
	fmt.Printf("  Type Name: %s\n", typeResp.Type.Name)
	fmt.Printf("  Properties: %d\n", len(typeResp.Type.PropertyDefinitions))
	fmt.Printf("\nProperty definitions:\n")
	for _, prop := range typeResp.Type.PropertyDefinitions {
		fmt.Printf("  - %s (%s): %s\n", prop.Name, prop.Key, prop.Format)
	}

	return nil
}
