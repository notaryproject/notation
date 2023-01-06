package envelope

import (
	"fmt"

	"github.com/notaryproject/notation-core-go/signature/cose"
	"github.com/notaryproject/notation-core-go/signature/jws"
)

// Supported envelope format.
const (
	COSE = "cose"
	JWS  = "jws"
)

// GetEnvelopeMediaType converts the envelope type to mediaType name.
func GetEnvelopeMediaType(sigFormat string) (string, error) {
	switch sigFormat {
	case JWS:
		return jws.MediaTypeEnvelope, nil
	case COSE:
		return cose.MediaTypeEnvelope, nil
	}
	return "", fmt.Errorf("signature format %q not supported", sigFormat)
}
