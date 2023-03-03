package cmd

import (
	"context"
	"errors"

	"github.com/notaryproject/notation-go"
	"github.com/notaryproject/notation-go/config"
	"github.com/notaryproject/notation-go/dir"
	"github.com/notaryproject/notation-go/plugin"
	"github.com/notaryproject/notation-go/signer"
	"github.com/notaryproject/notation/pkg/configutil"
)

// GetSigner returns a signer according to the CLI context.
func GetSigner(opts *SignerFlagOpts) (notation.Signer, error) {
	var key config.KeySuite
	var err error

	// Check if the options are valid for the key (Key is mutually exclusive with [KeyID, PluginName])
	if opts.KeyID != "" && opts.PluginName != "" {
		if opts.Key == "" {
			// Construct a signer from on-demand key
			mgr := plugin.NewCLIManager(dir.PluginFS())
			plugin, err := mgr.Get(context.Background(), opts.PluginName)
			if err != nil {
				return nil, err
			}
			return signer.NewFromPlugin(plugin, opts.KeyID, map[string]string{})
		} else {
			return nil, errors.New("incompatible options, do not provide a key name when providing a key ID and plugin name")
		}
	} else if opts.KeyID == "" && opts.PluginName == "" {
		// Construct a signer from preconfigured key pair in config.json
		// if key name is provided as the CLI argument
		key, err = configutil.ResolveKey(opts.Key)
	} else {
		if opts.Key == "" {
			return nil, errors.New("incompatible options, both a key ID and plugin name are required when not using an existing key")
		} else {
			return nil, errors.New("incompatible options, do not provide a key ID or plugin name when providing a key name")
		}
	}

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
