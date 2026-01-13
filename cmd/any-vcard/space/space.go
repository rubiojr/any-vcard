package space

import (
	"context"
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/rubiojr/anytype-go"
	"github.com/rubiojr/any-vcard/cmd/any-vcard/util"
	"github.com/urfave/cli/v3"
)

var Command = &cli.Command{
	Name:  "space",
	Usage: "Manage Anytype spaces",
	Commands: []*cli.Command{
		listCommand,
		createCommand,
	},
}

var listCommand = &cli.Command{
	Name:  "list",
	Usage: "List available spaces",
	Action: func(ctx context.Context, cmd *cli.Command) error {
		if err := util.RequireFlags(cmd, "app-key"); err != nil {
			return err
		}
		return listSpaces(ctx, cmd)
	},
}

var createCommand = &cli.Command{
	Name:      "create",
	Usage:     "Create a new space",
	ArgsUsage: "<space-name>",
	Action: func(ctx context.Context, cmd *cli.Command) error {
		if err := util.RequireFlags(cmd, "app-key"); err != nil {
			return err
		}
		if cmd.Args().Len() == 0 {
			return fmt.Errorf("space name is required")
		}
		return createSpace(ctx, cmd)
	},
}

func listSpaces(ctx context.Context, cmd *cli.Command) error {
	client := util.NewClientWithAppKey(cmd.String("url"), cmd.String("app-key"))

	resp, err := client.Spaces().List(ctx)
	if err != nil {
		return fmt.Errorf("failed to list spaces: %w", err)
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "NAME\tID\tDESCRIPTION")
	for _, s := range resp.Data {
		fmt.Fprintf(w, "%s\t%s\t%s\n", s.Name, s.ID, s.Description)
	}
	w.Flush()

	return nil
}

func createSpace(ctx context.Context, cmd *cli.Command) error {
	client := util.NewClientWithAppKey(cmd.String("url"), cmd.String("app-key"))
	spaceName := cmd.Args().Get(0)

	fmt.Printf("Creating space %q...\n", spaceName)

	resp, err := client.Spaces().Create(ctx, anytype.CreateSpaceRequest{
		Name: spaceName,
	})
	if err != nil {
		return fmt.Errorf("failed to create space: %w", err)
	}

	fmt.Printf("✓ Space created successfully!\n")
	fmt.Printf("  Name: %s\n", resp.Space.Name)
	fmt.Printf("  ID: %s\n", resp.Space.ID)

	return nil
}
