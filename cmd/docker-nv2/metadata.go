package main

import (
	"encoding/json"
	"os"

	"github.com/urfave/cli/v2"
)

var pluginMetadata = map[string]interface{}{
	"SchemaVersion":    "0.1.0",
	"Vendor":           "Sajay Antony, Shiwei Zhang",
	"Version":          "0.2.2",
	"ShortDescription": "Notary V2 Signature extension",
	"URL":              "https://github.com/notaryproject/nv2",
}

var metadataCommand = &cli.Command{
	Name: "docker-cli-plugin-metadata",
	Action: func(ctx *cli.Context) error {
		writer := json.NewEncoder(os.Stdout)
		return writer.Encode(pluginMetadata)
	},
	Hidden: true,
}
