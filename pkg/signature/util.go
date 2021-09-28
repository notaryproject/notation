package signature

import (
	"bytes"
	"crypto"
	"crypto/x509"
)

// HasPairedCert returns true if there is a certificate matching the given key.
func HasPairedCert(key crypto.PrivateKey, certs []*x509.Certificate) bool {
	var pk crypto.PublicKey
	if key, ok := key.(interface {
		Public() crypto.PublicKey
	}); ok {
		pk = key.Public()
	} else {
		return false
	}
	if len(certs) == 0 {
		return false
	}

	keyBytes, err := x509.MarshalPKIXPublicKey(pk)
	if err != nil {
		return false
	}
	for _, cert := range certs {
		certKeyBytes, err := x509.MarshalPKIXPublicKey(cert.PublicKey)
		if err != nil {
			continue
		}
		if bytes.Equal(keyBytes, certKeyBytes) {
			return true
		}
	}
	return false
}
