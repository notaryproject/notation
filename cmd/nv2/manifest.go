package main

import (
	"fmt"
	"io"
	"math"
	"net/url"
	"os"
	"strings"

	"github.com/notaryproject/notary/v2/signature"
	"github.com/notaryproject/nv2/pkg/registry"
	"github.com/opencontainers/go-digest"
	"github.com/urfave/cli/v2"
)

func getManifestFromContext(ctx *cli.Context) (signature.Manifest, error) {
	if uri := ctx.Args().First(); uri != "" {
		parsed, err := url.Parse(uri)
		if err != nil {
			return signature.Manifest{}, err
		}
		return getManfestsFromURI(ctx, parsed)
	}
	return getManifestFromReader(os.Stdin, ctx.String(mediaTypeFlag.Name))
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

func getManfestsFromURI(ctx *cli.Context, uri *url.URL) (signature.Manifest, error) {
	var r io.Reader
	switch strings.ToLower(uri.Scheme) {
	case "file":
		path := uri.Path
		if uri.Opaque != "" {
			path = uri.Opaque
		}
		file, err := os.Open(path)
		if err != nil {
			return signature.Manifest{}, err
		}
		defer file.Close()
		r = file
	case "docker", "oci":
		remote := registry.NewClient(
			registry.NewAuthtransport(
				nil,
				ctx.String(usernameFlag.Name),
				ctx.String(passwordFlag.Name),
			),
			ctx.Bool(plainHTTPFlag.Name),
		)
		return remote.GetManifestMetadata(uri)
	default:
		return signature.Manifest{}, fmt.Errorf("unsupported URI scheme: %s", uri.Scheme)
	}
	return getManifestFromReader(r, ctx.String(mediaTypeFlag.Name))
}
