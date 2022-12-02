package main

import (
	"context"
	"errors"

	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"oras.land/oras-go/v2/registry"
)

func getManifestDescriptorFromContext(ctx context.Context, opts *SecureFlagOpts, ref string, debug bool) (ocispec.Descriptor, error) {
	if ref == "" {
		return ocispec.Descriptor{}, errors.New("missing reference")
	}

	return getManifestDescriptorFromReference(ctx, opts, ref, debug)
}

func getManifestDescriptorFromReference(ctx context.Context, opts *SecureFlagOpts, reference string, debug bool) (ocispec.Descriptor, error) {
	ref, err := registry.ParseReference(reference)
	if err != nil {
		return ocispec.Descriptor{}, err
	}
	repo, err := getRepositoryClient(opts, ref, debug)
	if err != nil {
		return ocispec.Descriptor{}, err
	}
	return repo.Resolve(ctx, ref.ReferenceOrDefault())
}
