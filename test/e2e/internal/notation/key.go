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

package notation

import (
	"os"
	"path/filepath"
)

const (
	SigningKeysFileName     = "signingkeys.json"
	LocalKeysDirName        = "localkeys"
	LocalConfigJsonsDirName = "configjsons"
)

// X509KeyPair contains the paths of a public/private key pair files.
type X509KeyPair struct {
	KeyPath         string `json:"keyPath"`
	CertificatePath string `json:"certPath"`
}

// ExternalKey contains the necessary information to delegate
// the signing operation to the named plugin.
type ExternalKey struct {
	ID           string            `json:"id,omitempty"`
	PluginName   string            `json:"pluginName,omitempty"`
	PluginConfig map[string]string `json:"pluginConfig,omitempty"`
}

// KeySuite is a named key suite.
type KeySuite struct {
	Name string `json:"name"`
	*X509KeyPair
	*ExternalKey
}

// SigningKeys reflects the signingkeys.json file.
type SigningKeys struct {
	Default string     `json:"default"`
	Keys    []KeySuite `json:"keys"`
}

// AddKeyPairs creates the signingkeys.json file and the localkeys directory
// with e2e.key and e2e.crt
func AddKeyPairs(dir, keyName, certName string) error {
	// create signingkeys.json files
	if err := saveJSON(
		generateSigningKeys(dir),
		filepath.Join(dir, SigningKeysFileName)); err != nil {
		return err
	}

	// create localkeys directory
	localKeysDir := filepath.Join(dir, LocalKeysDirName)
	os.MkdirAll(localKeysDir, 0700)

	// copy key and cert files
	if err := copyFile(filepath.Join(NotationE2ELocalKeysDir, keyName), filepath.Join(localKeysDir, "e2e.key")); err != nil {
		return err
	}
	return copyFile(filepath.Join(NotationE2ELocalKeysDir, certName), filepath.Join(localKeysDir, "e2e.crt"))
}

// generateSigningKeys generates the signingkeys.json for notation.
func generateSigningKeys(dir string) *SigningKeys {
	return &SigningKeys{
		Default: "e2e",
		Keys: []KeySuite{
			{
				Name: "e2e",
				X509KeyPair: &X509KeyPair{
					KeyPath:         filepath.Join(dir, "localkeys", "e2e.key"),
					CertificatePath: filepath.Join(dir, "localkeys", "e2e.crt"),
				},
			},
		},
	}
}

// generatePluginKeys generates pluginkeys.json for e2e-plugin.
func generatePluginKeys(dir string) *SigningKeys {
	return &SigningKeys{
		Keys: []KeySuite{
			{
				Name: "e2e-plugin",
				X509KeyPair: &X509KeyPair{
					KeyPath:         filepath.Join(dir, "localkeys", "e2e.key"),
					CertificatePath: filepath.Join(dir, "localkeys", "e2e.crt"),
				},
				ExternalKey: &ExternalKey{
					ID:         "key1",
					PluginName: PluginName,
				},
			},
		},
	}
}
