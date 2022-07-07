package main

import (
	"errors"
	"fmt"

	"github.com/spf13/cobra"
)

func listCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "list [reference]",
		Aliases: []string{"ls"},
		Short:   "List signatures from remote",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runList(cmd)
		},
	}
	setFlagUserName(cmd)
	setFlagPassword(cmd)
	setFlagPlainHTTP(cmd)
	return cmd
}

func runList(command *cobra.Command) error {
	// initialize
	if command.Flags().NArg() == 0 {
		return errors.New("no reference specified")
	}

	reference := command.Flags().Arg(0)
	sigRepo, err := getSignatureRepository(command, reference)
	if err != nil {
		return err
	}

	// core process
	manifestDesc, err := getManifestDescriptorFromReference(command, reference)
	if err != nil {
		return err
	}

	sigManifests, err := sigRepo.ListSignatureManifests(command.Context(), manifestDesc.Digest)
	if err != nil {
		return fmt.Errorf("lookup signature failure: %v", err)
	}

	// write out
	for _, sigManifest := range sigManifests {
		fmt.Println(sigManifest.Blob.Digest)
	}

	return nil
}
