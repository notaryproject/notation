package envelope

import (
	"errors"

	"github.com/notaryproject/notation-core-go/signature/cose"
	"github.com/notaryproject/notation-core-go/signature/jws"
	gcose "github.com/veraison/go-cose"
)

// Supported envelope format.
const (
	COSE = "cose"
	JWS  = "jws"
)

// SpeculateSignatureEnvelopeFormat speculates envelope format by looping all builtin envelope format.
//
// TODO: do we nned to speculate evenlope format?
//
// Both push and verify will need to speculate format now not only for caching.
//
// For verifying, the new verification workflow with trust store
// and trust policy doesn't need to speculate format because it will pull from remote.
//
// For pushing, there is a flag named signatures, which is the paths to some signatures. CLI will first read the file and push its content to the remote.
//
// For RC1, verify doesn't need speculate, but we still need to speculate format if user pushes a signature from local file system
func SpeculateSignatureEnvelopeFormat(raw []byte) (string, error) {
	var msg gcose.Sign1Message
	if err := msg.UnmarshalCBOR(raw); err == nil {
		return cose.MediaTypeEnvelope, nil
	}
	if len(raw) == 0 || raw[0] != '{' {
		// very certain
		return "", errors.New("unsupported signature format")
	}
	return jws.MediaTypeEnvelope, nil
}
