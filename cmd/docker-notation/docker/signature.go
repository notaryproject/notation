package docker

import (
	notationregistry "github.com/notaryproject/notation/pkg/registry"
	"oras.land/oras-go/v2/registry"
)

// GetSignatureRepository returns a signature repository
func GetSignatureRepository(reference string) (notationregistry.SignatureRepository, error) {
	ref, err := registry.ParseReference(reference)
	if err != nil {
		return nil, err
	}
	return getRepositoryClient(ref)
}
