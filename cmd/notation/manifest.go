package main

import (
	"errors"
	"io"
	"math"
	"os"

	"github.com/notaryproject/notation-go-lib/signature"
	"github.com/notaryproject/notation/pkg/config"
	"github.com/notaryproject/notation/pkg/registry"
	"github.com/opencontainers/go-digest"
	"github.com/urfave/cli/v2"
)

func getManifestFromContext(ctx *cli.Context) (signature.Manifest, error) {
	ref := ctx.Args().First()
	if ref == "" {
		return signature.Manifest{}, errors.New("missing reference")
	}
	return getManifestFromContextWithReference(ctx, ref)
}

func getManifestFromContextWithReference(ctx *cli.Context, ref string) (signature.Manifest, error) {
	if ctx.Bool(localFlag.Name) {
		mediaType := ctx.String(mediaTypeFlag.Name)
		if ref == "-" {
			return getManifestFromReader(os.Stdin, mediaType)
		}
		return getManifestsFromFile(ref, mediaType)
	}

	return getManifestsFromReference(ctx, ref)
}

func getManifestsFromReference(ctx *cli.Context, reference string) (signature.Manifest, error) {
	ref, err := registry.ParseReference(reference)
	if err != nil {
		return signature.Manifest{}, err
	}
	plainHTTP := ctx.Bool(plainHTTPFlag.Name)
	if !plainHTTP {
		plainHTTP = config.IsRegistryInsecure(ref.Registry)
	}
	remote := registry.NewClient(
		registry.NewAuthtransport(
			nil,
			ctx.String(usernameFlag.Name),
			ctx.String(passwordFlag.Name),
		),
		plainHTTP,
	)
	return remote.GetManifestMetadata(ref)
}

func getManifestsFromFile(path, mediaType string) (signature.Manifest, error) {
	file, err := os.Open(path)
	if err != nil {
		return signature.Manifest{}, err
	}
	defer file.Close()
	return getManifestFromReader(file, mediaType)
}

func getManifestFromReader(r io.Reader, mediaType string) (signature.Manifest, error) {
	lr := &io.LimitedReader{
		R: r,
		N: math.MaxInt64,
	}
	digest, err := digest.SHA256.FromReader(lr)
	if err != nil {
		return signature.Manifest{}, err
	}
	return signature.Manifest{
		Descriptor: signature.Descriptor{
			MediaType: mediaType,
			Digest:    digest.String(),
			Size:      math.MaxInt64 - lr.N,
		},
	}, nil
}
