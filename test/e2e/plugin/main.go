package main

import (
	"encoding/json"
	"os"

	"github.com/spf13/cobra"
)

func main() {
	cmd := &cobra.Command{
		Use:           "plugin for Notation E2E test",
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	cmd.AddCommand(
		getPluginMetadataCommand(),
		describeKeyCommand(),
		generateSignatureCommand(),
		generateEnvelopeCommand(),
		verifySignatureCommand(),
	)

	if err := cmd.Execute(); err != nil {
		if newErr := json.NewEncoder(os.Stderr).Encode(err); newErr != nil {
			panic(newErr)
		}
		os.Exit(1)
	}
}
