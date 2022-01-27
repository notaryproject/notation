package registry

import (
	"context"

	"oras.land/oras-go/v2/registry/remote"
)

// RegistryClient is a customized registry client.
type RegistryClient struct {
	base *remote.Registry
}

// NewClient creates a new registry client.
func NewClient(client remote.Client, name string, plainHTTP bool) (*RegistryClient, error) {
	reg, err := remote.NewRegistry(name)
	if err != nil {
		return nil, err
	}
	reg.Client = client
	reg.PlainHTTP = plainHTTP

	return &RegistryClient{
		base: reg,
	}, nil
}

func (r *RegistryClient) Repository(ctx context.Context, name string) (*RepositoryClient, error) {
	repo, err := r.base.Repository(ctx, name)
	if err != nil {
		return nil, err
	}
	return &RepositoryClient{
		base: repo.(*remote.Repository),
	}, nil
}
