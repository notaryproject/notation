package registry

import (
	"context"
	"net/http"

	"github.com/notaryproject/notation-go-lib"
)

// GetManifestDescriptor returns signature manifest information.
func GetManifestDescriptor(ctx context.Context, tr http.RoundTripper, ref Reference, plainHTTP bool) (notation.Descriptor, error) {
	reg := NewClient(tr, ref.Host(), plainHTTP)
	repo := reg.Repository(ctx, ref.Repository)
	return repo.GetManifestDescriptor(ctx, ref.ReferenceOrDefault())
}
