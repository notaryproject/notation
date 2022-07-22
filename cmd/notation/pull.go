package main

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"

	notationregistry "github.com/notaryproject/notation-go/registry"
	"github.com/notaryproject/notation/internal/osutil"
	"github.com/notaryproject/notation/pkg/cache"
	"github.com/notaryproject/notation/pkg/config"
	"github.com/opencontainers/go-digest"
	"github.com/spf13/cobra"
	"oras.land/oras-go/v2/registry"
)

type pullOpts struct {
	SecureFlagOpts
	strict    bool
	reference string
	output    string
}

func pullCommand(opts *pullOpts) *cobra.Command {
	if opts == nil {
		opts = &pullOpts{}
	}
	cmd := &cobra.Command{
		Use:   "pull [reference]",
		Short: "Pull signatures from remote",
		Args:  cobra.ExactArgs(1),
		PreRun: func(cmd *cobra.Command, args []string) {
			opts.reference = args[0]
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return runPull(cmd, opts)
		},
	}
	cmd.Flags().BoolVar(&opts.strict, "strict", false, "pull the signature without lookup the manifest")
	setFlagOutput(cmd.Flags(), &opts.output)
	opts.ApplyFlags(cmd.Flags())
	return cmd
}

func runPull(command *cobra.Command, opts *pullOpts) error {
	// initialize
	if opts.reference == "" {
		return errors.New("no reference specified")
	}

	reference := opts.reference
	sigRepo, err := getSignatureRepository(&opts.SecureFlagOpts, reference)
	if err != nil {
		return err
	}

	// core process
	if opts.strict {
		return pullSignatureStrict(command.Context(), opts, sigRepo, reference)
	}

	manifestDesc, err := getManifestDescriptorFromReference(command.Context(), &opts.SecureFlagOpts, reference)
	if err != nil {
		return err
	}

	sigManifests, err := sigRepo.ListSignatureManifests(command.Context(), manifestDesc.Digest)
	if err != nil {
		return fmt.Errorf("list signature manifests failure: %v", err)
	}

	path := opts.output
	for _, sigManifest := range sigManifests {
		sigDigest := sigManifest.Blob.Digest
		if path != "" {
			outputPath := filepath.Join(path, sigDigest.Encoded()+config.SignatureExtension)
			sig, err := sigRepo.Get(command.Context(), sigDigest)
			if err != nil {
				return fmt.Errorf("get signature failure: %v: %v", sigDigest, err)
			}
			if err := osutil.WriteFile(outputPath, sig); err != nil {
				return fmt.Errorf("fail to write signature: %v: %v", sigDigest, err)
			}
		} else if err := cache.PullSignature(command.Context(), sigRepo, manifestDesc.Digest, sigDigest); err != nil {
			return err
		}

		// write out
		fmt.Println(sigDigest)
	}

	return nil
}

func pullSignatureStrict(ctx context.Context, opts *pullOpts, sigRepo notationregistry.SignatureRepository, reference string) error {
	ref, err := registry.ParseReference(reference)
	if err != nil {
		return err
	}
	sigDigest, err := ref.Digest()
	if err != nil {
		return fmt.Errorf("invalid signature digest: %v", err)
	}

	sig, err := sigRepo.Get(ctx, sigDigest)
	if err != nil {
		return fmt.Errorf("get signature failure: %v: %v", sigDigest, err)
	}
	outputPath := opts.output
	if outputPath == "" {
		outputPath = sigDigest.Encoded() + config.SignatureExtension
	}
	if err := osutil.WriteFile(outputPath, sig); err != nil {
		return fmt.Errorf("fail to write signature: %v: %v", sigDigest, err)
	}

	// write out
	fmt.Println(sigDigest)
	return nil
}

func pullSignatures(command *cobra.Command, reference string, opts *SecureFlagOpts, manifestDigest digest.Digest) error {
	sigRepo, err := getSignatureRepository(opts, reference)
	if err != nil {
		return err
	}

	sigManifests, err := sigRepo.ListSignatureManifests(command.Context(), manifestDigest)
	if err != nil {
		return fmt.Errorf("lookup signature failure: %v", err)
	}
	for _, sigManifest := range sigManifests {
		if err := cache.PullSignature(command.Context(), sigRepo, manifestDigest, sigManifest.Blob.Digest); err != nil {
			return err
		}
	}
	return nil
}
