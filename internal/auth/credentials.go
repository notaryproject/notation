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

package auth

import (
	"fmt"

	"github.com/notaryproject/notation-go/dir"
	"oras.land/oras-go/v2/registry/remote/credentials"
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
