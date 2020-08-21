package common

import (
	"fmt"
	"io"
	"math"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/notaryproject/nv2/pkg/reference"
	"github.com/notaryproject/nv2/pkg/registry"
	"github.com/opencontainers/go-digest"
	"github.com/urfave/cli/v2"
)

// GetManifestFromContext reterives the manifest according to CLI context
func GetManifestFromContext(ctx *cli.Context) (*reference.Manifest, error) {
	if uri := ctx.Args().First(); uri != "" {
		return getManfestsFromURI(ctx, uri)
	}
	return getManifestFromReader(os.Stdin, ctx.String(MediaTypeFlag.Name))
}

func getManifestFromReader(r io.Reader, mediaType string) (*reference.Manifest, error) {
	lr := &io.LimitedReader{
		R: r,
		N: math.MaxInt64,
	}
	manifestDigest, err := digest.SHA256.FromReader(lr)
	if err != nil {
		return nil, err
	}
	return &reference.Manifest{
		Descriptor: reference.Descriptor{
			MediaType: mediaType,
			Digests:   []digest.Digest{manifestDigest},
			Size:      math.MaxInt64 - lr.N,
		},
		AccessedAt: time.Now().UTC(),
	}, nil
}

func getManfestsFromURI(ctx *cli.Context, uri string) (*reference.Manifest, error) {
	parsed, err := url.Parse(uri)
	if err != nil {
		return nil, err
	}
	var r io.Reader
	switch strings.ToLower(parsed.Scheme) {
	case "file":
		path := parsed.Path
		if parsed.Opaque != "" {
			path = parsed.Opaque
		}
		file, err := os.Open(path)
		if err != nil {
			return nil, err
		}
		defer file.Close()
		r = file
	case "docker", "oci":
		remote := registry.NewClient(nil, &registry.ClientOptions{
			Username: ctx.String(UsernameFlag.Name),
			Password: ctx.String(PasswordFlag.Name),
			Insecure: ctx.Bool(InsecureFlag.Name),
		})
		return remote.GetManifestMetadata(parsed)
	default:
		return nil, fmt.Errorf("unsupported URI scheme: %s", parsed.Scheme)
	}
	return getManifestFromReader(r, ctx.String(MediaTypeFlag.Name))
}
