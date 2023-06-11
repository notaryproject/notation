package auth

import (
	"fmt"

	"github.com/notaryproject/notation-go/dir"
	credentials "github.com/oras-project/oras-credentials-go"
)

// NewCredentialsStore returns a new credentials store from the settings in the
// configuration file.
func NewCredentialsStore() (credentials.Store, error) {
	configPath, err := dir.ConfigFS().SysPath(dir.PathConfigFile)
	if err != nil {
		return nil, fmt.Errorf("failed to load config file: %w", err)
	}

	// use notation config
	opts := credentials.StoreOptions{AllowPlaintextPut: false}
	notationStore, err := credentials.NewStore(configPath, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to create credential store from config file: %w", err)
	}
	if notationStore.IsAuthConfigured() {
		return notationStore, nil
	}

	// use docker config
	dockerStore, err := credentials.NewStoreFromDocker(opts)
	if err != nil {
		return nil, fmt.Errorf("failed to create credential store from docker config file: %w", err)
	}
	if dockerStore.IsAuthConfigured() {
		return dockerStore, nil
	}

	// detect platform-default native store
	if osDefaultStore, ok := credentials.NewDefaultNativeStore(); ok {
		return osDefaultStore, nil
	}
	// if the default store is not available, still use notation store so that
	// there won't be errors when getting credentials
	return notationStore, nil
}
