package auth

import "oras.land/oras-go/v2/registry/remote/auth"

// CredentialStore is the interface that any credentials store must implement.
type CredentialStore interface {
	// Store saves credentials into the store
	Store(serverAddress string, credsConf auth.Credential) error
	// Erase removes credentials from the store for the given server
	Erase(serverAddress string) error
	// Get retrieves credentials from the store for the given server
	Get(serverAddress string) (auth.Credential, error)
}
