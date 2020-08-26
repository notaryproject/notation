package signature

import "errors"

// common errors
var (
	ErrInvalidToken         = errors.New("invalid token")
	ErrInvalidSignatureType = errors.New("invalid signature type")
	ErrUnknownSignatureType = errors.New("unknown signature type")
	ErrUnknownSigner        = errors.New("unknown signer")
	ErrDigestMismatch       = errors.New("digest mismatch")
	ErrSizeMismatch         = errors.New("size mismatch")
	ErrMediaTypeMismatch    = errors.New("media type mismatch")
)
