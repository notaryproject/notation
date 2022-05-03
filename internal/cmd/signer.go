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
	// read paths of the signing key and its corresponding cert.
	if keyPath := ctx.String(FlagKeyFile.Name); keyPath != "" {
		certPath := ctx.String(FlagCertFile.Name)
		return signature.NewSignerFromFiles(keyPath, certPath)
	}
	key, err := config.ResolveKey(ctx.String(FlagKey.Name))
	if err != nil {
		return nil, err
	}
	if key.X509KeyPair != nil {
		return signature.NewSignerFromFiles(key.X509KeyPair.KeyPath, key.X509KeyPair.CertificatePath)
	}
	if key.ExternalKey != nil {
		return &jws.PluginSigner{
			Runner:     manager.NewManager(),
			PluginName: key.PluginName,
			KeyID:      key.ExternalKey.ID,
			KeyName:    key.Name,
		}, nil
	}
	return nil, errors.New("unsupported key")
}

// GetExpiry returns the signature expiry according to the CLI context.
func GetExpiry(ctx *cli.Context) time.Time {
	expiry := ctx.Duration(FlagExpiry.Name)
	if expiry == 0 {
		return time.Time{}
	}
	return time.Now().Add(expiry)
}
