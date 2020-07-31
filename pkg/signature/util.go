package signature

import (
	"encoding/json"
)

// Pack packs content with its signature
func Pack(content Content, signature Signature) (Signed, error) {
	signed, err := json.Marshal(content)
	if err != nil {
		return Signed{}, err
	}
	return Signed{
		Signed:    signed,
		Signature: signature,
	}, nil
}
