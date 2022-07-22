package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

type listOpts struct {
	SecureFlagOpts
	reference string
}

func listCommand(opts *listOpts) *cobra.Command {
	if opts == nil {
		opts = &listOpts{}
	}
	cmd := &cobra.Command{
		Use:     "list [reference]",
		Aliases: []string{"ls"},
		Short:   "List signatures from remote",
		Args:    cobra.ExactArgs(1),
		PreRun: func(cmd *cobra.Command, args []string) {
			opts.reference = args[0]
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return runList(cmd, opts)
		},
	}
	opts.ApplyFlags(cmd.Flags())
	return cmd
}

func runList(command *cobra.Command, opts *listOpts) error {
	// initialize
	reference := opts.reference
	sigRepo, err := getSignatureRepository(&opts.SecureFlagOpts, reference)
	if err != nil {
		return err
	}

	// core process
	manifestDesc, err := getManifestDescriptorFromReference(command.Context(), &opts.SecureFlagOpts, reference)
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
