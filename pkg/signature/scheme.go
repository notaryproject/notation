package signature

import (
	"encoding/json"
	"fmt"
	"time"
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

// Sign signs content by a signer
func (s *Scheme) Sign(signerID string, content Content) (Signature, error) {
	bytes, err := json.Marshal(content)
	if err != nil {
		return Signature{}, err
	}
	return s.SignRaw(signerID, bytes)
}

// SignRaw signs raw content by a signer
func (s *Scheme) SignRaw(signerID string, content []byte) (Signature, error) {
	signer, found := s.signers[signerID]
	if !found {
		return Signature{}, ErrUnknownSigner
	}
	return signer.Sign(content)
}

// Verify verifies signed data
func (s *Scheme) Verify(signed Signed) (Content, Signature, error) {
	sig, err := s.verifySignature(signed)
	if err != nil {
		return Content{}, sig, err
	}

	var content Content
	if err := json.Unmarshal(signed.Signed, &content); err != nil {
		return Content{}, sig, err
	}

	return content, sig, s.verifyContent(content)
}

func (s *Scheme) verifySignature(signed Signed) (Signature, error) {
	sig := signed.Signature
	verifier, found := s.verifiers[sig.Type]
	if !found {
		return Signature{}, ErrUnknownSignatureType
	}

	content := []byte(signed.Signed)
	if err := verifier.Verify(content, sig); err != nil {
		return Signature{}, err
	}

	return sig, nil
}

func (s *Scheme) verifyContent(content Content) error {
	now := time.Now().Unix()
	if content.Expiration != 0 && now > content.Expiration {
		return fmt.Errorf("content expired: %d: current: %d", content.Expiration, now)
	}
	if content.NotBefore != 0 && now < content.NotBefore {
		return fmt.Errorf("content is not available yet: %d: current: %d", content.NotBefore, now)
	}
	return nil
}
