package gpg

import (
	"bytes"
	"fmt"

	"github.com/notaryproject/nv2/pkg/signature"
	"golang.org/x/crypto/openpgp"
)

type verifier struct {
	keyRing openpgp.EntityList
}

// NewVerifier creates a verifier
func NewVerifier(publicKeyRingPath string) (signature.Verifier, error) {
	keyRing, err := readKeyRingFromFile(publicKeyRingPath)
	if err != nil {
		return nil, err
	}
	return &verifier{
		keyRing: keyRing,
	}, nil
}

func (v *verifier) Type() string {
	return Type
}

func (v *verifier) Verify(content []byte, sig signature.Signature) error {
	if sig.Type != Type {
		return signature.ErrInvalidSignatureType
	}

	entity, err := openpgp.CheckDetachedSignature(
		v.keyRing,
		bytes.NewReader(content),
		bytes.NewReader(sig.Signature),
	)
	if err != nil {
		return err
	}

	if sig.Issuer != "" {
		found := false
		var signer string
		for identity := range entity.Identities {
			if identity == sig.Issuer {
				found = true
				signer = identity
				break
			}
		}
		if !found {
			return fmt.Errorf("signature verified for %q not matching the claimed issuer %q", signer, sig.Issuer)
		}
	}

	return nil
}
