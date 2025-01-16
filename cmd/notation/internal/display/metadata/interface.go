package metadata

import "github.com/notaryproject/notation-core-go/signature"

type InspectHandler interface {
	// SetReference sets the artifact reference for the handler.
	SetReference(reference string)

	// SetMediaType sets the media type for the handler.
	SetMediaType(mediaType string)

	// AddSignature adds a signature to the handler.
	AddSignature(digest string, envelopeMediaType string, sigEnvelope signature.Envelope) error

	// Print prints the metadata.
	Print() error
}
