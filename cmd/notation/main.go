package main

import (
	"os"

	"github.com/notaryproject/notation/cmd/notation/cert"
	"github.com/spf13/cobra"
)

func main() {
	cmd := &cobra.Command{
		Use:          "notation",
		Short:        "Notation - Notary V2",
		SilenceUsage: true,
	}
	cmd.AddCommand(
		signCommand(nil),
		verifyCommand(nil),
		listCommand(nil),
		cert.Cmd(),
		keyCommand(),
		pluginCommand(),
		loginCommand(nil),
		logoutCommand(nil),
		versionCommand(),
	)
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
