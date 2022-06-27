package auth

import (
	"fmt"

	"github.com/docker/docker-credential-helpers/credentials"
	"github.com/notaryproject/notation/pkg/config"
	"oras.land/oras-go/v2/registry/remote/auth"
)

// LoadConfig loads the configuration from the config file
func LoadConfig() (*config.File, error) {
	// load notation config first
	config, err := config.LoadOrDefault()
	if err != nil {
		return nil, err
	}
	if config != nil && containsAuth(config) {
		fmt.Println("Using notation config file")
		return config, nil
	}

	config, err = loadDockerConfig()
	if err != nil {
		return nil, err
	}
	if config != nil && containsAuth(config) {
		fmt.Println("Using docker config file")
		return config, nil
	}
	if !containsAuth(config) {
		return nil, fmt.Errorf("credentials store config is not set up")
	}
	return config, nil
}

// loadDockerConfig loads the configuration from the config file under .docker
// directory
func loadDockerConfig() (*config.File, error) {
	dockerConfig, err := config.LoadDockerConfig()
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
