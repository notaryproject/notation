package registry

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/notaryproject/notation-go-lib/signature"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	artifactspec "github.com/oras-project/artifacts-spec/specs-go/v1"
)

var supportedMediaTypes = []string{
	MediaTypeManifestList,
	MediaTypeManifest,
	ocispec.MediaTypeImageIndex,
	ocispec.MediaTypeImageManifest,
	artifactspec.MediaTypeArtifactManifest,
}

// GetManifestMetadata returns signature manifest information
func (c *Client) GetManifestMetadata(reference string) (signature.Manifest, error) {
	ref := ParseReference(reference)
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
		return signature.Manifest{}, fmt.Errorf("invalid reference: %v", reference)
	}
	req.Header.Set("Connection", "close")
	for _, mediaType := range supportedMediaTypes {
		req.Header.Add("Accept", mediaType)
	}

	resp, err := c.base.RoundTrip(req)
	if err != nil {
		return signature.Manifest{}, fmt.Errorf("%v: %v", url, err)
	}
	resp.Body.Close()
	switch resp.StatusCode {
	case http.StatusOK:
		// no op
	case http.StatusUnauthorized, http.StatusNotFound:
		return signature.Manifest{}, fmt.Errorf("%v: %s", reference, resp.Status)
	default:
		return signature.Manifest{}, fmt.Errorf("%v: %s", url, resp.Status)
	}

	header := resp.Header
	mediaType := header.Get("Content-Type")
	if mediaType == "" {
		return signature.Manifest{}, fmt.Errorf("%v: missing Content-Type", url)
	}
	digest := header.Get("Docker-Content-Digest")
	if digest == "" {
		return signature.Manifest{}, fmt.Errorf("%v: missing Docker-Content-Digest", url)
	}
	length := header.Get("Content-Length")
	if length == "" {
		return signature.Manifest{}, fmt.Errorf("%v: missing Content-Length", url)
	}
	size, err := strconv.ParseInt(length, 10, 64)
	if err != nil {
		return signature.Manifest{}, fmt.Errorf("%v: invalid Content-Length", url)
	}
	return signature.Manifest{
		Descriptor: signature.Descriptor{
			MediaType: mediaType,
			Digest:    digest,
			Size:      size,
		},
	}, nil
}
