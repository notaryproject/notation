package main

import (
	"log"

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
		certCommand(),
		keyCommand(),
		pluginCommand(),
		loginCommand(nil),
		logoutCommand(nil),
		versionCommand(),
	)
	cmd.PersistentFlags().Bool(flagPlainHTTP.Name, false, flagPlainHTTP.Usage)
	if err := cmd.Execute(); err != nil {
		log.Fatal(err)
	}
}
