package main

import (
	"fmt"
	"io"
	"math"
	"net/url"
	"os"

	"github.com/notaryproject/nv2/pkg/signature"
	"github.com/opencontainers/go-digest"
	"github.com/urfave/cli/v2"
)

func getManifestFromContext(ctx *cli.Context) (signature.Manifest, error) {
	if uri := ctx.Args().First(); uri != "" {
		return getManfestsFromURI(uri)
	}
	return getManifestFromReader(os.Stdin)
}

func getManifestFromReader(r io.Reader) (signature.Manifest, error) {
	lr := &io.LimitedReader{
		R: r,
		N: math.MaxInt64,
	}
	digest, err := digest.SHA256.FromReader(lr)
	if err != nil {
		return signature.Manifest{}, err
	}
	return signature.Manifest{
		Digest: digest.String(),
		Size:   math.MaxInt64 - lr.N,
	}, nil
}

func getManfestsFromURI(uri string) (signature.Manifest, error) {
	parsed, err := url.Parse(uri)
	if err != nil {
		return signature.Manifest{}, err
	}
	var r io.Reader
	switch parsed.Scheme {
	case "file":
		path := parsed.Path
		if parsed.Opaque != "" {
			path = parsed.Opaque
		}
		file, err := os.Open(path)
		if err != nil {
			return signature.Manifest{}, err
		}
		defer file.Close()
		r = file
	default:
		return signature.Manifest{}, fmt.Errorf("unsupported URI scheme: %s", parsed.Scheme)
	}
	return getManifestFromReader(r)
}
