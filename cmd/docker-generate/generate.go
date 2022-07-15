package main

import (
	"github.com/spf13/cobra"
)

func generateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use: "generate",
	}
	cmd.AddCommand(generateManifestCommand())
	return cmd
}
