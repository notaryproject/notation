package main

import (
	"context"
	"errors"

	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"oras.land/oras-go/v2/registry"
)

func getManifestDescriptorFromContext(ctx context.Context, opts *SecureFlagOpts, ref string) (ocispec.Descriptor, registry.Reference, error) {
	if ref == "" {
		return ocispec.Descriptor{}, registry.Reference{}, errors.New("missing reference")
	}

	return getManifestDescriptorFromReference(ctx, opts, ref)
}

func getManifestDescriptorFromReference(ctx context.Context, opts *SecureFlagOpts, reference string) (ocispec.Descriptor, registry.Reference, error) {
	ref, err := registry.ParseReference(reference)
	if err != nil {
		return ocispec.Descriptor{}, registry.Reference{}, err
	}
	repo, err := getRepositoryClient(opts, ref)
	if err != nil {
		return ocispec.Descriptor{}, registry.Reference{}, err
	}
	manifestDesc, err := repo.Resolve(ctx, ref.ReferenceOrDefault())
	return manifestDesc, ref, err
}
