package main

import (
	"encoding/json"
	"os"

	"github.com/notaryproject/notation/internal/docker"
	"github.com/spf13/cobra"
)

func metadataCommand() *cobra.Command {
	return &cobra.Command{
		Use: docker.PluginMetadataCommandName,
		RunE: func(cmd *cobra.Command, args []string) error {
			writer := json.NewEncoder(os.Stdout)
			return writer.Encode(pluginMetadata)
		},
		Hidden: true,
	}
}

var pluginMetadata = docker.PluginMetadata{
	SchemaVersion:    "0.1.0",
	Vendor:           "CNCF Notary Project",
	Version:          "0.1.1",
	ShortDescription: "Generate artifacts",
	URL:              "https://github.com/notaryproject/notation",
	Experimental:     true,
}
