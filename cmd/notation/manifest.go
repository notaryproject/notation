package main

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/notaryproject/notation-go/log"
	notationregistry "github.com/notaryproject/notation-go/registry"
	"github.com/opencontainers/go-digest"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"oras.land/oras-go/v2/registry"
)

// resolveReference resolves user input reference based on user input type
func resolveReference(ctx context.Context, inputType inputType, reference string, sigRepo notationregistry.Repository, fn func(string, ocispec.Descriptor)) (ocispec.Descriptor, string, error) {
	if reference == "" {
		return ocispec.Descriptor{}, "", errors.New("missing user input reference")
	}
	var tagOrDigestRef string
	var fullRef string
	switch inputType {
	case remoteRegistry:
		ref, err := registry.ParseReference(reference)
		if err != nil {
			return ocispec.Descriptor{}, "", fmt.Errorf("failed to resolve user input reference: %w", err)
		}
		tagOrDigestRef = ref.Reference
		fullRef = ref.Registry + "/" + ref.Repository
	case ociLayout:
		layoutPath, layoutReference, err := parseOCILayoutReference(reference)
		if err != nil {
			return ocispec.Descriptor{}, "", fmt.Errorf("failed to resolve user input reference: %w", err)
		}
		tagOrDigestRef = layoutReference
		fullRef = localTargetPath(layoutPath)
	default:
		return ocispec.Descriptor{}, "", errors.New("unsupported user input type")
	}

	manifestDesc, err := getManifestDescriptor(ctx, tagOrDigestRef, sigRepo)
	if err != nil {
		return ocispec.Descriptor{}, "", fmt.Errorf("failed to get manifest descriptor: %w", err)
	}
	fullRef = fullRef + "@" + manifestDesc.Digest.String()
	if _, err := digest.Parse(tagOrDigestRef); err == nil {
		// tagOrDigestRef is a digest reference
		return manifestDesc, fullRef, nil
	}
	// tagOrDigestRef is a tag reference
	fn(tagOrDigestRef, manifestDesc)
	return manifestDesc, fullRef, nil
}

// getManifestDescriptor returns target artifact manifest descriptor given
// reference (digest or tag) and Repository.
func getManifestDescriptor(ctx context.Context, reference string, sigRepo notationregistry.Repository) (ocispec.Descriptor, error) {
	logger := log.GetLogger(ctx)

	if reference == "" {
		return ocispec.Descriptor{}, errors.New("reference cannot be empty")
	}
	manifestDesc, err := sigRepo.Resolve(ctx, reference)
	if err != nil {
		return ocispec.Descriptor{}, err
	}
	logger.Infof("Reference %s resolved to manifest descriptor: %+v", reference, manifestDesc)
	return manifestDesc, nil
}

func localTargetPath(path string) string {
	reg := strings.ToLower(filepath.Base(filepath.Dir(path)))
	repo := strings.ToLower(filepath.Base(path))
	return fmt.Sprintf("%s/%s", reg, repo)
}
