package main

import (
	"log"

	"github.com/notaryproject/notation/internal/version"
	"github.com/spf13/cobra"
)

func main() {
	cmd := &cobra.Command{
		Use:          "notation",
		Short:        "Notation - Notary V2",
		Version:      version.GetVersion(),
		SilenceUsage: true,
	}
	cmd.AddCommand(
		signCommand(nil),
		verifyCommand(nil),
		listCommand(nil),
		certCommand(),
		keyCommand(),
		pluginCommand(),
		loginCommand(nil),
		logoutCommand(nil))

	if err := cmd.Execute(); err != nil {
		log.Fatal(err)
	}
}
