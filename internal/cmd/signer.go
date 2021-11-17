package cmd

import (
	"errors"
	"time"

	"github.com/notaryproject/notation-go-lib"
	"github.com/notaryproject/notation/pkg/config"
	"github.com/notaryproject/notation/pkg/signature"
	"github.com/urfave/cli/v2"
)

// GetSigner returns a signer according to the CLI context.
func GetSigner(ctx *cli.Context) (notation.Signer, error) {
	// read paths of the signing key and its corresponding cert.
	var keyPath, certPath string
	var pluginPath string
	var kmsKey config.KMSKeySuite
	var err error
	if path := ctx.String(FlagKeyFile.Name); path != "" {
		keyPath = path
		certPath = ctx.String(FlagCertFile.Name)
	} else {
		keyPath, certPath, err = config.ResolveKeyPath(ctx.String(FlagKey.Name))
		if err != nil {
			if !errors.Is(err, config.ErrKeyNotFound) {
				return nil, err
			}

			// check if the key is an external kms key
			if kmsKey, err = config.ResolveKMSKey(ctx.String(FlagKey.Name)); err != nil {
				return nil, err
			}
			// get the plugin path for the external kms key
			if pluginPath, err = config.ResolveKMSPluginPath(kmsKey.PluginName); err != nil {
				return nil, err
			}
		}
	}

	var signer notation.Signer
	if keyPath != "" && certPath != "" {
		// construct signer
		if signer, err = signature.NewSignerFromFiles(keyPath, certPath); err != nil {
			return nil, err
		}
	} else {
		// construct signer with external kms plugin
		if signer, err = signature.NewSignerWithPlugin(kmsKey, pluginPath); err != nil {
			return nil, err
		}
	}

	// if endpoint := ctx.String(FlagTimestamp.Name); endpoint != "" {
	// 	if signer.TSA, err = timestamp.NewHTTPTimestamper(nil, endpoint); err != nil {
	// 		return nil, err
	// 	}
	// }
	return signer, nil
}

// GetExpiry returns the signature expiry according to the CLI context.
func GetExpiry(ctx *cli.Context) time.Time {
	expiry := ctx.Duration(FlagExpiry.Name)
	if expiry == 0 {
		return time.Time{}
	}
	return time.Now().Add(expiry)
}
