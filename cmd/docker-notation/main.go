package main

import (
	"log"

	"github.com/spf13/cobra"
)

func main() {
	cmd := &cobra.Command{
		Use: "docker",
	}
	cmd.AddCommand(notationCommand(), metadataCommand())
	if err := cmd.Execute(); err != nil {
		log.Fatal(err)
	}
}
