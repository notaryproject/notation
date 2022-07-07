package main

import (
	"context"
	"net"

	notationregistry "github.com/notaryproject/notation-go/registry"
	"github.com/notaryproject/notation/internal/version"
	"github.com/notaryproject/notation/pkg/config"
	"github.com/spf13/cobra"
	"oras.land/oras-go/v2/registry"
	"oras.land/oras-go/v2/registry/remote/auth"
)

func getSignatureRepository(cmd *cobra.Command, reference string) (notationregistry.SignatureRepository, error) {
	ref, err := registry.ParseReference(reference)
	if err != nil {
		return nil, err
	}
	return getRepositoryClient(cmd, ref), nil
}

func getRepositoryClient(cmd *cobra.Command, ref registry.Reference) *notationregistry.RepositoryClient {
	var plainHTTP bool

	if cmd.Flags().Lookup(flagPlainHTTP.Name) != nil {
		plainHTTP, _ = cmd.Flags().GetBool(flagPlainHTTP.Name)
	} else {
		plainHTTP = config.IsRegistryInsecure(ref.Registry)
		if !plainHTTP {
			if host, _, _ := net.SplitHostPort(ref.Registry); host == "localhost" {
				plainHTTP = true
			}
		}
	}
	username, _ := cmd.Flags().GetString(flagUsername.Name)
	password, _ := cmd.Flags().GetString(flagPassword.Name)
	cred := auth.Credential{
		Username: username,
		Password: password,
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
