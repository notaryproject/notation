package main

import (
	"context"
	"errors"
	"fmt"

	notationregistry "github.com/notaryproject/notation-go/registry"
	"github.com/opencontainers/go-digest"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
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
	manifestDesc, _, err := getManifestDescriptor(command.Context(), &opts.SecureFlagOpts, reference)
	if err != nil {
		return err
	}

	// print all signature manifest digests
	return printSignatureManifestDigests(command.Context(), manifestDesc.Digest, sigRepo, reference)
}

// printSignatureManifestDigests returns the signature manifest digests of
// the subject manifest.
func printSignatureManifestDigests(ctx context.Context, manifestDigest digest.Digest, sigRepo notationregistry.Repository, reference string) error {
	// prepare title
	ref, err := registry.ParseReference(reference)
	if err != nil {
		return err
	}
	ref.Reference = manifestDigest.String()
	titlePrinted := false
	printTitle := func() {
		if !titlePrinted {
			fmt.Println(ref)
			fmt.Printf("└── %s\n", notationregistry.ArtifactTypeNotation)
			titlePrinted = true
		}
	}

	// traverse referrers
	artifactDescriptor, err := sigRepo.Resolve(ctx, reference)
	if err != nil {
		return err
	}
	var prevDigest digest.Digest
	err = sigRepo.ListSignatures(ctx, artifactDescriptor, func(signatureManifests []ocispec.Descriptor) error {
		for _, sigManifestDesc := range signatureManifests {
			if prevDigest != "" {
				// check and print title
				printTitle()

				// print each signature digest
				fmt.Printf("    ├── %s\n", prevDigest)
			}
			prevDigest = sigManifestDesc.Digest
		}
		return nil
	})

	if err != nil {
		return err
	}

	if prevDigest != "" {
		// check and print title
		printTitle()

		// print last signature digest
		fmt.Printf("    └── %s\n", prevDigest)
	}
	return nil
}
