package registry

import (
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/notaryproject/nv2/pkg/signature"
	v1 "github.com/opencontainers/image-spec/specs-go/v1"
)

// GetManifestMetadata returns signature manifest information by URI scheme
func (c *Client) GetManifestMetadata(uri *url.URL) (signature.Manifest, error) {
	switch scheme := strings.ToLower(uri.Scheme); scheme {
	case "docker":
		return c.GetDockerManifestMetadata(uri)
	case "oci":
		return c.GetOCIManifestMetadata(uri)
	default:
		return signature.Manifest{}, fmt.Errorf("unsupported scheme: %s", scheme)
	}
}

// GetDockerManifestMetadata returns signature manifest information
// from a remote Docker manifest
func (c *Client) GetDockerManifestMetadata(uri *url.URL) (signature.Manifest, error) {
	return c.getManifestMetadata(uri,
		MediaTypeManifestList,
		MediaTypeManifest,
	)
}

// GetOCIManifestMetadata returns signature manifest information
// from a remote OCI manifest
func (c *Client) GetOCIManifestMetadata(uri *url.URL) (signature.Manifest, error) {
	return c.getManifestMetadata(uri,
		v1.MediaTypeImageIndex,
		v1.MediaTypeImageManifest,
	)
}

// GetManifestMetadata returns signature manifest information
func (c *Client) getManifestMetadata(uri *url.URL, mediaTypes ...string) (signature.Manifest, error) {
	host := uri.Host
	if host == "docker.io" {
		host = "registry-1.docker.io"
	}
	var repository string
	var reference string
	path := strings.TrimPrefix(uri.Path, "/")
	if index := strings.Index(path, "@"); index != -1 {
		repository = path[:index]
		reference = path[index+1:]
	} else if index := strings.Index(path, ":"); index != -1 {
		repository = path[:index]
		reference = path[index+1:]
	} else {
		repository = path
		reference = "latest"
	}
	scheme := "https"
	if c.insecure {
		scheme = "http"
	}
	url := fmt.Sprintf("%s://%s/v2/%s/manifests/%s",
		scheme,
		host,
		repository,
		reference,
	)
	req, err := http.NewRequest(http.MethodHead, url, nil)
	if err != nil {
		return signature.Manifest{}, fmt.Errorf("invalid uri: %v", uri)
	}
	req.Header.Set("Connection", "close")
	for _, mediaType := range mediaTypes {
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
		return signature.Manifest{}, fmt.Errorf("%v: %s", uri, resp.Status)
	default:
		return signature.Manifest{}, fmt.Errorf("%v: %s", url, resp.Status)
	}

	header := resp.Header
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
		Digest: digest,
		Size:   size,
	}, nil
}
