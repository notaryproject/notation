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
		pushCommand(nil),
		pullCommand(nil),
		listCommand(nil),
		certCommand(),
		keyCommand(),
		policyCommand(),
		cacheCommand(),
		pluginCommand(),
		loginCommand(nil),
		logoutCommand(nil))
	cmd.PersistentFlags().Bool(flagPlainHTTP.Name, false, flagPlainHTTP.Usage)
	if err := cmd.Execute(); err != nil {
		log.Fatal(err)
	}
}
