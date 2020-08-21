package tuf

import (
	"github.com/docker/go/canonical/json"
	"github.com/theupdateframework/notary/tuf/data"
)

// Signature is a signature on a piece of metadata
type Signature struct {
	KeyID     string            `json:"keyid"`
	Method    data.SigAlgorithm `json:"method"`
	Signature []byte            `json:"sig"`
	X5c       [][]byte          `json:"x5c,omitempty"`
}

// Signed is the high level, partially deserialized metadata object
// used to verify signatures before fully unpacking, or to add signatures
// before fully packing
type Signed struct {
	Signed     *json.RawMessage `json:"signed"`
	Signatures []Signature      `json:"signatures"`
}

// ToTUF converts signed to TUF
func (s Signed) ToTUF() *data.Signed {
	signatures := make([]data.Signature, 0, len(s.Signatures))
	for _, s := range s.Signatures {
		signatures = append(signatures, data.Signature{
			KeyID:     s.KeyID,
			Method:    s.Method,
			Signature: s.Signature,
		})
	}
	return &data.Signed{
		Signed:     s.Signed,
		Signatures: signatures,
	}
}

// SignedFromTUF converts signed from TUF
func SignedFromTUF(signed *data.Signed) *Signed {
	signatures := make([]Signature, 0, len(signed.Signatures))
	for _, s := range signed.Signatures {
		signatures = append(signatures, Signature{
			KeyID:     s.KeyID,
			Method:    s.Method,
			Signature: s.Signature,
		})
	}
	return &Signed{
		Signed:     signed.Signed,
		Signatures: signatures,
	}
}
