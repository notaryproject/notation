package main

import (
	"os"

	"github.com/notaryproject/notation/internal/version"
	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
)

func main() {
	app := &cli.App{
		Name:    "notation",
		Usage:   "Notation - Notary V2",
		Version: version.GetVersion(),
		Authors: []*cli.Author{
			{
				Name: "CNCF Notary Project",
			},
		},
		Commands: []*cli.Command{
			signCommand,
			verifyCommand,
			pushCommand,
			pullCommand,
			listCommand,
			certCommand,
			keyCommand,
			cacheCommand,
			pluginCommand,
		},
	}
	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
