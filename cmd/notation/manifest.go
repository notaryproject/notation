package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"

	"github.com/notaryproject/notation-go/log"
	notationregistry "github.com/notaryproject/notation-go/registry"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"oras.land/oras-go/v2/content/oci"
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

	logger.Infof("Reference %s resolved to manifest descriptor: %+v", ref.Reference, manifestDesc)
	return manifestDesc, ref, nil
}

// getManifestDescriptorFromOCILayout returns target artifact manifest
// descriptor given OCI layout reference.
// layoutReference is a valid tag or digest in the OCI layout
// sigRepo should be oci.Store for an OCI layout folder
func getManifestDescriptorFromOCILayout(ctx context.Context, layoutReference string, sigRepo notationregistry.Repository) (ocispec.Descriptor, error) {
	logger := log.GetLogger(ctx)

	if err := validateOciStore(sigRepo); err != nil {
		return ocispec.Descriptor{}, err
	}
	manifestDesc, err := sigRepo.Resolve(ctx, layoutReference)
	if err != nil {
		return ocispec.Descriptor{}, err
	}

	logger.Infof("Reference %s resolved to manifest descriptor: %+v", layoutReference, manifestDesc)
	return manifestDesc, nil
}

// getManifestDescriptorFromFile parses target artifact manifest
// descriptor from path of a local descriptor json file.
func getManifestDescriptorFromFile(path string) (ocispec.Descriptor, error) {
	file, err := os.ReadFile(path)
	if err != nil {
		return ocispec.Descriptor{}, err
	}
	var targetDesc ocispec.Descriptor
	err = json.Unmarshal(file, &targetDesc)
	if err != nil {
		return ocispec.Descriptor{}, fmt.Errorf("reading a descriptor from JSON file: %w", err)
	}
	return plain(targetDesc), nil
}

// validateOciStore validates if repo is an oci.Store
func validateOciStore(repo notationregistry.Repository) error {
	switch r := repo.(type) {
	case *notationregistry.RepositoryClient:
		switch r.Target.(type) {
		case *oci.Store:
			return nil
		default:
			return fmt.Errorf("repo is not an oci.Store")
		}
	default:
		return nil
	}
}

// plain returns a plain descriptor that contains only MediaType, Digest and
// Size.
func plain(desc ocispec.Descriptor) ocispec.Descriptor {
	return ocispec.Descriptor{
		MediaType: desc.MediaType,
		Digest:    desc.Digest,
		Size:      desc.Size,
	}
}
