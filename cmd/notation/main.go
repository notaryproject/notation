package main

import (
	"log"

	"github.com/notaryproject/notation/internal/version"
	"github.com/spf13/cobra"
)

func main() {
	cmd := &cobra.Command{
		Use:     "notation",
		Short:   "Notation - Notary V2",
		Version: version.GetVersion(),
	}
	cmd.AddCommand(
		signCommand(),
		verifyCommand(),
		pushCommand(),
		pullCommand(),
		listCommand(),
		certCommand(),
		keyCommand(),
		cacheCommand(),
		pluginCommand())
	if err := cmd.Execute(); err != nil {
		log.Fatal(err)
	}
}
