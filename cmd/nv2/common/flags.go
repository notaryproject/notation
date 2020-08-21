package common

import "github.com/urfave/cli/v2"

// Common flags
var (
	UsernameFlag = &cli.StringFlag{
		Name:    "username",
		Aliases: []string{"u"},
		Usage:   "username for generic remote access",
	}
	PasswordFlag = &cli.StringFlag{
		Name:    "password",
		Aliases: []string{"p"},
		Usage:   "password for generic remote access",
	}
	InsecureFlag = &cli.BoolFlag{
		Name:  "insecure",
		Usage: "enable insecure remote access",
	}
	MediaTypeFlag = &cli.StringFlag{
		Name:  "media-type",
		Usage: "specify the media type of the manifest read from file or stdin",
		Value: "application/vnd.docker.distribution.manifest.v2+json",
	}
	ExpiryFlag = &cli.DurationFlag{
		Name:    "expiry",
		Aliases: []string{"e"},
		Usage:   "expire duration",
	}
	OutputFlag = &cli.StringFlag{
		Name:    "output",
		Aliases: []string{"o"},
		Usage:   "write signature to a specific path",
	}
)
