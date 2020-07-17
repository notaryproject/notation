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
	certs, err := cryptoutil.ReadCertificateFile(certPath)
	if err != nil {
		return nil, err
	}
	return NewSigner(key, certs)
}

// NewSigner creates a signer
func NewSigner(key libtrust.PrivateKey, certs []*x509.Certificate) (signature.Signer, error) {
	if len(certs) == 0 {
		return nil, errors.New("missing certificates")
	}

	cert := certs[0]
	publicKey, err := libtrust.FromCryptoPublicKey(crypto.PublicKey(cert.PublicKey))
	if err != nil {
		return nil, err
	}
	if key.KeyID() != publicKey.KeyID() {
		return nil, errors.New("key and certificate mismatch")
	}

	rawCerts := make([][]byte, 0, len(certs))
	for _, cert := range certs {
		rawCerts = append(rawCerts, cert.Raw)
	}

	return &signer{
		key:      key,
		cert:     cert,
		rawCerts: rawCerts,
		hash:     crypto.SHA256,
	}, nil
}

func (s *signer) Sign(raw []byte) (signature.Signature, error) {
	if err := verifyReferences(raw, s.cert); err != nil {
		return signature.Signature{}, err
	}
	sig, alg, err := s.key.Sign(bytes.NewReader(raw), s.hash)
	if err != nil {
		return signature.Signature{}, err
	}
	return signature.Signature{
		Type:      Type,
		Algorithm: alg,
		X5c:       s.rawCerts,
		Signature: sig,
	}, nil
}
