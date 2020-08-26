package signature

// Signer signs content
type Signer interface {
	Sign(claims string) (string, []byte, error)
}

// Verifier verifies content
type Verifier interface {
	Type() string
	Verify(header Header, signed string, sig []byte) error
}
