// Copyright The Notary Project Authors.
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package config provides utility methods related to Notation config.json.
package config

import (
	"slices"
	"strings"
	"sync"

	"github.com/notaryproject/notation-go/config"
	"github.com/notaryproject/notation/internal/envelope"
)

// loadConfigOnce is a function that invokes loadConfig only once.
var loadConfigOnce = sync.OnceValues(loadConfig)

// LoadConfigOnce returns the previously read config file.
// If previous config file does not exist, it reads the config from file
// or return a default config if not found.
// The returned config is only suitable for read only scenarios for short-lived processes.
func LoadConfigOnce() (*config.Config, error) {
	return loadConfigOnce()
}

// loadConfig reads the config from file or return a default config if not
// found.
func loadConfig() (*config.Config, error) {
	configInfo, err := config.LoadConfig()
	if err != nil {
		return nil, err
	}
	// set default value
	configInfo.SignatureFormat = strings.ToLower(configInfo.SignatureFormat)
	if configInfo.SignatureFormat == "" {
		configInfo.SignatureFormat = envelope.JWS
	}
	return configInfo, nil
}

// IsRegistryInsecure checks whether a registry is in the list of insecure
// registries under Notation's config file.
func IsRegistryInsecure(target string) bool {
	config, err := LoadConfigOnce()
	if err != nil {
		return false
	}
	return slices.ContainsFunc(config.InsecureRegistries, func(registry string) bool {
		return strings.EqualFold(registry, target)
	})
}
