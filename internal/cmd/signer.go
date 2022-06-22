package cmd

import (
	"errors"
	"time"

	"github.com/notaryproject/notation-go"
	"github.com/notaryproject/notation-go/plugin/manager"
	"github.com/notaryproject/notation-go/signature/jws"
	"github.com/notaryproject/notation/pkg/config"
	"github.com/notaryproject/notation/pkg/signature"
	"github.com/urfave/cli/v2"
)

// GetSigner returns a signer according to the CLI context.
func GetSigner(ctx *cli.Context) (notation.Signer, error) {
	// Construct a signer from key and cert file if provided as CLI arguments
	if keyPath := ctx.String(FlagKeyFile.Name); keyPath != "" {
		certPath := ctx.String(FlagCertFile.Name)
		return signature.NewSignerFromFiles(keyPath, certPath)
	}
	// Construct a signer from preconfigured key pair in config.json
	// if key name is provided as the CLI argument
	key, err := config.ResolveKey(ctx.String(FlagKey.Name))
	if err != nil {
		return nil, err
	}
	if key.X509KeyPair != nil {
		return signature.NewSignerFromFiles(key.X509KeyPair.KeyPath, key.X509KeyPair.CertificatePath)
	}
	// Construct a plugin signer if key name provided as the CLI argument
	// corresponds to an external key
	if key.ExternalKey != nil {
		mgr := manager.New(config.PluginDirPath)
		runner, err := mgr.Runner(key.PluginName)
		if err != nil {
			return nil, err
		}
		return jws.NewSignerPlugin(runner, key.ExternalKey.ID, key.PluginConfig)
	}
	return nil, errors.New("unsupported key, either provide a local key and certificate file paths, or a key name in config.json, check [DOC_PLACEHOLDER] for details")
}

// GetExpiry returns the signature expiry according to the CLI context.
func GetExpiry(ctx *cli.Context) time.Time {
	expiry := ctx.Duration(FlagExpiry.Name)
	if expiry == 0 {
		return time.Time{}
	}
	return time.Now().Add(expiry)
}
