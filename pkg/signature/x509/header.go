package x509

import (
	"github.com/notaryproject/nv2/pkg/signature"
)

// Header defines the signature header
type Header struct {
	signature.Header
	Parameters
}

// Parameters defines the signature parameters
type Parameters struct {
	Algorithm string   `json:"alg,omitempty"`
	KeyID     string   `json:"kid,omitempty"`
	X5c       [][]byte `json:"x5c,omitempty"`
}
