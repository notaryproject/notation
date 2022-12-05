package main

import (
	"context"
	"errors"

	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"oras.land/oras-go/v2/registry"
)

// getManifestDescriptor returns target artifact manifest descriptor and
// registry.Reference given user input reference.
func getManifestDescriptor(ctx context.Context, opts *SecureFlagOpts, reference string) (ocispec.Descriptor, registry.Reference, error) {
	if reference == "" {
		return ocispec.Descriptor{}, registry.Reference{}, errors.New("missing reference")
	}
	ref, err := registry.ParseReference(reference)
	if err != nil {
		return ocispec.Descriptor{}, registry.Reference{}, err
	}
	if ref.Reference == "" {
		return ocispec.Descriptor{}, registry.Reference{}, errors.New("reference is missing digest or tag")
	}
	repo, err := getRepositoryClient(opts, ref)
	if err != nil {
		return ocispec.Descriptor{}, registry.Reference{}, err
	}

	manifestDesc, err := repo.Resolve(ctx, ref.Reference)
	if err != nil {
		return ocispec.Descriptor{}, registry.Reference{}, err
	}
	return manifestDesc, ref, nil
}
