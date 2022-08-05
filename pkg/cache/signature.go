package cache

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/notaryproject/notation-go/dir"
	"github.com/opencontainers/go-digest"
)

// SignatureDigests returns the digest of signatures for a manifest
func SignatureDigests(manifestDigest digest.Digest) ([]digest.Digest, error) {
	rootPath := dir.Path.CachedSignatureRoot(manifestDigest)
	algorithmEntries, err := os.ReadDir(rootPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	var digests []digest.Digest
	for _, algorithmEntry := range algorithmEntries {
		if !algorithmEntry.Type().IsDir() {
			continue
		}

		algorithm := algorithmEntry.Name()
		signatureEntries, err := os.ReadDir(filepath.Join(rootPath, algorithm))
		if err != nil {
			return nil, err
		}

		for _, signatureEntry := range signatureEntries {
			if !signatureEntry.Type().IsRegular() {
				continue
			}
			encoded := signatureEntry.Name()
			if !strings.HasSuffix(encoded, dir.SignatureExtension) {
				continue
			}
			encoded = strings.TrimSuffix(encoded, dir.SignatureExtension)
			digest := digest.NewDigestFromEncoded(digest.Algorithm(algorithm), encoded)
			if err := digest.Validate(); err != nil {
				return nil, err
			}
			digests = append(digests, digest)
		}
	}
	return digests, nil
}
