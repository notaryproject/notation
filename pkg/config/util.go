package config

import (
	"errors"
	"strings"

	"github.com/notaryproject/notation/internal/slices"
)

var (
	// ErrKeyNotFound indicates that the signing key is not found.
	ErrKeyNotFound = errors.New("signing key not found")
)

// IsRegistryInsecure checks whether the registry is in the list of insecure registries.
func IsRegistryInsecure(target string) bool {
	config, err := LoadOrDefaultOnce()
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
func ResolveKey(name string) (KeySuite, error) {
	config, err := LoadOrDefaultOnce()
	if err != nil {
		return KeySuite{}, err
	}
	if name == "" {
		name = config.SigningKeys.Default
	}
	idx := slices.Index(config.SigningKeys.Keys, name)
	if idx < 0 {
		return KeySuite{}, ErrKeyNotFound
	}
	return config.SigningKeys.Keys[idx], nil
}
