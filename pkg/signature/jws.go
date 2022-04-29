package signature

import (
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"os"

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
	cert, err := tls.LoadX509KeyPair(certPath, keyPath)
	if err != nil {
		return nil, err
	}
	if len(cert.Certificate) == 0 {
		return nil, errors.New("missing signer certificate chain")
	}

	// parse cert
	certs := make([]*x509.Certificate, len(cert.Certificate))
	for i, c := range cert.Certificate {
		cert, err := x509.ParseCertificate(c)
		if err != nil {
			return nil, err
		}
		certs[i] = cert
	}

	key := cert.PrivateKey
	method, err := jws.SigningMethodFromKey(key)
	if err != nil {
		return nil, err
	}

	// create signer
	return jws.NewSignerWithCertificateChain(method, key, certs)
}

// NewSignerFromFiles creates a verifier from certificate files
func NewVerifierFromFiles(certPaths []string) (*jws.Verifier, error) {
	verifier := jws.NewVerifier()
	verifier.VerifyOptions.Roots = x509.NewCertPool()
	for _, path := range certPaths {
		data, err := os.ReadFile(path)
		if err != nil {
			return nil, err
		}
		if !verifier.VerifyOptions.Roots.AppendCertsFromPEM(data) {
			return nil, fmt.Errorf("failed to parse PEM certificate: %q", path)
		}
	}
	return verifier, nil
}
