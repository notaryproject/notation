package envelope

import (
	"encoding/json"
	"errors"

	"github.com/notaryproject/notation-core-go/signature/cose"
	"github.com/notaryproject/notation-core-go/signature/jws"
	gcose "github.com/veraison/go-cose"
)

// TODO:
// don't know how to get envelope format if passed a local signature file
// or verify with local signature
// now hard code
// but need a way to find out
func ParseSigEnvelopeFormat(raw []byte) (string, error) {
	var body interface{}
	if err := json.Unmarshal(raw, &body); err == nil {
		return jws.MediaTypeEnvelope, nil
	}
	msg := gcose.NewSign1Message()
	if err := msg.UnmarshalCBOR(raw); err == nil {
		return cose.MediaTypeEnvelope, nil
	}
	return "", errors.New("Unsupported signature format")
}
