package main

import (
	"github.com/notaryproject/notation/pkg/config"
	"github.com/notaryproject/notation/pkg/registry"
	"github.com/urfave/cli/v2"
)

func getSignatureRepository(ctx *cli.Context, reference string) (registry.SignatureRepository, error) {
	ref, err := registry.ParseReference(reference)
	if err != nil {
		return nil, err
	}
	plainHTTP := ctx.Bool(flagPlainHTTP.Name)
	if !plainHTTP {
		plainHTTP = config.IsRegistryInsecure(ref.Registry)
	}
	remote := registry.NewClient(
		registry.NewAuthtransport(
			nil,
			ctx.String(flagUsername.Name),
			ctx.String(flagPassword.Name),
		),
		ref.Registry,
		plainHTTP,
	)
	return remote.Repository(ctx.Context, ref.Repository), nil
}
