package config

import (
	"errors"
	"os"
)

// ErrNotationDisabled indicates that notation is disabled
var ErrNotationDisabled = errors.New("notation disabled")

// CheckNotationEnabled checks the config file whether notation is enabled or not.
func CheckNotationEnabled() error {
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
	return ErrNotationDisabled
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
