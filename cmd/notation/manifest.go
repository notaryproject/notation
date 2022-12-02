package main

import (
	"context"
	"errors"

	notationregistry "github.com/notaryproject/notation-go/registry"
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
	repository, err := getRepositoryClient(opts, ref)
	if err != nil {
		return ocispec.Descriptor{}, err
	}
	return notationregistry.NewRepository(repository).Resolve(ctx, ref.ReferenceOrDefault())
}
