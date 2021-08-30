package main

import (
	"os"

	"github.com/notaryproject/notation/pkg/config"
	"github.com/urfave/cli/v2"
)

var notationCommand = &cli.Command{
	Name:  "notation",
	Usage: pluginMetadata.ShortDescription,
	Subcommands: []*cli.Command{
		pullCommand,
		pushCommand,
		signCommand,
	},
	Flags: []cli.Flag{
		notationEnabledFlag,
	},
	Action: setNotation,
}

var notationEnabledFlag = &cli.BoolFlag{
	Name:  "enabled",
	Usage: "Enable Notation support",
}

func setNotation(ctx *cli.Context) error {
	if !ctx.IsSet(notationEnabledFlag.Name) {
		return cli.ShowCommandHelp(ctx, ctx.Command.Name)
	}

	cfg, err := config.Load()
	if err != nil {
		if !os.IsNotExist(err) {
			return err
		}
		cfg = config.New()
	}
	cfg.Enabled = ctx.Bool(notationEnabledFlag.Name)
	return cfg.Save()
}
