package docker

import (
	"context"
	"os/exec"

	"github.com/distribution/distribution/v3/manifest/schema2"
	"github.com/notaryproject/notation-go"
	"github.com/notaryproject/notation/pkg/docker"
	"github.com/opencontainers/go-digest"
	"oras.land/oras-go/v2/registry"
)

// GenerateManifest generate manifest from docker save
func GenerateManifest(reference string) ([]byte, error) {
	cmd := exec.Command("docker", "save", reference)
	reader, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}
	if err := cmd.Start(); err != nil {
		return nil, err
	}

	manifest, err := docker.GenerateSchema2FromDockerSave(reader)
	if err != nil {
		return nil, err
	}
	_, payload, err := manifest.Payload()
	return payload, err
}

// GenerateManifestDescriptor generate manifest descriptor from docker save
func GenerateManifestDescriptor(reference string) (notation.Descriptor, error) {
	manifest, err := GenerateManifest(reference)
	if err != nil {
		return notation.Descriptor{}, err
	}
	return notation.Descriptor{
		MediaType: schema2.MediaTypeManifest,
		Digest:    digest.FromBytes(manifest),
		Size:      int64(len(manifest)),
	}, nil
}

// GetManifestDescriptor get manifest descriptor from remote registry
func GetManifestDescriptor(ctx context.Context, ref registry.Reference) (notation.Descriptor, error) {
	client, err := getRepositoryClient(ref)
	if err != nil {
		return notation.Descriptor{}, err
	}
	return client.GetManifestDescriptor(ctx, ref.ReferenceOrDefault())
}
