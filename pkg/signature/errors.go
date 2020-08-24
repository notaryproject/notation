package signature

import "errors"

// common errors
var (
	ErrInvalidSignatureType = errors.New("invalid signature type")
	ErrUnknownSignatureType = errors.New("unknown signature type")
	ErrUnknownSigner        = errors.New("unknown signer")
)
