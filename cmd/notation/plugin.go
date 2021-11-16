package main

import (
	"errors"
	"fmt"
	"path/filepath"

	"github.com/notaryproject/notation/pkg/config"
	"github.com/urfave/cli/v2"
)

var (
	pluginCommand = &cli.Command{
		Name:  "plugin",
		Usage: "Manage KMS plugins",
		Subcommands: []*cli.Command{
			pluginAddCommand,
			pluginListCommand,
			pluginRemoveCommand,
		},
	}

	pluginDefaultFlag = &cli.BoolFlag{
		Name:    "default",
		Aliases: []string{"d"},
		Usage:   "mark as default",
	}

	pluginAddCommand = &cli.Command{
		Name:      "add",
		Usage:     "Register a plugin",
		ArgsUsage: "<plugin-name> <plugin-path>",
		Flags: []cli.Flag{
			pluginDefaultFlag,
		},
		Action: addPlugin,
	}

	pluginListCommand = &cli.Command{
		Name:    "list",
		Usage:   "List registered plugins",
		Aliases: []string{"ls"},
		Action:  listPlugins,
	}

	pluginRemoveCommand = &cli.Command{
		Name:      "remove",
		Usage:     "Remove a plugin",
		Aliases:   []string{"rm"},
		ArgsUsage: "<plugin-name> ...",
		Action:    removePlugins,
	}
)

func addPlugin(ctx *cli.Context) error {
	// initialize
	args := ctx.Args()
	switch args.Len() {
	case 0:
		return errors.New("missing plugin name and path")
	case 1:
		return errors.New("missing plugin path for the correspoding plugin name")
	}

	pluginPath, err := filepath.Abs(args.Get(1))
	if err != nil {
		return err
	}
	pluginName := args.Get(0)

	// core process
	cfg, err := config.LoadOrDefault()
	if err != nil {
		return err
	}
	if err := addPluginCore(cfg, pluginName, pluginPath); err != nil {
		return err
	}
	if err := cfg.Save(); err != nil {
		return err
	}

	// write out
	fmt.Printf("plugin %s added\n", pluginName)

	return nil
}

func addPluginCore(cfg *config.File, pluginName, pluginPath string) error {
	// Should we run discover plugin here?
	if ok := cfg.KMSPlugins.Plugins.Append(pluginName, pluginPath); !ok {
		return errors.New(pluginName + ": already exists")
	}
	return nil
}

func listPlugins(ctx *cli.Context) error {
	// core process
	cfg, err := config.LoadOrDefault()
	if err != nil {
		return err
	}

	// write out
	printPluginSet(cfg.KMSPlugins.Plugins)
	return nil
}

func removePlugins(ctx *cli.Context) error {
	// initialize
	names := ctx.Args().Slice()
	if len(names) == 0 {
		return errors.New("missing plugin names")
	}

	// core process
	cfg, err := config.LoadOrDefault()
	if err != nil {
		return err
	}

	var removedNames []string
	for _, name := range names {
		if ok := cfg.KMSPlugins.Plugins.Remove(name); !ok {
			return errors.New(name + ": not found")
		}
		removedNames = append(removedNames, name)
	}
	if err := cfg.Save(); err != nil {
		return err
	}

	// write out
	for _, name := range removedNames {
		fmt.Println(name)
	}

	return nil
}

func printPluginSet(s config.PluginMap) {
	maxNameSize := 0
	for _, ref := range s {
		if len(ref.Name) > maxNameSize {
			maxNameSize = len(ref.Name)
		}
	}
	format := fmt.Sprintf("%%-%ds\t%%s\n", maxNameSize)
	fmt.Printf(format, "NAME", "PATH")
	for _, ref := range s {
		fmt.Printf(format, ref.Name, ref.Path)
	}
}
