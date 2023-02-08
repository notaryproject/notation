package configutil

import (
	"strings"
	"sync"

	"github.com/notaryproject/notation-go/config"
	"github.com/notaryproject/notation/internal/envelope"
)

var (
	// configInfo is the config.json data
	configInfo *config.Config
	configOnce sync.Once
)

// LoadConfigOnce returns the previously read config file.
// If previous config file does not exist, it reads the config from file
// or return a default config if not found.
// The returned config is only suitable for read only scenarios for short-lived processes.
func LoadConfigOnce() (*config.Config, error) {
	var err error
	configOnce.Do(func() {
		configInfo, err = config.LoadConfig()
		if err != nil {
			return
		}
		// set default value
		configInfo.SignatureFormat = strings.ToLower(configInfo.SignatureFormat)
		if configInfo.SignatureFormat == "" {
			configInfo.SignatureFormat = envelope.JWS
		}
	})
	return configInfo, err
}
