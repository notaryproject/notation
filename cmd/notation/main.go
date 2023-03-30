package main

import (
	"os"

	"github.com/notaryproject/notation/cmd/notation/cert"
	"github.com/notaryproject/notation/cmd/notation/policy"
	"github.com/spf13/cobra"
)

func main() {
	cmd := &cobra.Command{
		Use:          "notation",
		Short:        "Notation - a tool to sign and verify artifacts",
		SilenceUsage: true,
	}
	cmd.AddCommand(
		signCommand(nil),
		verifyCommand(nil),
		listCommand(nil),
		cert.Cmd(),
		policy.Cmd(),
		keyCommand(),
		pluginCommand(),
		loginCommand(nil),
		logoutCommand(nil),
		versionCommand(),
		inspectCommand(nil),
	)
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
