package local

import (
	"context"
	"crypto"
	"crypto/rand"
	"crypto/x509"
	"errors"
	"io/ioutil"

	"github.com/notaryproject/nv2/pkg/tuf"
	"github.com/theupdateframework/notary/tuf/data"
	"github.com/theupdateframework/notary/tuf/utils"
)

type signer struct {
	key      data.PrivateKey
	keyID    string
	cert     *x509.Certificate
	rawCerts [][]byte
	hash     crypto.Hash
}

// NewSignerFromFiles creates a signer from files
func NewSignerFromFiles(keyPath, certPath string) (tuf.Signer, error) {
	keyPEM, err := ioutil.ReadFile(keyPath)
	if err != nil {
		return nil, err
	}
	key, err := utils.ParsePEMPrivateKey(keyPEM, "")
	if err != nil {
		return nil, err
	}
	if certPath == "" {
		return NewSigner(key, nil)
	}

	certs, err := utils.LoadCertBundleFromFile(certPath)
	if err != nil {
		return nil, err
	}
	return NewSigner(key, certs)
}

// NewSigner creates a signer
func NewSigner(key data.PrivateKey, certs []*x509.Certificate) (tuf.Signer, error) {
	s := &signer{
		key:  key,
		hash: crypto.SHA256,
	}
	if len(certs) == 0 {
		s.keyID = key.ID()
		return s, nil
	}

	cert := certs[0]
	publicKey := utils.CertToKey(cert)
	if publicKey == nil {
		return nil, errors.New("unknown certificate key type")
	}
	keyID, err := utils.CanonicalKeyID(publicKey)
	if err != nil {
		return nil, err
	}
	if keyID != s.key.ID() {
		return nil, errors.New("key and certificate mismatch")
	}
	// Docker Notary 0.6.0 implementation uses non-canonical key ID for delegation roles,
	// which should be canonical.
	s.keyID = publicKey.ID()
	s.cert = cert

	rawCerts := make([][]byte, 0, len(certs))
	for _, cert := range certs {
		rawCerts = append(rawCerts, cert.Raw)
	}
	s.rawCerts = rawCerts

	return s, nil
}

func (s *signer) Sign(_ context.Context, raw []byte) (tuf.Signature, error) {
	if s.cert != nil {
		if err := verifyReferences(raw, s.cert); err != nil {
			return tuf.Signature{}, err
		}
	}

	sig, err := s.key.Sign(rand.Reader, raw, nil)
	if err != nil {
		return tuf.Signature{}, err
	}
	sigma := tuf.Signature{
		KeyID:     s.keyID,
		Method:    s.key.SignatureAlgorithm(),
		Signature: sig,
	}

	if s.cert != nil {
		sigma.X5c = s.rawCerts
	}
	return sigma, nil
}
