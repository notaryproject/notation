package main

import (
	"context"
	"errors"

	"github.com/notaryproject/notation-go"
	"oras.land/oras-go/v2/registry"
)

func getManifestDescriptorFromContext(ctx context.Context, opts *SecureFlagOpts, ref string) (notation.Descriptor, error) {
	if ref == "" {
		return notation.Descriptor{}, errors.New("missing reference")
	}

	return getManifestDescriptorFromReference(ctx, opts, ref)
}

func getManifestDescriptorFromReference(ctx context.Context, opts *SecureFlagOpts, reference string) (notation.Descriptor, error) {
	ref, err := registry.ParseReference(reference)
	if err != nil {
		return notation.Descriptor{}, err
	}
	repo, err := getRepositoryClient(opts, ref)
	if err != nil {
		return notation.Descriptor{}, err
	}
	return repo.Resolve(ctx, ref.ReferenceOrDefault())
}
