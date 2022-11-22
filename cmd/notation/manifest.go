package main

import (
	"context"
	"errors"

	"github.com/notaryproject/notation-go"
	"oras.land/oras-go/v2/registry"
)

func getManifestDescriptorFromContext(ctx context.Context, opts *SecureFlagOpts, ref string) (notation.Descriptor, registry.Reference, error) {
	if ref == "" {
		return notation.Descriptor{}, registry.Reference{}, errors.New("missing reference")
	}

	return getManifestDescriptorFromReference(ctx, opts, ref)
}

func getManifestDescriptorFromReference(ctx context.Context, opts *SecureFlagOpts, reference string) (notation.Descriptor, registry.Reference, error) {
	ref, err := registry.ParseReference(reference)
	if err != nil {
		return notation.Descriptor{}, registry.Reference{}, err
	}
	repo, err := getRepositoryClient(opts, ref)
	if err != nil {
		return notation.Descriptor{}, registry.Reference{}, err
	}
	manifestDesc, err := repo.Resolve(ctx, ref.ReferenceOrDefault())
	return manifestDesc, ref, err
}
