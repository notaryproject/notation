package signature

import (
	"crypto"
	"crypto/x509"

	"github.com/notaryproject/notation-go-lib/crypto/cryptoutil"
	"github.com/notaryproject/notation-go-lib/signature/jws"
	"github.com/opencontainers/go-digest"
)

// KeyID returns a notation-specific key ID for the public key portion of the key.
func KeyID(key interface{}) (string, error) {
	if k, ok := key.(interface {
		Public() crypto.PublicKey
	}); ok {
		key = k.Public()
	}

	keyBytes, err := x509.MarshalPKIXPublicKey(key)
	if err != nil {
		return "", err
	}
	keyDigest := digest.SHA256.FromBytes(keyBytes)

	return keyDigest.Encoded(), nil
}

// NewSignerFromFiles creates a signer from key, certificate files
func NewSignerFromFiles(keyPath, certPath string) (*jws.Signer, error) {
	key, err := cryptoutil.ReadPrivateKeyFile(keyPath)
	if err != nil {
		return nil, err
	}
	method, err := jws.SigningMethodFromKey(key)
	if err != nil {
		return nil, err
	}

	if certPath == "" {
		keyID, err := KeyID(key)
		if err != nil {
			return nil, err
		}
		return jws.NewSignerWithKeyID(method, key, keyID)
	}

	certs, err := cryptoutil.ReadCertificateFile(certPath)
	if err != nil {
		return nil, err
	}
	return jws.NewSignerWithCertificateChain(method, key, certs)
}

// NewSignerFromFiles creates a verifier from certificate files
func NewVerifierFromFiles(certPaths, caCertPaths []string) (*jws.Verifier, error) {
	var keys []*jws.VerificationKey
	roots := x509.NewCertPool()
	for _, path := range certPaths {
		bundledCerts, err := cryptoutil.ReadCertificateFile(path)
		if err != nil {
			return nil, err
		}
		for _, cert := range bundledCerts {
			keyID, err := KeyID(cert.PublicKey)
			if err != nil {
				return nil, err
			}
			key, err := jws.NewVerificationKey(cert.PublicKey, keyID)
			if err != nil {
				return nil, err
			}
			keys = append(keys, key)
			roots.AddCert(cert)
		}
	}
	for _, path := range caCertPaths {
		bundledCerts, err := cryptoutil.ReadCertificateFile(path)
		if err != nil {
			return nil, err
		}
		for _, cert := range bundledCerts {
			roots.AddCert(cert)
		}
	}

	// construct verifier
	verifier := jws.NewVerifier(keys)
	verifier.VerifyOptions.Roots = roots
	return verifier, nil
}
