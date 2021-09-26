package docker

import (
	"context"
	"net"
	"os/exec"

	"github.com/distribution/distribution/v3/manifest/schema2"
	"github.com/notaryproject/notation/pkg/config"
	"github.com/notaryproject/notation/pkg/docker"
	"github.com/notaryproject/notation/pkg/registry"
	"github.com/opencontainers/go-digest"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
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

// GenerateManifestOCIDescriptor generate manifest descriptor from docker save
func GenerateManifestOCIDescriptor(reference string) (ocispec.Descriptor, error) {
	manifest, err := GenerateManifest(reference)
	if err != nil {
		return ocispec.Descriptor{}, err
	}
	return ocispec.Descriptor{
		MediaType: schema2.MediaTypeManifest,
		Digest:    digest.FromBytes(manifest),
		Size:      int64(len(manifest)),
	}, nil
}

// GetManifestOCIDescriptor get manifest descriptor from remote registry
func GetManifestOCIDescriptor(ctx context.Context, ref registry.Reference) (ocispec.Descriptor, error) {
	tr, err := Transport(ref.Registry)
	if err != nil {
		return ocispec.Descriptor{}, err
	}
	insecure := config.IsRegistryInsecure(ref.Registry)
	if host, _, _ := net.SplitHostPort(ref.Registry); host == "localhost" {
		insecure = true
	}
	return registry.GetManifestDescriptor(ctx, tr, ref, insecure)
}
