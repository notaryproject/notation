package main

import "github.com/urfave/cli/v2"

var (
	usernameFlag = &cli.StringFlag{
		Name:    "username",
		Aliases: []string{"u"},
		Usage:   "username for generic remote access",
	}
	passwordFlag = &cli.StringFlag{
		Name:    "password",
		Aliases: []string{"p"},
		Usage:   "password for generic remote access",
	}
	insecureFlag = &cli.BoolFlag{
		Name:  "insecure",
		Usage: "enable insecure remote access",
	}
	mediaTypeFlag = &cli.StringFlag{
		Name:  "media-type",
		Usage: "specify the media type of the manifest read from file or stdin",
		Value: "application/vnd.docker.distribution.manifest.v2+json",
	}
)
