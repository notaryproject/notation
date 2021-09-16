package main

import (
	"encoding/json"
	"os"

	"github.com/notaryproject/notation/internal/docker"
	"github.com/notaryproject/notation/internal/version"
	"github.com/urfave/cli/v2"
)

var pluginMetadata = docker.PluginMetadata{
	SchemaVersion:    "0.1.0",
	Vendor:           "CNCF Notary Project",
	Version:          version.GetVersion(),
	ShortDescription: "Manage signatures on Docker images",
	URL:              "https://github.com/notaryproject/notation",
}

var metadataCommand = &cli.Command{
	Name: docker.PluginMetadataCommandName,
	Action: func(ctx *cli.Context) error {
		writer := json.NewEncoder(os.Stdout)
		return writer.Encode(pluginMetadata)
	},
	Hidden: true,
}
