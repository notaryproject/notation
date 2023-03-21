package main

import (
	"context"
	"errors"

	"github.com/notaryproject/notation-go/log"
	notationregistry "github.com/notaryproject/notation-go/registry"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"oras.land/oras-go/v2/registry"
)

// getManifestDescriptor returns target artifact manifest descriptor and
// registry.Reference given user input reference.
func getManifestDescriptor(ctx context.Context, opts *SecureFlagOpts, reference string, sigRepo notationregistry.Repository) (ocispec.Descriptor, registry.Reference, error) {
	logger := log.GetLogger(ctx)

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

	manifestDesc, err := sigRepo.Resolve(ctx, ref.String())
	if err != nil {
		return ocispec.Descriptor{}, registry.Reference{}, err
	}

	logger.Infof("Reference %s resolved to manifest descriptor: %+v", ref.Reference, manifestDesc)
	return manifestDesc, ref, nil
}

func resolveReference(ctx context.Context, opts *SecureFlagOpts, reference string, sigRepo notationregistry.Repository, fn func(registry.Reference, ocispec.Descriptor)) (registry.Reference, error) {
	manifestDesc, ref, err := getManifestDescriptor(ctx, opts, reference, sigRepo)
	if err != nil {
		return registry.Reference{}, err
	}

	// reference is a digest reference
	if err := ref.ValidateReferenceAsDigest(); err == nil {
		return ref, nil
	}

	// reference is a tag reference
	fn(ref, manifestDesc)
	// resolve tag to digest reference
	ref.Reference = manifestDesc.Digest.String()

	return ref, nil
}
