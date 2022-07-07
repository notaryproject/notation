package main

import (
	"github.com/spf13/cobra"
)

func notationCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "notation",
		Short: pluginMetadata.ShortDescription,
	}
	cmd.AddCommand(pullCommand(), pushCommand(), signCommand())
	return cmd
}
