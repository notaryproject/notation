package main

import (
	"context"
	"errors"
	"net"

	notationregistry "github.com/notaryproject/notation-go/registry"
	"github.com/notaryproject/notation/internal/version"
	loginauth "github.com/notaryproject/notation/pkg/auth"
	"github.com/notaryproject/notation/pkg/configutil"
	"oras.land/oras-go/v2/registry"
	"oras.land/oras-go/v2/registry/remote"
	"oras.land/oras-go/v2/registry/remote/auth"
)

func getSignatureRepository(opts *SecureFlagOpts, reference string) (notationregistry.Repository, error) {
	ref, err := registry.ParseReference(reference)
	if err != nil {
		return nil, err
	}
	return getRepositoryClient(opts, ref)
}

func getRegistryClient(opts *SecureFlagOpts, serverAddress string) (*remote.Registry, error) {
	reg, err := remote.NewRegistry(serverAddress)
	if err != nil {
		return nil, err
	}

	reg.Client, reg.PlainHTTP, err = getAuthClient(opts, reg.Reference)
	if err != nil {
		return nil, err
	}
	return reg, nil
}

func getRepositoryClient(opts *SecureFlagOpts, ref registry.Reference) (notationregistry.Repository, error) {
	authClient, plainHTTP, err := getAuthClient(opts, ref)
	if err != nil {
		return nil, err
	}
	repo := &remote.Repository{
		Client:    authClient,
		Reference: ref,
		PlainHTTP: plainHTTP,
	}

	return notationregistry.NewRepository(repo), nil
}

func getAuthClient(opts *SecureFlagOpts, ref registry.Reference) (*auth.Client, bool, error) {
	var plainHTTP bool

	if opts.PlainHTTP {
		plainHTTP = opts.PlainHTTP
	} else {
		plainHTTP = configutil.IsRegistryInsecure(ref.Registry)
		if !plainHTTP {
			if host, _, _ := net.SplitHostPort(ref.Registry); host == "localhost" {
				plainHTTP = true
			}
		}
	}
	cred := auth.Credential{
		Username: opts.Username,
		Password: opts.Password,
	}
	if cred.Username == "" {
		cred = auth.Credential{
			RefreshToken: cred.Password,
		}
	}
	if cred == auth.EmptyCredential {
		var err error
		cred, err = getSavedCreds(ref.Registry)
		// local registry may not need credentials
		if err != nil && !errors.Is(err, loginauth.ErrCredentialsConfigNotSet) {
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
