package main

import (
	"os"

	"github.com/notaryproject/notation-go/plugin"
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
	mgr, err := plugin.NewManager()
	if err != nil {
		return err
	}
	plugins, err := mgr.List()
	if err != nil {
		return err
	}
	return ioutil.PrintPlugins(os.Stdout, plugins)
}
