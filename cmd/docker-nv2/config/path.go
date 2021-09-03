package config

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/docker/cli/cli/config"
	"github.com/opencontainers/go-digest"
)

const (
	// FileName is the name of config file
	FileName = "nv2.json"

	// SignatureStoreDirName is the name of the signature store directory
	SignatureStoreDirName = "nv2"

	// SignatureExtension defines the extension of the signature files
	SignatureExtension = ".jwt"
)

var (
	// FilePath is the path of config file
	FilePath = filepath.Join(config.Dir(), FileName)
	// SignatureStoreDirPath is the path of the signature store
	SignatureStoreDirPath = filepath.Join(config.Dir(), SignatureStoreDirName)
)

// SignatureRootPath returns the root path of signatures for a manifest
func SignatureRootPath(manifestDigest digest.Digest) string {
	return filepath.Join(
		SignatureStoreDirPath,
		manifestDigest.Algorithm().String(),
		manifestDigest.Encoded(),
	)
}

// SignaturePath returns the path of a signature for a manifest
func SignaturePath(manifestDigest, signatureDigest digest.Digest) string {
	return filepath.Join(
		SignatureRootPath(manifestDigest),
		signatureDigest.Algorithm().String(),
		signatureDigest.Encoded()+SignatureExtension,
	)
}

// SignatureDigests returns the digest of signatures for a manifest
func SignatureDigests(manifestDigest digest.Digest) ([]digest.Digest, error) {
	rootPath := SignatureRootPath(manifestDigest)
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
			if !strings.HasSuffix(encoded, SignatureExtension) {
				continue
			}
			encoded = strings.TrimSuffix(encoded, SignatureExtension)
			digest := digest.NewDigestFromEncoded(digest.Algorithm(algorithm), encoded)
			if err := digest.Validate(); err != nil {
				return nil, err
			}
			digests = append(digests, digest)
		}
	}
	return digests, nil
}
