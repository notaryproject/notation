package x509

import (
	"bytes"
	"crypto"
	"crypto/x509"
	"encoding/json"
	"errors"
	"strings"

	"github.com/docker/libtrust"
	"github.com/notaryproject/nv2/pkg/signature"
)

type verifier struct {
	keys  map[string]libtrust.PublicKey
	certs map[string]*x509.Certificate
	roots *x509.CertPool
}

// NewVerifier creates a verifier
func NewVerifier(certs []*x509.Certificate, roots *x509.CertPool) (signature.Verifier, error) {
	if roots == nil {
		if certs == nil {
			pool, err := x509.SystemCertPool()
			if err != nil {
				return nil, err
			}
			roots = pool
		} else {
			roots = x509.NewCertPool()
		}
		for _, cert := range certs {
			roots.AddCert(cert)
		}
	}

	keys := make(map[string]libtrust.PublicKey, len(certs))
	keyedCerts := make(map[string]*x509.Certificate, len(certs))
	for _, cert := range certs {
		key, err := libtrust.FromCryptoPublicKey(crypto.PublicKey(cert.PublicKey))
		if err != nil {
			return nil, err
		}
		keyID := key.KeyID()
		keys[keyID] = key
		keyedCerts[keyID] = cert
	}

	return &verifier{
		keys:  keys,
		certs: keyedCerts,
		roots: roots,
	}, nil
}

func (v *verifier) Type() string {
	return Type
}

func (v *verifier) Verify(content []byte, sig signature.Signature) error {
	if sig.Type != Type {
		return signature.ErrInvalidSignatureType
	}

	key, cert, err := v.getVerificationKeyPair(sig)
	if err != nil {
		return err
	}
	if err := key.Verify(bytes.NewReader(content), sig.Algorithm, sig.Signature); err != nil {
		return err
	}
	return verifyReferences(content, cert)
}

func (v *verifier) getVerificationKeyPair(sig signature.Signature) (libtrust.PublicKey, *x509.Certificate, error) {
	switch {
	case len(sig.X5c) > 0:
		return v.getVerificationKeyPairFromX5c(sig.X5c)
	case sig.KeyID != "":
		return v.getVerificationKeyPairFromKeyID(sig.KeyID)
	default:
		return nil, nil, errors.New("missing verification key")
	}
}

func (v *verifier) getVerificationKeyPairFromKeyID(keyID string) (libtrust.PublicKey, *x509.Certificate, error) {
	key, found := v.keys[keyID]
	if !found {
		return nil, nil, errors.New("key not found: " + keyID)
	}
	cert, found := v.certs[keyID]
	if !found {
		return nil, nil, errors.New("cert not found: " + keyID)
	}
	return key, cert, nil
}

func (v *verifier) getVerificationKeyPairFromX5c(x5c [][]byte) (libtrust.PublicKey, *x509.Certificate, error) {
	certs := make([]*x509.Certificate, 0, len(x5c))
	for _, certBytes := range x5c {
		cert, err := x509.ParseCertificate(certBytes)
		if err != nil {
			return nil, nil, err
		}
		certs = append(certs, cert)
	}

	intermediates := x509.NewCertPool()
	for _, cert := range certs[1:] {
		intermediates.AddCert(cert)
	}

	cert := certs[0]
	if _, err := cert.Verify(x509.VerifyOptions{
		Intermediates: intermediates,
		Roots:         v.roots,
		KeyUsages:     []x509.ExtKeyUsage{x509.ExtKeyUsageAny},
	}); err != nil {
		return nil, nil, err
	}

	key, err := libtrust.FromCryptoPublicKey(crypto.PublicKey(cert.PublicKey))
	if err != nil {
		return nil, nil, err
	}
	return key, cert, nil
}

func verifyReferences(raw []byte, cert *x509.Certificate) error {
	var content signature.Content
	if err := json.Unmarshal(raw, &content); err != nil {
		return err
	}
	roots := x509.NewCertPool()
	roots.AddCert(cert)
	for _, manifest := range content.Manifests {
		for _, reference := range manifest.References {
			if _, err := cert.Verify(x509.VerifyOptions{
				DNSName: strings.SplitN(reference, "/", 2)[0],
				Roots:   roots,
			}); err != nil {
				return err
			}
		}
	}
	return nil
}
