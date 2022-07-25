package configutil

import (
	"sync"

	"github.com/notaryproject/notation-go/config"
)

var (
	// configInfo is the config.json data
	configInfo *config.Config
	configOnce sync.Once

	// signingKeyInfo if the signingkeys.json data
	signingKeysInfo *config.SigningKeys
	signingKeysOnce sync.Once
)

// LoadConfigOnce returns the previously read config file.
// If previous config file does not exist, it reads the config from file
// or return a default config if not found.
// The returned config is only suitable for read only scenarios for short-lived processes.
func LoadConfigOnce() (*config.Config, error) {
	var err error
	configOnce.Do(func() {
		configInfo, err = config.LoadConfig()
	})
	return configInfo, err
}

// LoadSigningKeysOnce returns the previously read config file.
// If previous config file does not exist, it reads the config from file
// or return a default config if not found.
// The returned config is only suitable for read only scenarios for short-lived processes.
func LoadSigningkeysOnce() (*config.SigningKeys, error) {
	var err error
	signingKeysOnce.Do(func() {
		signingKeysInfo, err = config.LoadSigningKeys()
	})
	return signingKeysInfo, err
}
