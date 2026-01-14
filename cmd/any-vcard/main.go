package main

import (
	"context"
	"log"
	"os"

	"github.com/rubiojr/any-vcard/cmd/any-vcard/auth"
	"github.com/rubiojr/any-vcard/cmd/any-vcard/diff"
	vcardimport "github.com/rubiojr/any-vcard/cmd/any-vcard/import"
	"github.com/rubiojr/any-vcard/cmd/any-vcard/space"
	"github.com/rubiojr/any-vcard/cmd/any-vcard/template"
	"github.com/rubiojr/any-vcard/cmd/any-vcard/types"
	"github.com/rubiojr/any-vcard/cmd/any-vcard/util"
	"github.com/rubiojr/any-vcard/cmd/any-vcard/version"
	"github.com/urfave/cli/v3"
)

func main() {
	cmd := &cli.Command{
		Name:    util.AppName,
		Usage:   "Import vCard files into Anytype",
		Version: util.Version,
		Flags:   util.GlobalFlags(),
		Commands: []*cli.Command{
			auth.Command,
			diff.Command,
			vcardimport.Command,
			space.Command,
			template.Command,
			types.Command,
			version.Command,
		},
	}

	if err := cmd.Run(context.Background(), os.Args); err != nil {
		log.Fatal(err)
	}
}
