package x509

import (
	"crypto"
	"crypto/x509"
	"errors"
	"io"
	"strings"

	"github.com/docker/go/canonical/json"
	"github.com/docker/libtrust"
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
	key, err := ReadPrivateKeyFile(keyPath)
	if err != nil {
		return nil, err
	}
	if certPath == "" {
		return NewSigner(key, nil)
	}

	certs, err := ReadCertificateFile(certPath)
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

func (s *signer) Sign(claims string) (string, []byte, error) {
	if s.cert != nil {
		if err := verifyReferences(claims, s.cert); err != nil {
			return "", nil, err
		}
	}

	// Generate header
	// We have to sign an empty string for the proper algorithm string first.
	_, alg, err := s.key.Sign(io.MultiReader(), s.hash)
	if err != nil {
		return "", nil, err
	}
	header := Header{
		Header: signature.Header{
			Type: Type,
		},
		Parameters: Parameters{
			Algorithm: alg,
		},
	}
	if s.cert != nil {
		header.X5c = s.rawCerts
	} else {
		header.KeyID = s.keyID
	}
	headerJSON, err := json.MarshalCanonical(header)
	if err != nil {
		return "", nil, err
	}

	// Generate signature
	signed := strings.Join([]string{
		signature.EncodeSegment(headerJSON),
		claims,
	}, ".")

	sig, _, err := s.key.Sign(strings.NewReader(signed), s.hash)
	if err != nil {
		return "", nil, err
	}
	return signed, sig, nil
}
