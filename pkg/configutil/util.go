package configutil

import (
	"errors"
	"strings"

	"github.com/notaryproject/notation-go/config"
)

var (
	// ErrKeyNotFound indicates that the signing key is not found.
	ErrKeyNotFound = errors.New("signing key not found")
)

// IsRegistryInsecure checks whether the registry is in the list of insecure registries.
func IsRegistryInsecure(target string) bool {
	config, err := LoadConfigOnce()
	if err != nil {
		return false
	}
	for _, registry := range config.InsecureRegistries {
		if strings.EqualFold(registry, target) {
			return true
		}
	}
	return false
}

// ResolveKey resolves the key by name.
// The default key is attempted if name is empty.
func ResolveKey(name string) (config.KeySuite, error) {
	signingKeys, err := config.LoadSigningKeys()
	if err != nil {
		return config.KeySuite{}, err
	}

	// if name is empty, look for default signing key
	if name == "" {
		return signingKeys.GetDefault()
	}

	return signingKeys.Get(name)
}
