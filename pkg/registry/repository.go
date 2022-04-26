package registry

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"

	"github.com/notaryproject/notation-go"
	"github.com/opencontainers/go-digest"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	artifactspec "github.com/oras-project/artifacts-spec/specs-go/v1"
	"oras.land/oras-go/v2/content"
	"oras.land/oras-go/v2/registry"
	"oras.land/oras-go/v2/registry/remote"
)

const (
	maxBlobSizeLimit     = 32 * 1024 * 1024 // 32 MiB
	maxManifestSizeLimit = 4 * 1024 * 1024  // 4 MiB
)

type RepositoryClient struct {
	remote.Repository
}

// NewRepositoryClient creates a new registry client.
func NewRepositoryClient(client remote.Client, ref registry.Reference, plainHTTP bool) *RepositoryClient {
	return &RepositoryClient{
		Repository: remote.Repository{
			Client:    client,
			Reference: ref,
			PlainHTTP: plainHTTP,
		},
	}
}

// GetManifestDescriptor returns signature manifest information by tag or digest.
func (c *RepositoryClient) GetManifestDescriptor(ctx context.Context, ref string) (notation.Descriptor, error) {
	desc, err := c.Repository.Resolve(ctx, ref)
	if err != nil {
		return notation.Descriptor{}, err
	}
	return notationDescriptorFromOCI(desc), nil
}

// Lookup finds all signatures for the specified manifest
func (c *RepositoryClient) Lookup(ctx context.Context, manifestDigest digest.Digest) ([]digest.Digest, error) {
	var digests []digest.Digest
	// TODO(shizhMSFT): filter artifact type at the server side
	if err := c.Repository.Referrers(ctx, ocispec.Descriptor{
		Digest: manifestDigest,
	}, func(referrers []artifactspec.Descriptor) error {
		for _, desc := range referrers {
			if desc.ArtifactType != ArtifactTypeNotation || desc.MediaType != artifactspec.MediaTypeArtifactManifest {
				continue
			}
			artifact, err := c.getArtifactManifest(ctx, desc.Digest)
			if err != nil {
				return fmt.Errorf("failed to fetch manifest: %v: %v", desc.Digest, err)
			}
			for _, blob := range artifact.Blobs {
				digests = append(digests, blob.Digest)
			}
		}
		return nil
	}); err != nil {
		return nil, err
	}
	return digests, nil
}

// Get downloads the signature by the specified digest
func (c *RepositoryClient) Get(ctx context.Context, signatureDigest digest.Digest) ([]byte, error) {
	desc, err := c.Repository.Resolve(ctx, signatureDigest.String())
	if err != nil {
		return nil, err
	}
	if desc.Size > maxBlobSizeLimit {
		return nil, fmt.Errorf("signature blob too large: %d", desc.Size)
	}
	return content.FetchAll(ctx, c.Repository.Blobs(), desc)
}

// Put uploads the signature to the registry
func (c *RepositoryClient) Put(ctx context.Context, signature []byte) (notation.Descriptor, error) {
	desc := ocispec.Descriptor{
		MediaType: MediaTypeNotationSignature,
		Digest:    digest.FromBytes(signature),
		Size:      int64(len(signature)),
	}
	if err := c.Repository.Blobs().Push(ctx, desc, bytes.NewReader(signature)); err != nil {
		return notation.Descriptor{}, err
	}
	return notationDescriptorFromOCI(desc), nil
}

// Link creates an signature artifact linking the manifest and the signature
func (c *RepositoryClient) Link(ctx context.Context, manifest, signature notation.Descriptor) (notation.Descriptor, error) {
	// generate artifact manifest
	artifact := artifactspec.Manifest{
		MediaType:    artifactspec.MediaTypeArtifactManifest,
		ArtifactType: ArtifactTypeNotation,
		Blobs: []artifactspec.Descriptor{
			artifactDescriptorFromNotation(signature),
		},
		Subject: artifactDescriptorFromNotation(manifest),
	}
	artifactJSON, err := json.Marshal(artifact)
	if err != nil {
		return notation.Descriptor{}, err
	}

	// upload manifest
	desc := ocispec.Descriptor{
		MediaType: artifactspec.MediaTypeArtifactManifest,
		Digest:    digest.FromBytes(artifactJSON),
		Size:      int64(len(artifactJSON)),
	}
	if err := c.Repository.Manifests().Push(ctx, desc, bytes.NewReader(artifactJSON)); err != nil {
		return notation.Descriptor{}, err
	}
	return notationDescriptorFromOCI(desc), nil
}

func (c *RepositoryClient) getArtifactManifest(ctx context.Context, manifestDigest digest.Digest) (artifactspec.Manifest, error) {
	repo := c.Repository
	repo.ManifestMediaTypes = []string{
		artifactspec.MediaTypeArtifactManifest,
	}
	store := repo.Manifests()
	desc, err := store.Resolve(ctx, manifestDigest.String())
	if err != nil {
		return artifactspec.Manifest{}, err
	}
	if desc.Size > maxManifestSizeLimit {
		return artifactspec.Manifest{}, fmt.Errorf("manifest too large: %d", desc.Size)
	}
	manifestJSON, err := content.FetchAll(ctx, store, desc)
	if err != nil {
		return artifactspec.Manifest{}, err
	}

	var manifest artifactspec.Manifest
	err = json.Unmarshal(manifestJSON, &manifest)
	if err != nil {
		return artifactspec.Manifest{}, err
	}
	return manifest, nil
}

func artifactDescriptorFromNotation(desc notation.Descriptor) artifactspec.Descriptor {
	return artifactspec.Descriptor{
		MediaType: desc.MediaType,
		Digest:    desc.Digest,
		Size:      desc.Size,
	}
}

func notationDescriptorFromOCI(desc ocispec.Descriptor) notation.Descriptor {
	return notation.Descriptor{
		MediaType: desc.MediaType,
		Digest:    desc.Digest,
		Size:      desc.Size,
	}
}
