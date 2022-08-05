package configutil

import (
	"errors"
	"strings"

	"github.com/notaryproject/notation-go/config"
	"github.com/notaryproject/notation/internal/slices"
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
	signingKeys, err := LoadSigningkeysOnce()
	if err != nil {
		return config.KeySuite{}, err
	}
	if name == "" {
		name = signingKeys.Default
	}
	idx := slices.Index(signingKeys.Keys, name)
	if idx < 0 {
		return config.KeySuite{}, ErrKeyNotFound
	}
	return signingKeys.Keys[idx], nil
}
