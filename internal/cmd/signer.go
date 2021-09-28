package cmd

import (
	"time"

	"github.com/notaryproject/notation-go-lib"
	"github.com/notaryproject/notation-go-lib/crypto/timestamp"
	"github.com/notaryproject/notation/pkg/config"
	"github.com/notaryproject/notation/pkg/signature"
	"github.com/urfave/cli/v2"
)

// GetSigner returns a signer according to the CLI context.
func GetSigner(ctx *cli.Context) (notation.Signer, error) {
	// read paths of the signing key and its corresponding cert.
	var keyPath, certPath string
	if path := ctx.String(FlagKeyFile.Name); path != "" {
		keyPath = path
		certPath = ctx.String(FlagCertFile.Name)
	} else {
		var err error
		keyPath, certPath, err = config.ResolveKeyPath(ctx.String(FlagKey.Name))
		if err != nil {
			return nil, err
		}
	}

	// construct signer
	signer, err := signature.NewSignerFromFiles(keyPath, certPath)
	if err != nil {
		return nil, err
	}
	if endpoint := ctx.String(FlagTimestamp.Name); endpoint != "" {
		signer.TSA = timestamp.NewHTTPTimestamper(nil, endpoint)
	}
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
