package main

import (
	"github.com/urfave/cli/v2"
)

var generateCommand = &cli.Command{
	Name: "generate",
	Subcommands: []*cli.Command{
		manifestCommand,
	},
}
