package main

import (
	"github.com/notaryproject/notation-go-lib"
	registryn "github.com/notaryproject/notation-go-lib/registry"
	"github.com/notaryproject/notation/pkg/registry"
	"github.com/urfave/cli/v2"
)

func getSignatureRepository(ctx *cli.Context, reference string) (notation.SignatureRepository, error) {
	ref := registry.ParseReference(reference)
	remote := registryn.NewClient(
		registry.NewAuthtransport(
			nil,
			ctx.String(usernameFlag.Name),
			ctx.String(passwordFlag.Name),
		),
		ref.Registry,
		ctx.Bool(plainHTTPFlag.Name),
	)
	return remote.Repository(ctx.Context, ref.Repository), nil
}
