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
)
