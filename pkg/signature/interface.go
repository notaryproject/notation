package signature

// Signer signs content
type Signer interface {
	Sign(content []byte) (Signature, error)
}

// Verifier verifies content
type Verifier interface {
	Type() string
	Verify(content []byte, signature Signature) error
}
