package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
	"unicode"

	"github.com/notaryproject/notation-go/log"
	notationregistry "github.com/notaryproject/notation-go/registry"
	notationerrors "github.com/notaryproject/notation/cmd/notation/internal/errors"
	"github.com/opencontainers/go-digest"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"oras.land/oras-go/v2/registry"
)

func resolveReferenceWithWarning(ctx context.Context, inputType inputType, reference string, sigRepo notationregistry.Repository, operation string) (ocispec.Descriptor, string, error) {
	return resolveReference(ctx, inputType, reference, sigRepo, func(ref string, manifestDesc ocispec.Descriptor) {
		fmt.Fprintf(os.Stderr, "Warning: Always %s the artifact using digest(@sha256:...) rather than a tag(:%s) because resolved digest may not point to the same signed artifact, as tags are mutable.\n", operation, ref)
	})
}

// resolveReference resolves user input reference based on user input type.
// Returns the resolved manifest descriptor and resolvedRef in digest
func resolveReference(ctx context.Context, inputType inputType, reference string, sigRepo notationregistry.Repository, fn func(string, ocispec.Descriptor)) (ocispec.Descriptor, string, error) {
	// sanity check
	if reference == "" {
		return ocispec.Descriptor{}, "", errors.New("missing user input reference")
	}
	var tagOrDigestRef string
	var resolvedRef string
	switch inputType {
	case inputTypeRegistry:
		ref, err := registry.ParseReference(reference)
		if err != nil {
			return ocispec.Descriptor{}, "", fmt.Errorf("failed to resolve user input reference: %w", err)
		}
		tagOrDigestRef = ref.Reference
		resolvedRef = ref.Registry + "/" + ref.Repository
	case inputTypeOCILayout:
		layoutPath, layoutReference, err := parseOCILayoutReference(reference)
		if err != nil {
			return ocispec.Descriptor{}, "", fmt.Errorf("failed to resolve user input reference: %w", err)
		}
		layoutPathInfo, err := os.Stat(layoutPath)
		if err != nil {
			return ocispec.Descriptor{}, "", fmt.Errorf("failed to resolve user input reference: %w", err)
		}
		if !layoutPathInfo.IsDir() {
			return ocispec.Descriptor{}, "", errors.New("failed to resolve user input reference: input path is not a dir")
		}
		tagOrDigestRef = layoutReference
		resolvedRef = layoutPath
	default:
		return ocispec.Descriptor{}, "", fmt.Errorf("unsupported user inputType: %d", inputType)
	}

	manifestDesc, err := getManifestDescriptor(ctx, tagOrDigestRef, sigRepo)
	if err != nil {
		return ocispec.Descriptor{}, "", fmt.Errorf("failed to get manifest descriptor: %w", err)
	}
	resolvedRef = resolvedRef + "@" + manifestDesc.Digest.String()
	if _, err := digest.Parse(tagOrDigestRef); err == nil {
		// tagOrDigestRef is a digest reference
		if tagOrDigestRef != manifestDesc.Digest.String() {
			// tagOrDigestRef does not match the resolved digest
			return ocispec.Descriptor{}, "", fmt.Errorf("user input digest %s does not match the resolved digest %s", tagOrDigestRef, manifestDesc.Digest.String())
		}
		return manifestDesc, resolvedRef, nil
	}
	// tagOrDigestRef is a tag reference
	if fn != nil {
		fn(tagOrDigestRef, manifestDesc)
	}
	return manifestDesc, resolvedRef, nil
}

// resolveArtifactDigestReference creates reference in Verification given user input
// trust policy scope
func resolveArtifactDigestReference(reference, policyScope string) string {
	if policyScope != "" {
		if _, digest, ok := strings.Cut(reference, "@"); ok {
			return policyScope + "@" + digest
		}
	}
	return reference
}

// parseOCILayoutReference parses the raw in format of <path>[:<tag>|@<digest>].
// Returns the path to the OCI layout and the reference (tag or digest).
func parseOCILayoutReference(raw string) (string, string, error) {
	var path string
	var ref string
	if idx := strings.LastIndex(raw, "@"); idx != -1 {
		// `digest` found
		path, ref = raw[:idx], raw[idx+1:]
	} else {
		// find `tag`
		idx := strings.LastIndex(raw, ":")
		if idx == -1 || (idx == 1 && len(raw) > 2 && unicode.IsLetter(rune(raw[0])) && raw[2] == '\\') {
			return "", "", notationerrors.ErrorOCILayoutMissingReference{}
		} else {
			path, ref = raw[:idx], raw[idx+1:]
		}
	}
	if path == "" {
		return "", "", fmt.Errorf("found empty file path in %q", raw)
	}
	if ref == "" {
		return "", "", fmt.Errorf("found empty reference in %q", raw)
	}
	return path, ref, nil
}

// getManifestDescriptor returns target artifact's manifest descriptor given
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
