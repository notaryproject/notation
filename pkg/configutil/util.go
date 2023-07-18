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
