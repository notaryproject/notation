package registry

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/distribution/distribution/v3/manifest/manifestlist"
	"github.com/distribution/distribution/v3/manifest/schema2"
	"github.com/notaryproject/notation-go-lib"
	"github.com/notaryproject/notation/internal/ioutil"
	"github.com/opencontainers/go-digest"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	artifactspec "github.com/oras-project/artifacts-spec/specs-go/v1"
)

const (
	maxBlobSizeLimit     = 32 * 1024 * 1024 // 32 MiB
	maxManifestSizeLimit = 4 * 1024 * 1024  // 4 MiB
	maxMetadataReadLimit = 4 * 1024 * 1024  // 4 MiB
)

var supportedMediaTypes = []string{
	manifestlist.MediaTypeManifestList,
	schema2.MediaTypeManifest,
	ocispec.MediaTypeImageIndex,
	ocispec.MediaTypeImageManifest,
	artifactspec.MediaTypeArtifactManifest,
}

type RepositoryClient struct {
	tr   http.RoundTripper
	base string
	name string
}

// GetManifestDescriptor returns signature manifest information by tag or digest.
func (r *RepositoryClient) GetManifestDescriptor(ctx context.Context, ref string) (notation.Descriptor, error) {
	url := fmt.Sprintf("%s/v2/%s/manifests/%s", r.base, r.name, ref)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return notation.Descriptor{}, fmt.Errorf("invalid reference: %v", ref)
	}
	req.Header.Set("Connection", "close")
	for _, mediaType := range supportedMediaTypes {
		req.Header.Add("Accept", mediaType)
	}

	resp, err := r.tr.RoundTrip(req)
	if err != nil {
		return notation.Descriptor{}, fmt.Errorf("%v: %v", url, err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return notation.Descriptor{}, fmt.Errorf("%v: %s", url, resp.Status)
	}

	header := resp.Header
	mediaType := header.Get("Content-Type")
	if mediaType == "" {
		return notation.Descriptor{}, fmt.Errorf("%v: missing Content-Type", url)
	}
	contentDigest := header.Get("Docker-Content-Digest")
	if contentDigest == "" {
		return notation.Descriptor{}, fmt.Errorf("%v: missing Docker-Content-Digest", url)
	}
	parsedDigest, err := digest.Parse(contentDigest)
	if err != nil {
		return notation.Descriptor{}, fmt.Errorf("%v: invalid Docker-Content-Digest: %s", url, contentDigest)
	}
	length := header.Get("Content-Length")
	if length == "" {
		return notation.Descriptor{}, fmt.Errorf("%v: missing Content-Length", url)
	}
	size, err := strconv.ParseInt(length, 10, 64)
	if err != nil {
		return notation.Descriptor{}, fmt.Errorf("%v: invalid Content-Length", url)
	}
	return notation.Descriptor{
		MediaType: mediaType,
		Digest:    parsedDigest,
		Size:      size,
	}, nil
}

func (r *RepositoryClient) Lookup(ctx context.Context, manifestDigest digest.Digest) ([]digest.Digest, error) {
	url, err := url.Parse(fmt.Sprintf("%s/oras/artifacts/v1/%s/manifests/%s/referrers", r.base, r.name, manifestDigest.String()))
	if err != nil {
		return nil, err
	}
	q := url.Query()
	q.Add("artifactType", ArtifactTypeNotation)
	url.RawQuery = q.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url.String(), nil)
	if err != nil {
		return nil, err
	}
	resp, err := r.tr.RoundTrip(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to lookup signatures: %s", resp.Status)
	}

	reader := io.LimitReader(resp.Body, maxMetadataReadLimit)
	result := struct {
		References []artifactspec.Descriptor `json:"references"`
	}{}
	if err := json.NewDecoder(reader).Decode(&result); err != nil {
		return nil, err
	}
	digests := make([]digest.Digest, 0, len(result.References))
	for _, desc := range result.References {
		if desc.ArtifactType != ArtifactTypeNotation || desc.MediaType != artifactspec.MediaTypeArtifactManifest {
			continue
		}
		artifact, err := r.getArtifactManifest(ctx, desc.Digest)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch manifest: %v: %v", desc.Digest, err)
		}
		for _, blob := range artifact.Blobs {
			digests = append(digests, blob.Digest)
		}
	}
	return digests, nil
}

func (r *RepositoryClient) Get(ctx context.Context, signatureDigest digest.Digest) ([]byte, error) {
	return r.getBlob(ctx, signatureDigest)
}

func (r *RepositoryClient) Put(ctx context.Context, signature []byte) (notation.Descriptor, error) {
	desc := DescriptorFromBytes(signature)
	desc.MediaType = MediaTypeNotationSignature
	return desc, r.putBlob(ctx, signature, desc.Digest)
}

func (r *RepositoryClient) Link(ctx context.Context, manifest, signature notation.Descriptor) (notation.Descriptor, error) {
	artifact := artifactspec.Manifest{
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
	desc := DescriptorFromBytes(artifactJSON)
	return desc, r.putManifest(ctx, artifactJSON, desc.Digest)
}

func (r *RepositoryClient) getBlob(ctx context.Context, digest digest.Digest) ([]byte, error) {
	url := fmt.Sprintf("%s/v2/%s/blobs/%s", r.base, r.name, digest.String())
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	resp, err := r.tr.RoundTrip(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusOK {
		return ioutil.ReadAllVerified(io.LimitReader(resp.Body, maxBlobSizeLimit), digest)
	}
	if resp.StatusCode != http.StatusTemporaryRedirect {
		return nil, fmt.Errorf("failed to get blob: %s", resp.Status)
	}
	resp.Body.Close()

	location, err := resp.Location()
	if err != nil {
		return nil, err
	}
	req, err = http.NewRequestWithContext(ctx, http.MethodGet, location.String(), nil)
	if err != nil {
		return nil, err
	}
	resp, err = r.tr.RoundTrip(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get blob: %s", resp.Status)
	}
	return ioutil.ReadAllVerified(io.LimitReader(resp.Body, maxBlobSizeLimit), digest)
}

func (r *RepositoryClient) putBlob(ctx context.Context, blob []byte, digest digest.Digest) error {
	url := fmt.Sprintf("%s/v2/%s/blobs/uploads/", r.base, r.name)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, nil)
	if err != nil {
		return err
	}
	resp, err := r.tr.RoundTrip(req)
	if err != nil {
		return err
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusAccepted {
		return fmt.Errorf("failed to init upload: %s", resp.Status)
	}

	url = resp.Header.Get("Location")
	if url == "" {
		return http.ErrNoLocation
	}

	if !strings.HasPrefix(url, r.base) {
		url = fmt.Sprintf("%s%s", r.base, url)
	}

	req, err = http.NewRequestWithContext(ctx, http.MethodPut, url, bytes.NewReader(blob))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/octet-stream")
	q := req.URL.Query()
	q.Add("digest", digest.String())
	req.URL.RawQuery = q.Encode()
	resp, err = r.tr.RoundTrip(req)
	if err != nil {
		return err
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("failed to upload: %s", resp.Status)
	}
	return nil
}

func (r *RepositoryClient) putManifest(ctx context.Context, blob []byte, digest digest.Digest) error {
	url := fmt.Sprintf("%s/v2/%s/manifests/%s", r.base, r.name, digest.String())
	req, err := http.NewRequestWithContext(ctx, http.MethodPut, url, bytes.NewReader(blob))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", artifactspec.MediaTypeArtifactManifest)
	resp, err := r.tr.RoundTrip(req)
	if err != nil {
		return err
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("failed to put manifest: %s", resp.Status)
	}
	return nil
}

func (r *RepositoryClient) getManifest(ctx context.Context, mediaType string, digest digest.Digest) ([]byte, error) {
	url := fmt.Sprintf("%s/v2/%s/manifests/%s", r.base, r.name, digest.String())
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", mediaType)
	resp, err := r.tr.RoundTrip(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get manifest: %s", resp.Status)
	}
	return ioutil.ReadAllVerified(io.LimitReader(resp.Body, maxManifestSizeLimit), digest)
}

func (r *RepositoryClient) getArtifactManifest(ctx context.Context, digest digest.Digest) (artifactspec.Manifest, error) {
	manifestJSON, err := r.getManifest(ctx, artifactspec.MediaTypeArtifactManifest, digest)
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
