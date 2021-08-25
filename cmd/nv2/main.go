package main

import (
	"log"
	"os"

	"github.com/urfave/cli/v2"
)

func main() {
	app := &cli.App{
		Name:    "nv2",
		Usage:   "Notary V2 - Prototype",
		Version: "0.4.0",
		Authors: []*cli.Author{
			{
				Name:  "Shiwei Zhang",
				Email: "shizh@microsoft.com",
			},
		},
		Commands: []*cli.Command{
			signCommand,
			verifyCommand,
			pushCommand,
			pullCommand,
			certificatesCommand,
		},
	}
	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
