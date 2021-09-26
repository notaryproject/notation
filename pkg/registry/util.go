package registry

import (
	"context"
	"net/http"

	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
)

// GetManifestDescriptor returns signature manifest information.
func GetManifestDescriptor(ctx context.Context, tr http.RoundTripper, ref Reference, plainHTTP bool) (ocispec.Descriptor, error) {
	reg := NewClient(tr, ref.Host(), plainHTTP)
	repo := reg.Repository(ctx, ref.Repository)
	return repo.GetManifestDescriptor(ctx, ref.ReferenceOrDefault())

}
