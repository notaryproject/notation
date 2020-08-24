package gpg

import (
	"bytes"

	"github.com/notaryproject/nv2/pkg/signature"
	"golang.org/x/crypto/openpgp"
)

type signer struct {
	entity   *openpgp.Entity
	identity string
}

// NewSigner creates a signer
func NewSigner(secretKeyRingPath, identity string) (signature.Signer, error) {
	entity, identity, err := findEntityFromFile(secretKeyRingPath, identity)
	if err != nil {
		return nil, err
	}
	return &signer{
		entity:   entity,
		identity: identity,
	}, nil
}

func (s *signer) Sign(raw []byte) (signature.Signature, error) {
	sig := bytes.NewBuffer(nil)
	if err := openpgp.DetachSign(sig, s.entity, bytes.NewReader(raw), nil); err != nil {
		return signature.Signature{}, err
	}
	return signature.Signature{
		Type:      Type,
		Issuer:    s.identity,
		Signature: sig.Bytes(),
	}, nil
}
