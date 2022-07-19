package main

import (
	"log"

	"github.com/spf13/cobra"
)

func main() {
	cmd := &cobra.Command{
		Use: "docker",
	}
	cmd.AddCommand(generateCommand(), metadataCommand())
	if err := cmd.Execute(); err != nil {
		log.Fatal(err)
	}
}
