package signature

import (
	"crypto/tls"
	"crypto/x509"
	"errors"
	"os"

	"github.com/notaryproject/notation-go/crypto/cryptoutil"
	"github.com/notaryproject/notation-go/signature/jws"
)

// NewSignerFromFiles creates a signer from key, certificate files
func NewSignerFromFiles(keyPath, certPath string) (*jws.Signer, error) {
	if keyPath == "" {
		return nil, errors.New("key path not specified")
	}
	if certPath == "" {
		return nil, errors.New("certificate path not specified")
	}

	// read key / cert pair
	keyPEM, err := os.ReadFile(keyPath)
	if err != nil {
		return nil, err
	}
	certPEM, err := os.ReadFile(certPath)
	if err != nil {
		return nil, err
	}
	keyPair, err := tls.X509KeyPair(certPEM, keyPEM)
	if err != nil {
		return nil, err
	}
	key := keyPair.PrivateKey
	method, err := jws.SigningMethodFromKey(key)
	if err != nil {
		return nil, err
	}

	// parse cert
	certs, err := cryptoutil.ParseCertificatePEM(certPEM)
	if err != nil {
		return nil, err
	}

	// create signer
	return jws.NewSignerWithCertificateChain(method, key, certs)
}

// NewSignerFromFiles creates a verifier from certificate files
func NewVerifierFromFiles(certPaths []string) (*jws.Verifier, error) {
	roots := x509.NewCertPool()
	for _, path := range certPaths {
		bundledCerts, err := cryptoutil.ReadCertificateFile(path)
		if err != nil {
			return nil, err
		}
		for _, cert := range bundledCerts {
			roots.AddCert(cert)
		}
	}
	verifier := jws.NewVerifier()
	verifier.VerifyOptions.Roots = roots
	return verifier, nil
}
