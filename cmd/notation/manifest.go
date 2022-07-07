package main

import (
	"errors"
	"io"
	"math"
	"os"

	"github.com/notaryproject/notation-go"
	"github.com/opencontainers/go-digest"
	"github.com/spf13/cobra"
	"oras.land/oras-go/v2/registry"
)

func getManifestDescriptorFromContext(cmd *cobra.Command) (notation.Descriptor, error) {
	ref := cmd.Flags().Arg(0)
	if ref == "" {
		return notation.Descriptor{}, errors.New("missing reference")
	}
	return getManifestDescriptorFromContextWithReference(cmd, ref)
}

func getManifestDescriptorFromContextWithReference(cmd *cobra.Command, ref string) (notation.Descriptor, error) {
	if isLocal, _ := cmd.Flags().GetBool(flagLocal.Name); isLocal {
		mediaType, _ := cmd.Flags().GetString(flagMediaType.Name)
		if ref == "-" {
			return getManifestDescriptorFromReader(os.Stdin, mediaType)
		}
		return getManifestDescriptorFromFile(ref, mediaType)
	}

	return getManifestDescriptorFromReference(cmd, ref)
}

func getManifestDescriptorFromReference(cmd *cobra.Command, reference string) (notation.Descriptor, error) {
	ref, err := registry.ParseReference(reference)
	if err != nil {
		return notation.Descriptor{}, err
	}
	repo := getRepositoryClient(cmd, ref)
	return repo.Resolve(cmd.Context(), ref.ReferenceOrDefault())
}

func getManifestDescriptorFromFile(path, mediaType string) (notation.Descriptor, error) {
	file, err := os.Open(path)
	if err != nil {
		return notation.Descriptor{}, err
	}
	defer file.Close()
	return getManifestDescriptorFromReader(file, mediaType)
}

func getManifestDescriptorFromReader(r io.Reader, mediaType string) (notation.Descriptor, error) {
	lr := &io.LimitedReader{
		R: r,
		N: math.MaxInt64,
	}
	digest, err := digest.SHA256.FromReader(lr)
	if err != nil {
		return notation.Descriptor{}, err
	}
	return notation.Descriptor{
		MediaType: mediaType,
		Digest:    digest,
		Size:      math.MaxInt64 - lr.N,
	}, nil
}
