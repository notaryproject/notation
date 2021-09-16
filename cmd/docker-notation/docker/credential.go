package docker

import (
	"github.com/docker/cli/cli/config"
	"github.com/docker/cli/cli/config/credentials"
)

// BasicCredentialFromDockerConfig fetches the credentials for basic auth
// from docker config
func BasicCredentialFromDockerConfig(hostname string) (string, string, error) {
	cfg, err := config.Load(config.Dir())
	if err != nil {
		return "", "", err
	}
	if !cfg.ContainsAuth() {
		cfg.CredentialsStore = credentials.DetectDefaultStore(cfg.CredentialsStore)
	}

	auth, err := cfg.GetAuthConfig(hostname)
	if err != nil {
		return "", "", err
	}
	return auth.Username, auth.Password, nil
}
