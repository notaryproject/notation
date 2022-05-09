package main

import (
	"os"

	"github.com/notaryproject/notation-go/plugin/manager"
	"github.com/notaryproject/notation/internal/ioutil"
	"github.com/notaryproject/notation/pkg/config"
	"github.com/urfave/cli/v2"
)

var (
	pluginCommand = &cli.Command{
		Name:  "plugin",
		Usage: "Manage plugins",
		Subcommands: []*cli.Command{
			pluginListCommand,
		},
	}

	pluginListCommand = &cli.Command{
		Name:    "list",
		Usage:   "List registered plugins",
		Aliases: []string{"ls"},
		Action:  listPlugins,
	}
)

func listPlugins(ctx *cli.Context) error {
	mgr := manager.New(config.PluginDirPath)
	plugins, err := mgr.List(ctx.Context)
	if err != nil {
		return err
	}
	return ioutil.PrintPlugins(os.Stdout, plugins)
}
