package main

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/notaryproject/notary/v2"
	notaryregistry "github.com/notaryproject/notary/v2/registry"
	"github.com/notaryproject/nv2/pkg/registry"
	"github.com/urfave/cli/v2"
)

func getSignatureRepositoryFromURI(ctx *cli.Context, uri *url.URL) (notary.SignatureRepository, error) {
	switch strings.ToLower(uri.Scheme) {
	case "docker", "oci":
		ref := registry.ParseReferenceFromURL(uri)
		remote := notaryregistry.NewClient(
			registry.NewAuthtransport(
				nil,
				ctx.String(usernameFlag.Name),
				ctx.String(passwordFlag.Name),
			),
			ref.Registry,
			ctx.Bool(plainHTTPFlag.Name),
		)
		return remote.Repository(ctx.Context, ref.Repository), nil
	default:
		return nil, fmt.Errorf("unsupported URI scheme: %s", uri.Scheme)
	}
}
