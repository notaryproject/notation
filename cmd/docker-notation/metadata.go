package main

import (
	"encoding/json"
	"os"

	"github.com/notaryproject/notation/internal/docker"
	"github.com/notaryproject/notation/internal/version"
	"github.com/spf13/cobra"
)

var pluginMetadata = docker.PluginMetadata{
	SchemaVersion:    "0.1.0",
	Vendor:           "CNCF Notary Project",
	Version:          version.GetVersion(),
	ShortDescription: "Manage signatures on Docker images",
	URL:              "https://github.com/notaryproject/notation",
}

func metadataCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use: docker.PluginMetadataCommandName,
		RunE: func(cmd *cobra.Command, args []string) error {
			writer := json.NewEncoder(os.Stdout)
			return writer.Encode(pluginMetadata)
		},
		Hidden: true,
	}
	return cmd
}
