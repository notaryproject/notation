package main

import (
	"os"

	"github.com/notaryproject/notation-go/plugin/manager"
	"github.com/notaryproject/notation/internal/ioutil"
	"github.com/spf13/cobra"
)

func pluginCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "plugin",
		Short: "Manage plugins",
	}
	cmd.AddCommand(pluginListCommand())
	return cmd
}

func pluginListCommand() *cobra.Command {
	return &cobra.Command{
		Use:     "list [flags]",
		Aliases: []string{"ls"},
		Short:   "List installed plugins",
		Long: `List installed plugins

Example - List installed Notation plugins:
  notation plugin list
`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return listPlugins(cmd)
		},
	}
}

func listPlugins(command *cobra.Command) error {
	mgr := manager.New()
	plugins, err := mgr.List(command.Context())
	if err != nil {
		return err
	}
	return ioutil.PrintPlugins(os.Stdout, plugins)
}
