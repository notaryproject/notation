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

	manifestDesc, err := sigRepo.Resolve(ctx, ref.Reference)
	if err != nil {
		return ocispec.Descriptor{}, registry.Reference{}, err
	}

	logger.Infof("Reference resolved to manifest descriptor: {MediaType:%v, Digest:%v, Size:%v}", manifestDesc.MediaType, manifestDesc.Digest, manifestDesc.Size)
	return manifestDesc, ref, nil
}
