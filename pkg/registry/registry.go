package registry

import (
	"context"
	"fmt"
	"net/http"
)

// RegistryClient is a customized registry client.
type RegistryClient struct {
	tr   http.RoundTripper
	base string
}

// NewClient creates a new registry client.
func NewClient(tr http.RoundTripper, name string, plainHTTP bool) *RegistryClient {
	if tr == nil {
		tr = http.DefaultTransport
	}
	scheme := "https"
	if plainHTTP {
		scheme = "http"
	}
	return &RegistryClient{
		tr:   tr,
		base: fmt.Sprintf("%s://%s", scheme, name),
	}
}

func (r *RegistryClient) Repository(ctx context.Context, name string) *RepositoryClient {
	return &RepositoryClient{
		tr:   r.tr,
		base: r.base,
		name: name,
	}
}
