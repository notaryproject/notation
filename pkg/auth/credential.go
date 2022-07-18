package auth

import (
	"fmt"

	"github.com/docker/docker-credential-helpers/credentials"
	"github.com/notaryproject/notation/pkg/config"
	"oras.land/oras-go/v2/registry/remote/auth"
)

// var for unit tests
var (
	loadOrDefault    = config.LoadOrDefault
	loadDockerConfig = config.LoadDockerConfig
)

// LoadConfig loads the configuration from the config file
func LoadConfig() (*config.File, error) {
	// load notation config first
	config, err := loadOrDefault()
	if err != nil {
		return nil, err
	}
	if config != nil && containsAuth(config) {
		return config, nil
	}

	config, err = loadDockerCredentials()
	if err != nil {
		return nil, err
	}
	if containsAuth(config) {
		return config, nil
	}
	return nil, fmt.Errorf("credentials store config is not set up")
}

// loadDockerCredentials loads the configuration from the config file under .docker
// directory
func loadDockerCredentials() (*config.File, error) {
	dockerConfig, err := loadDockerConfig()
	if err != nil {
		return nil, err
	}
	return &config.File{
		CredentialHelpers: dockerConfig.CredentialHelpers,
		CredentialsStore:  dockerConfig.CredentialsStore,
	}, nil
}

// containsAuth returns whether there is authentication configured in this file
// or not.
func containsAuth(configFile *config.File) bool {
	return configFile.CredentialsStore != "" || len(configFile.CredentialHelpers) > 0
}

// newCredentialFromDockerCreds creates a new auth.Credential from the docker-cli credentials
func newCredentialFromDockerCreds(dockerCreds *credentials.Credentials) auth.Credential {
	var credsConf auth.Credential
	if dockerCreds.Username == tokenUsername {
		credsConf.RefreshToken = dockerCreds.Secret
	} else {
		credsConf.Password = dockerCreds.Secret
		credsConf.Username = dockerCreds.Username
	}
	return credsConf
}
