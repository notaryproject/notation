package tuf

import "github.com/urfave/cli/v2"

// TUFCommand contains the TUF related commands
var TUFCommand = &cli.Command{
	Name:  "tuf",
	Usage: "TUF related commands",
	Subcommands: []*cli.Command{
		SignCommand,
		VerifyCommand,
	},
}
