package main

import (
	"context"
	"os"

	"github.com/notaryproject/notation-go/plugin/manager"
	"github.com/notaryproject/notation/internal/ioutil"
	"github.com/urfave/cli/v2"
)

var (
	pluginCommand = &cli.Command{
		Name:    "plugin",
		Aliases: []string{"ls"},
		Usage:   "Manage plugins",
		Subcommands: []*cli.Command{
			pluginListCommand,
		},
	}

	pluginListCommand = &cli.Command{
		Name:   "list",
		Usage:  "List registered plugins",
		Action: listPlugins,
	}
)

func listPlugins(ctx *cli.Context) error {
	mgr := manager.NewManager()
	plugins, err := mgr.List(ctx.Context)
	if err != nil {
		return err
	}
	return ioutil.PrintPlugins(os.Stdout, plugins)
}
