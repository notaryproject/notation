package docker

import (
	"context"
	"net"

	dockerconfig "github.com/docker/cli/cli/config"
	"github.com/docker/cli/cli/config/credentials"
	"github.com/notaryproject/notation/internal/version"
	"github.com/notaryproject/notation/pkg/config"
	notationregistry "github.com/notaryproject/notation/pkg/registry"
	"oras.land/oras-go/v2/registry"
	"oras.land/oras-go/v2/registry/remote/auth"
)

func getRepositoryClient(ref registry.Reference) (*notationregistry.RepositoryClient, error) {
	plainHTTP := config.IsRegistryInsecure(ref.Registry)
	if host, _, _ := net.SplitHostPort(ref.Registry); host == "localhost" {
		plainHTTP = true
	}

	cfg, err := dockerconfig.Load(dockerconfig.Dir())
	if err != nil {
		return nil, err
	}
	if !cfg.ContainsAuth() {
		cfg.CredentialsStore = credentials.DetectDefaultStore(cfg.CredentialsStore)
	}
	authConfig, err := cfg.GetAuthConfig(ref.Host())
	if err != nil {
		return nil, err
	}
	cred := auth.Credential{
		Username:     authConfig.Username,
		Password:     authConfig.Password,
		RefreshToken: authConfig.IdentityToken,
		AccessToken:  authConfig.RegistryToken,
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
		ClientID: "docker-notation",
	}
	authClient.SetUserAgent("docker-notation/" + version.GetVersion())

	return notationregistry.NewRepositoryClient(authClient, ref, plainHTTP), nil
}
