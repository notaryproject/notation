package auth

import (
	"fmt"

	"github.com/docker/docker-credential-helpers/client"
	"github.com/docker/docker-credential-helpers/credentials"
	"github.com/notaryproject/notation/pkg/config"
	"oras.land/oras-go/v2/registry/remote/auth"
)

// nativeAuthStore implements a credentials store using native keychain to keep
// credentials secure.
type nativeAuthStore struct {
	programFunc client.ProgramFunc
}

// NewNativeAuthStore creates a new native store that uses a remote helper
// program to manage credentials. Note: it's different from the nativeStore in
// docker-cli which may fall back to plain text store
func NewNativeAuthStore(helperSuffix string) CredentialStore {
	name := remoteCredentialsPrefix + helperSuffix
	return &nativeAuthStore{
		programFunc: client.NewShellProgramFunc(name),
	}
}

// GetCredentialsStore returns a new credentials store from the settings in the
// configuration file
func GetCredentialsStore(registryHostname string) (CredentialStore, error) {
	configFile, err := LoadConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to load config file, error: %v", err)
	}
	if helper := getConfiguredCredentialStore(configFile, registryHostname); helper != "" {
		return newNativeStore(helper), nil
	}
	return nil, fmt.Errorf("could not get the configured credentials store for registry: %s", registryHostname)
}

// var for unit testing.
var newNativeStore = NewNativeAuthStore

// getConfiguredCredentialStore returns the credential helper configured for the
// given registry, the default credsStore, or the empty string if neither are
// configured.
func getConfiguredCredentialStore(c *config.File, registryHostname string) string {
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
