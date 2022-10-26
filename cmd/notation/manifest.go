package main

import (
	"context"
	"errors"

	"github.com/notaryproject/notation-go"
	"oras.land/oras-go/v2/registry"
)

func getManifestDescriptorFromContext(ctx context.Context, opts *SecureFlagOpts, ref string) (notation.Descriptor, error) {
	if ref == "" {
		return notation.Descriptor{}, errors.New("missing reference")
	}
	// return getManifestDescriptorFromContextWithReference(ctx, opts, ref)
	return getManifestDescriptorFromReference(ctx, opts, ref)
}

// func getManifestDescriptorFromContextWithReference(ctx context.Context, opts *RemoteFlagOpts, ref string) (notation.Descriptor, error) {
// 	if opts.Local {
// 		mediaType := opts.MediaType
// 		if ref == "-" {
// 			return getManifestDescriptorFromReader(os.Stdin, mediaType)
// 		}
// 		return getManifestDescriptorFromFile(ref, mediaType)
// 	}

// 	return getManifestDescriptorFromReference(ctx, &opts.SecureFlagOpts, ref)
// }

func getManifestDescriptorFromReference(ctx context.Context, opts *SecureFlagOpts, reference string) (notation.Descriptor, error) {
	ref, err := registry.ParseReference(reference)
	if err != nil {
		return notation.Descriptor{}, err
	}
	repo, err := getRepositoryClient(opts, ref)
	if err != nil {
		return notation.Descriptor{}, err
	}
	return repo.Resolve(ctx, ref.ReferenceOrDefault())
}

// func getManifestDescriptorFromFile(path, mediaType string) (notation.Descriptor, error) {
// 	file, err := os.Open(path)
// 	if err != nil {
// 		return notation.Descriptor{}, err
// 	}
// 	defer file.Close()
// 	return getManifestDescriptorFromReader(file, mediaType)
// }

// func getManifestDescriptorFromReader(r io.Reader, mediaType string) (notation.Descriptor, error) {
// 	lr := &io.LimitedReader{
// 		R: r,
// 		N: math.MaxInt64,
// 	}
// 	digest, err := digest.SHA256.FromReader(lr)
// 	if err != nil {
// 		return notation.Descriptor{}, err
// 	}
// 	return notation.Descriptor{
// 		MediaType: mediaType,
// 		Digest:    digest,
// 		Size:      math.MaxInt64 - lr.N,
// 	}, nil
// }
