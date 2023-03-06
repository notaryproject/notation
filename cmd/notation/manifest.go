package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"

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

func getManifestDescriptorFromOCILayout(ctx context.Context, reference string, sigRepo notationregistry.Repository) (ocispec.Descriptor, error) {
	logger := log.GetLogger(ctx)

	if err := validateOciStore(sigRepo); err != nil {
		return ocispec.Descriptor{}, err
	}
	manifestDesc, err := sigRepo.Resolve(ctx, reference)
	if err != nil {
		return ocispec.Descriptor{}, err
	}
	fmt.Printf("%+v\n", manifestDesc)

	logger.Infof("Reference %s resolved to manifest descriptor: %+v", reference, manifestDesc)
	return manifestDesc, nil
}

func getManifestDescriptorFromFile(path string) (ocispec.Descriptor, error) {
	root, err := filepath.Abs(path)
	if err != nil {
		return ocispec.Descriptor{}, err
	}
	file, err := os.ReadFile(root)
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
