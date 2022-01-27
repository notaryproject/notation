package main

import (
	"context"
	"net"

	"github.com/notaryproject/notation/internal/version"
	"github.com/notaryproject/notation/pkg/config"
	notationregistry "github.com/notaryproject/notation/pkg/registry"
	"github.com/urfave/cli/v2"
	"oras.land/oras-go/v2/registry"
	"oras.land/oras-go/v2/registry/remote/auth"
)

func getSignatureRepository(ctx *cli.Context, reference string) (notationregistry.SignatureRepository, error) {
	ref, err := registry.ParseReference(reference)
	if err != nil {
		return nil, err
	}
	return getRepositoryClient(ctx, ref), nil
}

func getRepositoryClient(ctx *cli.Context, ref registry.Reference) *notationregistry.RepositoryClient {
	var plainHTTP bool
	if ctx.IsSet(flagPlainHTTP.Name) {
		plainHTTP = ctx.Bool(flagPlainHTTP.Name)
	} else {
		plainHTTP = config.IsRegistryInsecure(ref.Registry)
		if !plainHTTP {
			if host, _, _ := net.SplitHostPort(ref.Registry); host == "localhost" {
				plainHTTP = true
			}
		}
	}

	cred := auth.Credential{
		Username: ctx.String(flagUsername.Name),
		Password: ctx.String(flagPassword.Name),
	}
	if cred.Username == "" {
		cred = auth.Credential{
			RefreshToken: cred.Password,
		}
	}
	authClient := &auth.Client{
		Credential: func(ctx context.Context, registry string) (auth.Credential, error) {
			switch registry {
			case ref.Host():
				return cred, nil
			default:
				return auth.EmptyCredential, nil
			}
		},
		Cache:    auth.NewCache(),
		ClientID: "notation",
	}
	authClient.SetUserAgent("notation/" + version.GetVersion())

	return notationregistry.NewRepositoryClient(authClient, ref, plainHTTP)
}
