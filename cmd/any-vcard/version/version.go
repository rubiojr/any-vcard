package version

import (
	"context"
	"fmt"

	"github.com/rubiojr/any-vcard/cmd/any-vcard/util"
	"github.com/urfave/cli/v3"
)

var Command = &cli.Command{
	Name:  "version",
	Usage: "Print the version",
	Action: func(ctx context.Context, cmd *cli.Command) error {
		fmt.Printf("%s version %s\n", util.AppName, util.Version)
		return nil
	},
}
