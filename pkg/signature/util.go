package signature

import (
	"encoding/json"
	"errors"
)

// Pack packs content with its signatures
func Pack(content Content, signatures ...Signature) (Signed, error) {
	signed, err := json.Marshal(content)
	if err != nil {
		return Signed{}, err
	}
	if len(signatures) == 0 {
		return Signed{}, errors.New("missing signatures")
	}
	return Signed{
		Signed:     signed,
		Signatures: signatures,
	}, nil
}
