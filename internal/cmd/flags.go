// Package cmd contains common flags and routines for all CLIs.
package cmd

import "github.com/urfave/cli/v2"

var (
	FlagKey = &cli.StringFlag{
		Name:    "key",
		Aliases: []string{"k"},
		Usage:   "signing key name",
	}

	FlagKeyFile = &cli.StringFlag{
		Name:      "key-file",
		Usage:     "signing key file",
		TakesFile: true,
	}

	FlagCertFile = &cli.StringFlag{
		Name:      "cert-file",
		Usage:     "signing certificate file",
		TakesFile: true,
	}

	FlagTimestamp = &cli.StringFlag{
		Name:    "timestamp",
		Aliases: []string{"t"},
		Usage:   "timestamp the signed signature via the remote TSA",
	}

	FlagExpiry = &cli.DurationFlag{
		Name:    "expiry",
		Aliases: []string{"e"},
		Usage:   "expire duration",
	}

	FlagReference = &cli.StringFlag{
		Name:    "reference",
		Aliases: []string{"r"},
		Usage:   "original reference",
	}
)
