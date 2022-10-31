package main

import (
	"errors"
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
		Use:     "list [flags] <reference>",
		Aliases: []string{"ls"},
		Short:   "List signatures of the signed artifact",
		Long:    "List all the signatures associated with signed artifact",
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return errors.New("no reference specified")
			}
			opts.reference = args[0]
			return nil
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
