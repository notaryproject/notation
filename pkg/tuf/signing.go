package tuf

import (
	"context"
	"errors"
)

// Signer (possibly remote) signs content
type Signer interface {
	Sign(ctx context.Context, content []byte) (Signature, error)
}

// Verifier (possibly remote) verifies content
type Verifier interface {
	Verify(ctx context.Context, content []byte, signature Signature) error
}

// Sign signs TUF metadata and appends the signature
func Sign(ctx context.Context, signer Signer, signed *Signed) error {
	sig, err := signer.Sign(ctx, *signed.Signed)
	if err != nil {
		return err
	}
	signed.Signatures = append(signed.Signatures, sig)
	return nil
}

// Verify verifies TUF metadata. Returns the number of valid signatures
func Verify(ctx context.Context, verifier Verifier, signed *Signed) (int, error) {
	var err error
	valid := 0
	for _, s := range signed.Signatures {
		err = verifier.Verify(ctx, *signed.Signed, s)
		if err == nil {
			valid++
		}
	}
	if valid == 0 {
		if err != nil {
			return 0, err
		}
		return 0, errors.New("no valid signature")
	}
	return valid, nil
}
