package main

import "github.com/urfave/cli/v2"

var (
	flagUsername = &cli.StringFlag{
		Name:    "username",
		Aliases: []string{"u"},
		Usage:   "Username for registry operations",
		EnvVars: []string{"NOTATION_USERNAME"},
	}
	flagPassword = &cli.StringFlag{
		Name:    "password",
		Aliases: []string{"p"},
		Usage:   "Password for registry operations",
		EnvVars: []string{"NOTATION_PASSWORD"},
	}
	flagPlainHTTP = &cli.BoolFlag{
		Name:  "plain-http",
		Usage: "Registry access via plain HTTP",
	}
	flagMediaType = &cli.StringFlag{
		Name:  "media-type",
		Usage: "specify the media type of the manifest read from file or stdin",
		Value: "application/vnd.docker.distribution.manifest.v2+json",
	}
	flagOutput = &cli.StringFlag{
		Name:    "output",
		Aliases: []string{"o"},
		Usage:   "write signature to a specific path",
	}
	flagLocal = &cli.BoolFlag{
		Name:    "local",
		Aliases: []string{"l"},
		Usage:   "reference is a local file",
	}
	flagSignature = &cli.StringSliceFlag{
		Name:      "signature",
		Aliases:   []string{"s", "f"},
		Usage:     "signature files",
		TakesFile: true,
	}
)
