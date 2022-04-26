package registry

import (
	"context"

	"github.com/notaryproject/notation-go"
	"github.com/opencontainers/go-digest"
)

// SignatureRepository provides a storage for signatures
type SignatureRepository interface {
	// Lookup finds all signatures for the specified manifest
	Lookup(ctx context.Context, manifestDigest digest.Digest) ([]digest.Digest, error)

	// Get downloads the signature by the specified digest
	Get(ctx context.Context, signatureDigest digest.Digest) ([]byte, error)

	// Put uploads the signature to the registry
	Put(ctx context.Context, signature []byte) (notation.Descriptor, error)

	// Link creates an signature artifact linking the manifest and the signature
	Link(ctx context.Context, manifest, signature notation.Descriptor) (notation.Descriptor, error)
}
