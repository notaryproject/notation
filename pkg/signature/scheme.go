package signature

import (
	"fmt"
	"strings"
	"time"

	"github.com/docker/go/canonical/json"
)

// Scheme is a signature scheme
type Scheme struct {
	signers   map[string]Signer
	verifiers map[string]Verifier
}

// NewScheme creates a new scheme
func NewScheme() *Scheme {
	return &Scheme{
		signers:   make(map[string]Signer),
		verifiers: make(map[string]Verifier),
	}
}

// RegisterSigner registers signer with a name
func (s *Scheme) RegisterSigner(signerID string, signer Signer) {
	s.signers[signerID] = signer
}

// RegisterVerifier registers verifier
func (s *Scheme) RegisterVerifier(verifier Verifier) {
	s.verifiers[verifier.Type()] = verifier
}

// Sign signs claims by a signer
func (s *Scheme) Sign(signerID string, claims Claims) (string, error) {
	bytes, err := json.MarshalCanonical(claims)
	if err != nil {
		return "", err
	}
	return s.SignRaw(signerID, bytes)
}

// SignRaw signs raw content by a signer
func (s *Scheme) SignRaw(signerID string, content []byte) (string, error) {
	signer, found := s.signers[signerID]
	if !found {
		return "", ErrUnknownSigner
	}

	signed, sig, err := signer.Sign(EncodeSegment(content))
	if err != nil {
		return "", err
	}

	return strings.Join([]string{
		signed,
		EncodeSegment(sig),
	}, "."), nil
}

// Verify verifies the JWT-like token
func (s *Scheme) Verify(token string) (Claims, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return Claims{}, ErrInvalidToken
	}

	if err := s.verifySignature(parts); err != nil {
		return Claims{}, err
	}

	claims, err := DecodeClaims(parts[1])
	if err != nil {
		return Claims{}, err
	}

	return claims, s.verifyClaims(claims)
}

func (s *Scheme) verifySignature(parts []string) error {
	rawHeader, err := DecodeSegment(parts[0])
	if err != nil {
		return ErrInvalidToken
	}
	var header Header
	if json.Unmarshal(rawHeader, &header); err != nil {
		return ErrInvalidToken
	}
	header.Raw = rawHeader

	verifier, found := s.verifiers[header.Type]
	if !found {
		return ErrUnknownSignatureType
	}

	sig, err := DecodeSegment(parts[2])
	if err != nil {
		return ErrInvalidToken
	}

	return verifier.Verify(
		header,
		strings.Join(parts[:2], "."),
		sig,
	)
}

func (s *Scheme) verifyClaims(claims Claims) error {
	now := time.Now().Unix()
	if claims.Expiration != 0 && now > claims.Expiration {
		return fmt.Errorf("content expired: %d: current: %d", claims.Expiration, now)
	}
	if claims.NotBefore != 0 && now < claims.NotBefore {
		return fmt.Errorf("content is not available yet: %d: current: %d", claims.NotBefore, now)
	}
	return nil
}
