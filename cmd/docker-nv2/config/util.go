package config

import (
	"errors"
	"os"
)

// ErrNotaryDisabled indicates that notary is disabled
var ErrNotaryDisabled = errors.New("notary disabled")

// CheckNotaryEnabled checks the config file whether notary is enabled or not.
func CheckNotaryEnabled() error {
	config, err := Load()
	if err != nil {
		if !os.IsNotExist(err) {
			return err
		}
		config = New()
	}
	if config.Enabled {
		return nil
	}
	return ErrNotaryDisabled
}

// IsRegistryInsecure checks whether the registry is in the list of insecure registries.
func IsRegistryInsecure(target string) bool {
	config, err := Load()
	if err != nil {
		return false
	}
	for _, registry := range config.InsecureRegistries {
		if registry == target {
			return true
		}
	}
	return false
}
