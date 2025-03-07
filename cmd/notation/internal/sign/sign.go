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

// Package sign provides utility methods related to sign commands.
package sign

import (
	"context"
	"errors"

	"github.com/notaryproject/notation-go"
	"github.com/notaryproject/notation-go/config"
	"github.com/notaryproject/notation-go/dir"
	"github.com/notaryproject/notation-go/plugin"
	"github.com/notaryproject/notation-go/signer"
	"github.com/notaryproject/notation/cmd/notation/internal/flag"
)

// Signer is embedded with notation.BlobSigner and notation.Signer.
type Signer interface {
	notation.BlobSigner
	notation.Signer
}

// GetSigner returns a Signer based on user opts.
func GetSigner(ctx context.Context, opts *flag.SignerFlagOpts) (Signer, error) {
	// Check if using on-demand key
	if opts.KeyID != "" && opts.PluginName != "" && opts.Key == "" {
		// Construct a signer from on-demand key
		mgr := plugin.NewCLIManager(dir.PluginFS())
		plugin, err := mgr.Get(ctx, opts.PluginName)
		if err != nil {
			return nil, err
		}
		return signer.NewPluginSigner(plugin, opts.KeyID, map[string]string{})
	}

	// Construct a signer from preconfigured key pair in config.json
	// if key name is provided as the CLI argument
	key, err := resolveKey(opts.Key)
	if err != nil {
		return nil, err
	}
	if key.X509KeyPair != nil {
		return signer.NewGenericSignerFromFiles(key.X509KeyPair.KeyPath, key.X509KeyPair.CertificatePath)
	}

	// Construct a plugin signer if key name provided as the CLI argument
	// corresponds to an external key
	if key.ExternalKey != nil {
		mgr := plugin.NewCLIManager(dir.PluginFS())
		plugin, err := mgr.Get(ctx, key.PluginName)
		if err != nil {
			return nil, err
		}
		return signer.NewPluginSigner(plugin, key.ExternalKey.ID, key.PluginConfig)
	}
	return nil, errors.New("unsupported key, either provide a local key and certificate file paths, or a key name in config.json, check https://notaryproject.dev/docs/user-guides/how-to/notation-config-file/ for details")
}

// resolveKey resolves the key by name.
// The default key is attempted if name is empty.
func resolveKey(name string) (config.KeySuite, error) {
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
