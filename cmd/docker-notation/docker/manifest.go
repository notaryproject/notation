package docker

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os/exec"
	"strconv"
	"strings"

	"github.com/distribution/distribution/v3/manifest/schema2"
	"github.com/notaryproject/notation/pkg/config"
	"github.com/notaryproject/notation/pkg/docker"
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
func GetManifestOCIDescriptor(ctx context.Context, hostname, repository, ref string) (ocispec.Descriptor, error) {
	tr, err := Transport(hostname)
	if err != nil {
		return ocispec.Descriptor{}, err
	}

	scheme := "https"
	if config.IsRegistryInsecure(hostname) {
		scheme = "http"
	}
	if host, _, _ := net.SplitHostPort(hostname); host == "localhost" {
		scheme = "http"
	}
	url := fmt.Sprintf("%s://%s/v2/%s/manifests/%s",
		scheme,
		hostname,
		repository,
		ref,
	)
	req, err := http.NewRequest(http.MethodHead, url, nil)
	if err != nil {
		return ocispec.Descriptor{}, err
	}
	req.Header.Set("Connection", "close")
	req.Header.Set("Accept", schema2.MediaTypeManifest)

	resp, err := tr.RoundTrip(req)
	if err != nil {
		return ocispec.Descriptor{}, fmt.Errorf("%v: %v", url, err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return ocispec.Descriptor{}, fmt.Errorf("%v: %s", url, resp.Status)
	}

	header := resp.Header
	mediaType := header.Get("Content-Type")
	if mediaType != schema2.MediaTypeManifest {
		return ocispec.Descriptor{}, fmt.Errorf("%v: media type mismatch: %s", url, mediaType)
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
		MediaType: schema2.MediaTypeManifest,
		Digest:    parsedDigest,
		Size:      size,
	}, nil
}

// GetManifestReference returns the tag or the digest of the reference string
func GetManifestReference(ref string) string {
	if index := strings.Index(ref, "@"); index != -1 {
		return ref[index+1:]
	} else if index := strings.LastIndex(ref, ":"); index != -1 {
		return ref[index+1:]
	} else {
		return "latest"
	}
}
