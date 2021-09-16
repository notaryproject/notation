package docker

import (
	"context"
	"net"

	"github.com/distribution/distribution/v3/reference"
	"github.com/notaryproject/notation-go-lib"
	"github.com/notaryproject/notation-go-lib/registry"
	"github.com/notaryproject/notation/pkg/config"
)

// GetSignatureRepository returns a signature repository
func GetSignatureRepository(ctx context.Context, ref string) (notation.SignatureRepository, error) {
	named, err := reference.ParseNamed(ref)
	if err != nil {
		return nil, err
	}
	hostname, repository := reference.SplitHostname(named)

	tr, err := Transport(hostname)
	if err != nil {
		return nil, err
	}

	insecure := config.IsRegistryInsecure(hostname)
	if host, _, _ := net.SplitHostPort(hostname); host == "localhost" {
		insecure = true
	}
	client := registry.NewClient(tr, hostname, insecure)

	return client.Repository(ctx, repository), nil
}
