package main

import (
	"errors"
	"io"
	"math"
	"os"

	"github.com/notaryproject/notation-go"
	"github.com/opencontainers/go-digest"
	"github.com/urfave/cli/v2"
	"oras.land/oras-go/v2/registry"
)

func getManifestDescriptorFromContext(ctx *cli.Context) (notation.Descriptor, error) {
	ref := ctx.Args().First()
	if ref == "" {
		return notation.Descriptor{}, errors.New("missing reference")
	}
	return getManifestDescriptorFromContextWithReference(ctx, ref)
}

func getManifestDescriptorFromContextWithReference(ctx *cli.Context, ref string) (notation.Descriptor, error) {
	if ctx.Bool(flagLocal.Name) {
		mediaType := ctx.String(flagMediaType.Name)
		if ref == "-" {
			return getManifestDescriptorFromReader(os.Stdin, mediaType)
		}
		return getManifestDescriptorFromFile(ref, mediaType)
	}

	return getManifestDescriptorFromReference(ctx, ref)
}

func getManifestDescriptorFromReference(ctx *cli.Context, reference string) (notation.Descriptor, error) {
	ref, err := registry.ParseReference(reference)
	if err != nil {
		return notation.Descriptor{}, err
	}
	repo := getRepositoryClient(ctx, ref)
	return repo.GetManifestDescriptor(ctx.Context, ref.ReferenceOrDefault())
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
