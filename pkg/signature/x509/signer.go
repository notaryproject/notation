package x509

import (
	"bytes"
	"crypto"
	"crypto/x509"
	"errors"

	"github.com/docker/libtrust"
	cryptoutil "github.com/notaryproject/nv2/internal/crypto"
	"github.com/notaryproject/nv2/pkg/signature"
)

type signer struct {
	key      libtrust.PrivateKey
	keyID    string
	cert     *x509.Certificate
	rawCerts [][]byte
	hash     crypto.Hash
}

// NewSignerFromFiles creates a signer from files
func NewSignerFromFiles(keyPath, certPath string) (signature.Signer, error) {
	key, err := cryptoutil.ReadPrivateKeyFile(keyPath)
	if err != nil {
		return nil, err
	}
	if certPath == "" {
		return NewSigner(key, nil)
	}

	certs, err := cryptoutil.ReadCertificateFile(certPath)
	if err != nil {
		return nil, err
	}
	return NewSigner(key, certs)
}

// NewSigner creates a signer
func NewSigner(key libtrust.PrivateKey, certs []*x509.Certificate) (signature.Signer, error) {
	s := &signer{
		key:   key,
		keyID: key.KeyID(),
		hash:  crypto.SHA256,
	}
	if len(certs) == 0 {
		return s, nil
	}

	cert := certs[0]
	publicKey, err := libtrust.FromCryptoPublicKey(crypto.PublicKey(cert.PublicKey))
	if err != nil {
		return nil, err
	}
	if s.keyID != publicKey.KeyID() {
		return nil, errors.New("key and certificate mismatch")
	}
	s.cert = cert

	rawCerts := make([][]byte, 0, len(certs))
	for _, cert := range certs {
		rawCerts = append(rawCerts, cert.Raw)
	}
	s.rawCerts = rawCerts

	return s, nil
}

func (s *signer) Sign(raw []byte) (signature.Signature, error) {
	if s.cert != nil {
		if err := verifyReferences(raw, s.cert); err != nil {
			return signature.Signature{}, err
		}
	}

	sig, alg, err := s.key.Sign(bytes.NewReader(raw), s.hash)
	if err != nil {
		return signature.Signature{}, err
	}
	sigma := signature.Signature{
		Type:      Type,
		Algorithm: alg,
		Signature: sig,
	}

	if s.cert != nil {
		sigma.X5c = s.rawCerts
	} else {
		sigma.KeyID = s.keyID
	}
	return sigma, nil
}
