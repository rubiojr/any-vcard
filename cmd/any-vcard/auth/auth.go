package auth

import (
	"context"
	"fmt"

	"github.com/rubiojr/any-vcard/cmd/any-vcard/util"
	"github.com/urfave/cli/v3"
)

var Command = &cli.Command{
	Name:  "auth",
	Usage: "Authenticate with Anytype to get an app key",
	Action: func(ctx context.Context, cmd *cli.Command) error {
		return authenticate(ctx, cmd.String("url"))
	},
}

func authenticate(ctx context.Context, baseURL string) error {
	client := util.NewClientWithURL(baseURL)

	fmt.Printf("Initiating authentication with %s...\n", baseURL)
	authResp, err := client.Auth().CreateChallenge(ctx, util.AppName)
	if err != nil {
		return fmt.Errorf("failed to create challenge: %w", err)
	}

	fmt.Printf("\nPlease enter the authentication code shown in Anytype:\n")
	fmt.Print("> ")

	var code string
	if _, err := fmt.Scanln(&code); err != nil {
		return fmt.Errorf("failed to read code: %w", err)
	}

	tokenResp, err := client.Auth().CreateApiKey(ctx, authResp.ChallengeID, code)
	if err != nil {
		return fmt.Errorf("authentication failed: %w", err)
	}

	fmt.Printf("\n✓ Authentication successful!\n")
	fmt.Printf("\nYour App Key:\n%s\n", tokenResp.ApiKey)
	fmt.Printf("\nSave this key and use it with --app-key flag or ANYTYPE_APP_KEY environment variable.\n")

	return nil
}
