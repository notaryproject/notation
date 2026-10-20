// Copyright The Notary Project Authors.
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package registryutil provides OCI registry blob fetching utilities.
package registryutil

import (
	"context"
	"encoding/json"
	"fmt"
	"io"

	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"oras.land/oras-go/v2/registry"
	"oras.land/oras-go/v2/registry/remote"
)

// BlobFetcher fetches blobs from an OCI registry.
type BlobFetcher struct {
	remoteRepo *remote.Repository
	reference  registry.Reference
}

// NewBlobFetcher creates a new BlobFetcher. The remoteRepo should have
// authentication already configured.
func NewBlobFetcher(ctx context.Context, reference string, remoteRepo *remote.Repository) (*BlobFetcher, error) {
	ref, err := registry.ParseReference(reference)
	if err != nil {
		return nil, fmt.Errorf("failed to parse reference %s: %w", reference, err)
	}

	return &BlobFetcher{
		remoteRepo: remoteRepo,
		reference:  ref,
	}, nil
}

// FetchManifest fetches the manifest from the registry and extracts it into an ocispec.Manifest struct.
// With a json object returned by the registry, we can iterate through the layers and fetch each layer blob to compute the root hash and signature.
func (f *BlobFetcher) FetchManifest(ctx context.Context, manifestDesc ocispec.Descriptor) (*ocispec.Manifest, error) {
	reader, err := f.remoteRepo.Fetch(ctx, manifestDesc)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch manifest blob: %w", err)
	}
	defer reader.Close()

	manifestBytes, err := io.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("failed to read manifest content: %w", err)
	}

	var manifest ocispec.Manifest
	if err := json.Unmarshal(manifestBytes, &manifest); err != nil {
		return nil, fmt.Errorf("failed to unmarshal manifest JSON: %w", err)
	}

	return &manifest, nil
}

// FetchBlob fetches a blob by descriptor and returns its raw content.
func (f *BlobFetcher) FetchBlob(ctx context.Context, desc ocispec.Descriptor) ([]byte, error) {
	reader, err := f.remoteRepo.Fetch(ctx, desc)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch blob %s: %w", desc.Digest, err)
	}
	defer reader.Close()

	content, err := io.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("failed to read blob content: %w", err)
	}

	return content, nil
}
