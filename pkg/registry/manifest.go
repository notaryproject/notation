package registry

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/distribution/distribution/v3/manifest/manifestlist"
	"github.com/distribution/distribution/v3/manifest/schema2"
	digest "github.com/opencontainers/go-digest"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	artifactspec "github.com/oras-project/artifacts-spec/specs-go/v1"
)

var supportedMediaTypes = []string{
	manifestlist.MediaTypeManifestList,
	schema2.MediaTypeManifest,
	ocispec.MediaTypeImageIndex,
	ocispec.MediaTypeImageManifest,
	artifactspec.MediaTypeArtifactManifest,
}

// GetManifestDescriptor returns signature manifest information
func (c *Client) GetManifestDescriptor(ref Reference) (ocispec.Descriptor, error) {
	scheme := "https"
	if c.plainHTTP {
		scheme = "http"
	}
	url := fmt.Sprintf("%s://%s/v2/%s/manifests/%s",
		scheme,
		ref.Host(),
		ref.Repository,
		ref.ReferenceOrDefault(),
	)
	req, err := http.NewRequest(http.MethodHead, url, nil)
	if err != nil {
		return ocispec.Descriptor{}, fmt.Errorf("invalid reference: %v", ref)
	}
	req.Header.Set("Connection", "close")
	for _, mediaType := range supportedMediaTypes {
		req.Header.Add("Accept", mediaType)
	}

	resp, err := c.base.RoundTrip(req)
	if err != nil {
		return ocispec.Descriptor{}, fmt.Errorf("%v: %v", url, err)
	}
	resp.Body.Close()
	switch resp.StatusCode {
	case http.StatusOK:
		// no op
	case http.StatusUnauthorized, http.StatusNotFound:
		return ocispec.Descriptor{}, fmt.Errorf("%v: %s", ref, resp.Status)
	default:
		return ocispec.Descriptor{}, fmt.Errorf("%v: %s", url, resp.Status)
	}

	header := resp.Header
	mediaType := header.Get("Content-Type")
	if mediaType == "" {
		return ocispec.Descriptor{}, fmt.Errorf("%v: missing Content-Type", url)
	}
	contentDigest := header.Get("Docker-Content-Digest")
	if contentDigest == "" {
		return ocispec.Descriptor{}, fmt.Errorf("%v: missing Docker-Content-Digest", url)
	}
	parsedDigest, err := digest.Parse(contentDigest)
	if err != nil {
		return ocispec.Descriptor{}, fmt.Errorf("%v: invalid Docker-Content-Digest: %s", url, contentDigest)
	}
	length := header.Get("Content-Length")
	if length == "" {
		return ocispec.Descriptor{}, fmt.Errorf("%v: missing Content-Length", url)
	}
	size, err := strconv.ParseInt(length, 10, 64)
	if err != nil {
		return ocispec.Descriptor{}, fmt.Errorf("%v: invalid Content-Length", url)
	}
	return ocispec.Descriptor{
		MediaType: mediaType,
		Digest:    parsedDigest,
		Size:      size,
	}, nil
}
