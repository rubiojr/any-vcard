package template

import (
	"context"
	"fmt"

	"github.com/rubiojr/any-vcard/cmd/any-vcard/util"
	"github.com/urfave/cli/v3"
)

var Command = &cli.Command{
	Name:  "template",
	Usage: "Manage templates",
	Commands: []*cli.Command{
		listCommand,
	},
}

var listCommand = &cli.Command{
	Name:  "list",
	Usage: "List available templates for all types",
	Flags: []cli.Flag{
		&cli.BoolFlag{
			Name:    "verbose",
			Aliases: []string{"v"},
			Usage:   "Show verbose output",
		},
	},
	Action: func(ctx context.Context, cmd *cli.Command) error {
		if err := util.RequireFlags(cmd, "app-key", "space"); err != nil {
			return err
		}
		return listTemplates(ctx, cmd)
	},
}

func listTemplates(ctx context.Context, cmd *cli.Command) error {
	client := util.NewClient(cmd)
	spaceID := cmd.String("space")
	verbose := cmd.Bool("verbose")

	// List all types
	typesResp, err := client.Space(spaceID).Types().List(ctx)
	if err != nil {
		return fmt.Errorf("failed to list types: %w", err)
	}

	if verbose {
		fmt.Printf("Found %d types\n\n", len(typesResp))
	}

	totalTemplates := 0

	for _, t := range typesResp {
		if verbose {
			fmt.Printf("Checking type: %s (key: %s, id: %s)\n", t.Name, t.Key, t.ID)
		}

		// Use type ID for template lookup (API requires ID, not key)
		templates, err := client.Space(spaceID).Type(t.ID).Templates().List(ctx)
		if err != nil {
			if verbose {
				fmt.Printf("  Error: %v\n", err)
			}
			continue
		}

		if verbose {
			fmt.Printf("  Found %d templates\n", len(templates))
		}

		if len(templates) == 0 {
			continue
		}

		fmt.Printf("%s:\n", t.Name)
		for _, tmpl := range templates {
			status := ""
			if tmpl.Archived {
				status = " (archived)"
			}
			fmt.Printf("  %s%s\n", tmpl.Name, status)
			fmt.Printf("    ID: %s\n", tmpl.ID)
			totalTemplates++
		}
		fmt.Println()
	}

	if totalTemplates == 0 {
		fmt.Println("No templates found")
	}

	return nil
}
