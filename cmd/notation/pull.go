package main

import (
	"fmt"

	"github.com/notaryproject/notation/pkg/cache"
	"github.com/opencontainers/go-digest"
	"github.com/spf13/cobra"
)

// TODO: This can be deprecated once the new verify command is merged into main.
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
