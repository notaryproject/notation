package main

import (
	"log"
	"os"

	"github.com/notaryproject/nv2/cmd/nv2/signature"
	"github.com/notaryproject/nv2/cmd/nv2/tuf"
	"github.com/urfave/cli/v2"
)

func main() {
	app := &cli.App{
		Name:    "nv2",
		Usage:   "Notary V2 - Prototype",
		Version: "0.3.1",
		Authors: []*cli.Author{
			{
				Name:  "Shiwei Zhang",
				Email: "shizh@microsoft.com",
			},
		},
		Commands: []*cli.Command{
			signature.SignCommand,
			signature.VerifyCommand,
			tuf.TUFCommand,
		},
	}
	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
