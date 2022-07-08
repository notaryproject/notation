package main

import (
	"context"
	"net"

	notationregistry "github.com/notaryproject/notation-go/registry"
	"github.com/notaryproject/notation/internal/version"
	loginauth "github.com/notaryproject/notation/pkg/auth"
	"github.com/notaryproject/notation/pkg/config"
	"github.com/urfave/cli/v2"
	"oras.land/oras-go/v2/registry"
	"oras.land/oras-go/v2/registry/remote"
	"oras.land/oras-go/v2/registry/remote/auth"
)

func getSignatureRepository(ctx *cli.Context, reference string) (notationregistry.SignatureRepository, error) {
	ref, err := registry.ParseReference(reference)
	if err != nil {
		return nil, err
	}
	return getRepositoryClient(ctx, ref)
}

func getRegistryClient(ctx *cli.Context, serverAddress string) (*remote.Registry, error) {
	reg, err := remote.NewRegistry(serverAddress)
	if err != nil {
		return nil, err
	}
	ref := registry.Reference{
		Registry: serverAddress,
	}
	reg.Client, reg.PlainHTTP, err = getAuthClient(ctx, ref)
	if err != nil {
		return nil, err
	}
	return reg, nil
}

func getRepositoryClient(ctx *cli.Context, ref registry.Reference) (*notationregistry.RepositoryClient, error) {
	authClient, plainHTTP, err := getAuthClient(ctx, ref)
	if err != nil {
		return nil, err
	}
	return notationregistry.NewRepositoryClient(authClient, ref, plainHTTP), nil
}

func getAuthClient(ctx *cli.Context, ref registry.Reference) (*auth.Client, bool, error) {
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
	if cred == auth.EmptyCredential {
		var err error
		if cred, err = getSavedCreds(ref.Registry); err != nil {
			return nil, false, err
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

	return authClient, plainHTTP, nil
}

func getSavedCreds(serverAddress string) (auth.Credential, error) {
	nativeStore, err := loginauth.GetCredentialsStore(serverAddress)
	if err != nil {
		return auth.EmptyCredential, err
	}

	return nativeStore.Get(serverAddress)
}
