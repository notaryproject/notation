package config

import (
	"path/filepath"

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

// SignaturePath returns the path of a signature for a manifest
func SignaturePath(manifestDigest digest.Digest) string {
	return filepath.Join(
		SignatureStoreDirPath,
		manifestDigest.Algorithm().String(),
		manifestDigest.Encoded()+SignatureExtension,
	)
}
