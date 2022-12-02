package cmd

import (
	"context"
	"errors"

	"github.com/notaryproject/notation-go"
	"github.com/notaryproject/notation-go/dir"
	"github.com/notaryproject/notation-go/plugin"
	"github.com/notaryproject/notation-go/signer"
	"github.com/notaryproject/notation/pkg/configutil"
)

// GetSigner returns a signer according to the CLI context.
func GetSigner(opts *SignerFlagOpts) (notation.Signer, error) {
	// Construct a signer from preconfigured key pair in config.json
	// if key name is provided as the CLI argument
	key, err := configutil.ResolveKey(opts.Key)
	if err != nil {
		return nil, err
	}
	if key.X509KeyPair != nil {
		return signer.NewFromFiles(key.X509KeyPair.KeyPath, key.X509KeyPair.CertificatePath)
	}
	// Construct a plugin signer if key name provided as the CLI argument
	// corresponds to an external key
	if key.ExternalKey != nil {
		mgr := plugin.NewCLIManager(dir.PluginFS())
		plugin, err := mgr.Get(context.Background(), key.PluginName)
		if err != nil {
			return nil, err
		}
		return signer.NewFromPlugin(plugin, key.ExternalKey.ID, key.PluginConfig)
	}
	return nil, errors.New("unsupported key, either provide a local key and certificate file paths, or a key name in config.json, check [DOC_PLACEHOLDER] for details")
}
