package cmd

import (
	"errors"
	"fmt"
	"time"

	"github.com/notaryproject/notation-core-go/signature/cose"
	"github.com/notaryproject/notation-core-go/signature/jws"
	"github.com/notaryproject/notation-go"
	"github.com/notaryproject/notation-go/plugin/manager"
	"github.com/notaryproject/notation-go/signature"
	"github.com/notaryproject/notation/internal/envelope"
	"github.com/notaryproject/notation/pkg/configutil"
)

// GetSigner returns a signer according to the CLI context.
func GetSigner(opts *SignerFlagOpts) (notation.Signer, error) {
	// Construct a signer from key and cert file if provided as CLI arguments
	mediaType, err := GetEnvelopeMediaType(opts.EnvelopeType)
	if err != nil {
		return nil, err
	}
	if keyPath := opts.KeyFile; keyPath != "" {
		certPath := opts.CertFile
		return signature.NewSignerFromFiles(keyPath, certPath, mediaType)
	}
	// Construct a signer from preconfigured key pair in config.json
	// if key name is provided as the CLI argument
	key, err := configutil.ResolveKey(opts.Key)
	if err != nil {
		return nil, err
	}
	if key.X509KeyPair != nil {
		return signature.NewSignerFromFiles(key.X509KeyPair.KeyPath, key.X509KeyPair.CertificatePath, mediaType)
	}
	// Construct a plugin signer if key name provided as the CLI argument
	// corresponds to an external key
	if key.ExternalKey != nil {
		mgr := manager.New()
		runner, err := mgr.Runner(key.PluginName)
		if err != nil {
			return nil, err
		}
		return signature.NewSignerPlugin(runner, key.ExternalKey.ID, key.PluginConfig, mediaType)
	}
	return nil, errors.New("unsupported key, either provide a local key and certificate file paths, or a key name in config.json, check [DOC_PLACEHOLDER] for details")
}

// GetExpiry returns the signature expiry according to the CLI context.
func GetExpiry(expiry time.Duration) time.Time {
	if expiry == 0 {
		return time.Time{}
	}
	return time.Now().Add(expiry)
}

func GetEnvelopeMediaType(sigFormat string) (string, error) {
	switch sigFormat {
	case envelope.JWS:
		return jws.MediaTypeEnvelope, nil
	case envelope.COSE:
		return cose.MediaTypeEnvelope, nil
	}
	return "", fmt.Errorf("signature format %s not supported", sigFormat)
}
