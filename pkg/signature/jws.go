package signature

import (
	"crypto/x509"
	"errors"

	"github.com/notaryproject/notation-go-lib/crypto/cryptoutil"
	"github.com/notaryproject/notation-go-lib/signature/jws"
)

// NewSignerFromFiles creates a signer from key, certificate files
func NewSignerFromFiles(keyPath, certPath string) (*jws.Signer, error) {
	if keyPath == "" {
		return nil, errors.New("key path not specified")
	}
	if certPath == "" {
		return nil, errors.New("certificate path not specified")
	}

	// read key
	key, err := cryptoutil.ReadPrivateKeyFile(keyPath)
	if err != nil {
		return nil, err
	}
	method, err := jws.SigningMethodFromKey(key)
	if err != nil {
		return nil, err
	}

	// read cert
	certs, err := cryptoutil.ReadCertificateFile(certPath)
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
