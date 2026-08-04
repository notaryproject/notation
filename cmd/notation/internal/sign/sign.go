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
	"crypto"
	"crypto/ecdsa"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"

	"github.com/notaryproject/notation-core-go/signature"
	"github.com/notaryproject/notation-go"
	"github.com/notaryproject/notation-go/config"
	"github.com/notaryproject/notation-go/dir"
	"github.com/notaryproject/notation-go/plugin"
	"github.com/notaryproject/notation-go/signer"
	"github.com/notaryproject/notation/v2/cmd/notation/internal/flag"
)

// localPrimitiveSigner implements signature.Signer for local key signing
// using PKCS#1 v1.5 (RSA) or ECDSA, required for PKCS#7 dm-verity signatures.
type localPrimitiveSigner struct {
	keySpec signature.KeySpec
	key     crypto.PrivateKey
	certs   []*x509.Certificate
}

func (s *localPrimitiveSigner) Sign(payload []byte) ([]byte, []*x509.Certificate, error) {
	var hash crypto.Hash
	switch s.keySpec.Size {
	case 256:
		hash = crypto.SHA256
	case 384:
		hash = crypto.SHA384
	case 512, 521:
		hash = crypto.SHA512
	default:
		hash = crypto.SHA256
	}

	h := hash.New()
	h.Write(payload)
	digest := h.Sum(nil)

	var sig []byte
	var err error

	switch k := s.key.(type) {
	case *rsa.PrivateKey:
		sig, err = rsa.SignPKCS1v15(rand.Reader, k, hash, digest)
	case *ecdsa.PrivateKey:
		sig, err = ecdsa.SignASN1(rand.Reader, k, digest)
	default:
		return nil, nil, fmt.Errorf("unsupported key type: %T", s.key)
	}

	if err != nil {
		return nil, nil, err
	}

	return sig, s.certs, nil
}

func (s *localPrimitiveSigner) KeySpec() (signature.KeySpec, error) {
	return s.keySpec, nil
}

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

// SigningSchemeConfigKey is the plugin config key for requesting PKCS#1 v1.5 signing.
const SigningSchemeConfigKey = "signing_scheme"

// SigningSchemePKCS1v15 requests RSASSA-PKCS1-v1_5 from plugins (required for kernel dm-verity).
const SigningSchemePKCS1v15 = "rsassa-pkcs1-v1_5"

// GetPrimitiveSigner returns a signature.Signer for PKCS#7 dm-verity signing.
// Unlike GetSigner (which returns notation.Signer for JWS/COSE), this returns
// a raw signer used with the PKCS#7 envelope. For plugins, it automatically
// sets signing_scheme=rsassa-pkcs1-v1_5 since the Linux kernel only supports
// PKCS#1 v1.5 signature verification.
func GetPrimitiveSigner(ctx context.Context, opts *flag.SignerFlagOpts) (signature.Signer, error) {
	if opts.KeyID != "" && opts.PluginName != "" && opts.Key == "" {
		mgr := plugin.NewCLIManager(dir.PluginFS())
		signPlugin, err := mgr.Get(ctx, opts.PluginName)
		if err != nil {
			return nil, err
		}

		pluginConfig := map[string]string{
			SigningSchemeConfigKey: SigningSchemePKCS1v15,
		}

		keySpec, err := signer.GetKeySpecFromPlugin(ctx, signPlugin, opts.KeyID, pluginConfig)
		if err != nil {
			return nil, err
		}

		return signer.NewPluginPrimitiveSigner(ctx, signPlugin, opts.KeyID, keySpec, pluginConfig), nil
	}

	key, err := resolveKey(opts.Key)
	if err != nil {
		return nil, err
	}

	if key.X509KeyPair != nil {
		return NewLocalSignerFromFiles(key.X509KeyPair.KeyPath, key.X509KeyPair.CertificatePath)
	}

	if key.ExternalKey != nil {
		mgr := plugin.NewCLIManager(dir.PluginFS())
		signPlugin, err := mgr.Get(ctx, key.PluginName)
		if err != nil {
			return nil, err
		}

		pluginConfig := make(map[string]string)
		for k, v := range key.PluginConfig {
			pluginConfig[k] = v
		}
		pluginConfig[SigningSchemeConfigKey] = SigningSchemePKCS1v15

		keySpec, err := signer.GetKeySpecFromPlugin(ctx, signPlugin, key.ExternalKey.ID, pluginConfig)
		if err != nil {
			return nil, err
		}

		return signer.NewPluginPrimitiveSigner(ctx, signPlugin, key.ExternalKey.ID, keySpec, pluginConfig), nil
	}

	return nil, errors.New("unsupported key for primitive signing, either provide a local key and certificate file paths, or a key name in config.json")
}

// NewLocalSignerFromFiles creates a signature.Signer from local key and certificate files.
func NewLocalSignerFromFiles(keyPath, certPath string) (signature.Signer, error) {
	if keyPath == "" {
		return nil, errors.New("key path not specified")
	}
	if certPath == "" {
		return nil, errors.New("certificate path not specified")
	}

	cert, err := tls.LoadX509KeyPair(certPath, keyPath)
	if err != nil {
		return nil, err
	}
	if len(cert.Certificate) == 0 {
		return nil, fmt.Errorf("%q does not contain certificate", certPath)
	}

	certs := make([]*x509.Certificate, len(cert.Certificate))
	for i, c := range cert.Certificate {
		certs[i], err = x509.ParseCertificate(c)
		if err != nil {
			return nil, err
		}
	}

	keySpec, err := signature.ExtractKeySpec(certs[0])
	if err != nil {
		return nil, fmt.Errorf("failed to extract key spec: %w", err)
	}

	return &localPrimitiveSigner{
		keySpec: keySpec,
		key:     cert.PrivateKey,
		certs:   certs,
	}, nil
}
