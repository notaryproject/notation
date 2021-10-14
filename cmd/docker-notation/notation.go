package main

import (
	"github.com/urfave/cli/v2"
)

var notationCommand = &cli.Command{
	Name:  "notation",
	Usage: pluginMetadata.ShortDescription,
	Subcommands: []*cli.Command{
		pullCommand,
		pushCommand,
		signCommand,
	},
}
