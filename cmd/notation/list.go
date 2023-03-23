package main

import (
	"context"
	"errors"
	"fmt"

	notationregistry "github.com/notaryproject/notation-go/registry"
	"github.com/notaryproject/notation/internal/cmd"
	"github.com/opencontainers/go-digest"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/spf13/cobra"
)

type listOpts struct {
	cmd.LoggingFlagOpts
	SecureFlagOpts
	reference    string
	localContent bool
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
			return runList(cmd.Context(), opts)
		},
	}
	opts.LoggingFlagOpts.ApplyFlags(cmd.Flags())
	opts.SecureFlagOpts.ApplyFlags(cmd.Flags())
	cmd.Flags().BoolVar(&opts.localContent, "local-content", false, "list signatures from OCI layout")
	return cmd
}

func runList(ctx context.Context, opts *listOpts) error {
	// set log level
	ctx = opts.LoggingFlagOpts.SetLoggerLevel(ctx)

	// initialize
	reference := opts.reference
	if opts.localContent {
		var layout ociLayout
		var err error
		layout.path, layout.reference, err = parseOCILayoutReference(opts.reference)
		if err != nil {
			return err
		}
		sigRepo, err := ociLayoutRepository(layout.path)
		if err != nil {
			return err
		}
		targetDesc, err := sigRepo.Resolve(ctx, layout.reference)
		if err != nil {
			return fmt.Errorf("failed to resolve OCI layout reference: %s", err)
		}
		printOut := layout.path + "@" + targetDesc.Digest.String()
		// print all signature manifest digests
		return printSignatureManifestDigests(ctx, targetDesc, sigRepo, printOut)
	}
	sigRepo, err := getSignatureRepository(ctx, &opts.SecureFlagOpts, reference)
	if err != nil {
		return err
	}
	targetDesc, ref, err := getManifestDescriptor(ctx, &opts.SecureFlagOpts, reference, sigRepo)
	if err != nil {
		return err
	}
	ref.Reference = targetDesc.Digest.String()
	// print all signature manifest digests
	return printSignatureManifestDigests(ctx, targetDesc, sigRepo, ref.String())
}

// printSignatureManifestDigests returns the signature manifest digests of
// the subject manifest.
func printSignatureManifestDigests(ctx context.Context, targetDesc ocispec.Descriptor, sigRepo notationregistry.Repository, ref string) error {
	titlePrinted := false
	printTitle := func() {
		if !titlePrinted {
			fmt.Println(ref)
			fmt.Printf("└── %s\n", notationregistry.ArtifactTypeNotation)
			titlePrinted = true
		}
	}

	var prevDigest digest.Digest
	err := sigRepo.ListSignatures(ctx, targetDesc, func(signatureManifests []ocispec.Descriptor) error {
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
