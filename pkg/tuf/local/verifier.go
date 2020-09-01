package local

import (
	"context"
	"crypto/x509"
	"errors"
	"fmt"
	"strings"

	"github.com/docker/go/canonical/json"
	"github.com/notaryproject/nv2/pkg/tuf"
	"github.com/theupdateframework/notary/tuf/data"
	"github.com/theupdateframework/notary/tuf/signed"
	"github.com/theupdateframework/notary/tuf/utils"
)

type verifier struct {
	keys  map[string]data.PublicKey
	certs map[string]*x509.Certificate
	roots *x509.CertPool
}

// NewVerifier creates a verifier
func NewVerifier(certs []*x509.Certificate, roots *x509.CertPool) (tuf.Verifier, error) {
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

	keys := make(map[string]data.PublicKey, len(certs))
	keyedCerts := make(map[string]*x509.Certificate, len(certs))
	for _, cert := range certs {
		key := utils.CertToKey(cert)
		if key == nil {
			return nil, errors.New("unknown certificate key type")
		}
		keyID, err := utils.CanonicalKeyID(key)
		if err != nil {
			return nil, err
		}
		keys[keyID] = key
		keyedCerts[keyID] = cert
	}

	return &verifier{
		keys:  keys,
		certs: keyedCerts,
		roots: roots,
	}, nil
}

func (v *verifier) Verify(ctx context.Context, content []byte, sig tuf.Signature) error {
	key, cert, err := v.getVerificationKeyPair(sig)
	if err != nil {
		return err
	}
	alg, ok := signed.Verifiers[sig.Method]
	if !ok {
		return fmt.Errorf("signing method is not supported: %s", sig.Method)
	}
	if err := alg.Verify(key, sig.Signature, content); err != nil {
		return err
	}
	return verifyReferences(content, cert)
}

func (v *verifier) getVerificationKeyPair(sig tuf.Signature) (data.PublicKey, *x509.Certificate, error) {
	if len(sig.X5c) > 0 {
		return v.getVerificationKeyPairFromX5c(sig.KeyID, sig.X5c)
	}
	return v.getVerificationKeyPairFromKeyID(sig.KeyID)
}

func (v *verifier) getVerificationKeyPairFromKeyID(keyID string) (data.PublicKey, *x509.Certificate, error) {
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

func (v *verifier) getVerificationKeyPairFromX5c(claimedKeyID string, x5c [][]byte) (data.PublicKey, *x509.Certificate, error) {
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

	key := utils.CertToKey(cert)
	if key == nil {
		return nil, nil, errors.New("unknown certificate key type")
	}
	// Docker Notary 0.6.0 implementation uses non-canonical key ID for delegation roles,
	// which should be canonical.
	keyID := key.ID()
	if keyID != claimedKeyID {
		return nil, nil, errors.New("certificate key ID mismatch")
	}

	return key, cert, nil
}

func verifyReferences(raw []byte, cert *x509.Certificate) error {
	// Skip unrecognizable contents
	var targets data.Targets
	if err := json.Unmarshal(raw, &targets); err != nil {
		return nil
	}
	if targets.Type != data.TUFTypes[data.CanonicalTargetsRole] {
		return nil
	}

	roots := x509.NewCertPool()
	roots.AddCert(cert)
	for reference := range targets.Targets {
		domain := strings.SplitN(reference, "/", 2)[0]
		domain = strings.SplitN(domain, ":", 2)[0]
		if _, err := cert.Verify(x509.VerifyOptions{
			DNSName: domain,
			Roots:   roots,
		}); err != nil {
			return err
		}
	}
	return nil
}
