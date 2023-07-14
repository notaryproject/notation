package cmd

import (
	"context"
	"errors"

	"github.com/notaryproject/notation-go"
	"github.com/notaryproject/notation-go/dir"
	"github.com/notaryproject/notation-go/plugin"
	"github.com/notaryproject/notation-go/signer"
	"github.com/notaryproject/notation-go/signingkeys"
)

// GetSigner returns a signer according to the CLI context.
func GetSigner(ctx context.Context, opts *SignerFlagOpts) (notation.Signer, error) {
	// Check if using on-demand key
	if opts.KeyID != "" && opts.PluginName != "" && opts.Key == "" {
		// Construct a signer from on-demand key
		mgr := plugin.NewCLIManager(dir.PluginFS())
		plugin, err := mgr.Get(ctx, opts.PluginName)
		if err != nil {
			return nil, err
		}
		return signer.NewFromPlugin(plugin, opts.KeyID, map[string]string{})
	}

	// Construct a signer from preconfigured key pair in config.json
	// if key name is provided as the CLI argument
	sKeys, err := signingkeys.LoadFromCache()
	if err != nil {
		return nil, err
	}
	key, err := sKeys.Resolve(opts.Key)
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
		plugin, err := mgr.Get(ctx, key.PluginName)
		if err != nil {
			return nil, err
		}
		return signer.NewFromPlugin(plugin, key.ExternalKey.ID, key.PluginConfig)
	}
	return nil, errors.New("unsupported key, either provide a local key and certificate file paths, or a key name in config.json, check [DOC_PLACEHOLDER] for details")
}
