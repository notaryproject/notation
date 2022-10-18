package main

import (
	"context"
	"fmt"

	"github.com/notaryproject/notation-go"
	"github.com/notaryproject/notation/internal/envelope"
)

func pushSignature(ctx context.Context, opts *SecureFlagOpts, ref string, sig []byte) (notation.Descriptor, error) {
	// initialize
	sigRepo, err := getSignatureRepository(opts, ref)
	if err != nil {
		return notation.Descriptor{}, err
	}
	manifestDesc, err := getManifestDescriptorFromReference(ctx, opts, ref)
	if err != nil {
		return notation.Descriptor{}, err
	}

	// core process
	// pass in nonempty annotations if needed
	sigMediaType, err := envelope.SpeculateSignatureEnvelopeFormat(sig)
	if err != nil {
		return notation.Descriptor{}, err
	}
	sigDesc, _, err := sigRepo.PutSignatureManifest(ctx, sig, sigMediaType, manifestDesc, make(map[string]string))
	if err != nil {
		return notation.Descriptor{}, fmt.Errorf("put signature manifest failure: %v", err)
	}

	return sigDesc, nil
}
