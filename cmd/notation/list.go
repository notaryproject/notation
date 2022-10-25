package main

import (
	"errors"
	"fmt"

	notationRegistry "github.com/notaryproject/notation-go/registry"
	"github.com/opencontainers/go-digest"
	"github.com/spf13/cobra"
	"oras.land/oras-go/v2/registry"
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
		Short:   "List signatures from remote",
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
	return output(manifestDesc.Digest, sigManifests, reference)
}

func output(digest digest.Digest, sigManifests []notationRegistry.SignatureManifest, reference string) error {
	ref, err := registry.ParseReference(reference)
	if err != nil {
		return err
	}

	sigCount := len(sigManifests)
	if sigCount > 0 {
		// print title
		fmt.Printf("%s/%s@%s\n", ref.Registry, ref.Repository, digest)
		fmt.Printf("└── %s\n", notationRegistry.ArtifactTypeNotation)

		for _, sigManifest := range sigManifests[:sigCount-1] {
			// print each signature digest
			fmt.Printf("    ├── %s\n", sigManifest.Blob.Digest)
		}
		fmt.Printf("    └── %s\n", sigManifests[sigCount-1].Blob.Digest)
	}
	return nil
}
