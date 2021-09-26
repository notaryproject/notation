package crypto

import (
	"crypto"
	"crypto/x509"

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
