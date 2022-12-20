package auth

import (
	"context"
	"fmt"

	"github.com/docker/docker-credential-helpers/client"
	"github.com/docker/docker-credential-helpers/credentials"
	"github.com/notaryproject/notation-go/config"
	"github.com/notaryproject/notation-go/log"
	"oras.land/oras-go/v2/registry/remote/auth"
)

const (
	remoteCredentialsPrefix = "docker-credential-"
	tokenUsername           = "<token>"
)

// var for unit testing.
var loadConfig = LoadConfig

// nativeAuthStore implements a credentials store using native keychain to keep
// credentials secure.
type nativeAuthStore struct {
	programFunc client.ProgramFunc
}

// GetCredentialsStore returns a new credentials store from the settings in the
// configuration file
func GetCredentialsStore(ctx context.Context, registryHostname string) (CredentialStore, error) {
	configFile, err := loadConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to load config file, error: %w", err)
	}
	if helper := getConfiguredCredentialStore(configFile, registryHostname); helper != "" {
		return newNativeAuthStore(ctx, helper), nil
	}
	return nil, fmt.Errorf("could not get the configured credentials store for registry: %s", registryHostname)
}

// newNativeAuthStore creates a new native store that uses a remote helper
// program to manage credentials. Note: it's different from the nativeStore in
// docker-cli which may fall back to plain text store
func newNativeAuthStore(ctx context.Context, helperSuffix string) CredentialStore {
	logger := log.GetLogger(ctx)
	name := remoteCredentialsPrefix + helperSuffix
	logger.Infoln("Executing remote credential helper program:", name)
	return &nativeAuthStore{
		programFunc: client.NewShellProgramFunc(name),
	}
}

// getConfiguredCredentialStore returns the credential helper configured for the
// given registry, the default credsStore, or the empty string if neither are
// configured.
func getConfiguredCredentialStore(c *config.Config, registryHostname string) string {
	if c.CredentialHelpers != nil && registryHostname != "" {
		if helper, exists := c.CredentialHelpers[registryHostname]; exists {
			return helper
		}
	}
	return c.CredentialsStore
}

// Store saves credentials into the native store
func (s *nativeAuthStore) Store(serverAddress string, authCreds auth.Credential) error {
	creds := &credentials.Credentials{
		ServerURL: serverAddress,
		Username:  authCreds.Username,
		Secret:    authCreds.Password,
	}

	if authCreds.RefreshToken != "" {
		creds.Username = tokenUsername
		creds.Secret = authCreds.RefreshToken
	}

	return client.Store(s.programFunc, creds)
}

// Get retrieves credentials from the store for the given server
func (s *nativeAuthStore) Get(serverAddress string) (auth.Credential, error) {
	creds, err := client.Get(s.programFunc, serverAddress)
	if err != nil {
		if credentials.IsErrCredentialsNotFound(err) {
			// do not return an error if the credentials are not in the keychain.
			return auth.EmptyCredential, nil
		}
		return auth.EmptyCredential, err
	}
	return newCredentialFromDockerCreds(creds), nil
}

// Erase removes credentials from the store for the given server
func (s *nativeAuthStore) Erase(serverAddress string) error {
	return client.Erase(s.programFunc, serverAddress)
}
