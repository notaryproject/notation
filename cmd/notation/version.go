package main

import (
	"fmt"
	"runtime"

	"github.com/notaryproject/notation/internal/version"
	"github.com/spf13/cobra"
)

func versionCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "version",
		Short: "Show the notation version information",
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			runVersion()
		},
	}
	return cmd
}

func runVersion() {
	fmt.Printf("Notation: Notary v2, A tool to sign, store, and verify artifacts.\n\n")

	fmt.Printf("Version:     %s\n", version.GetVersion())
	fmt.Printf("Go version:  %s\n", runtime.Version())

	if version.GitCommit != "" {
		fmt.Printf("Git commit:  %s\n", version.GitCommit)
	}
}
