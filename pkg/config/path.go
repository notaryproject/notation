package config

import (
	"os"
	"path/filepath"

	"github.com/opencontainers/go-digest"
)

const (
	// ApplicationName is the name of the application
	ApplicationName = "notation"

	// FileName is the name of config file
	FileName = "config.json"

	// SignatureStoreDirName is the name of the signature store directory
	SignatureStoreDirName = "signature"

	// SignatureExtension defines the extension of the signature files
	SignatureExtension = ".sig"

	// KeyStoreDirName is the name of the key store directory
	KeyStoreDirName = "key"

	// KeyExtension defines the extension of the key files
	KeyExtension = ".key"

	// CertificateStoreDirName is the name of the certificate store directory
	CertificateStoreDirName = "certificate"

	// CertificateExtension defines the extension of the certificate files
	CertificateExtension = ".crt"

	// PluginStoreDirName is the name of the plugin store directory
	PluginStoreDirName = "plugins"
)

var (
	// FilePath is the path of config file
	FilePath string

	// SignatureStoreDirPath is the path of the signature store
	SignatureStoreDirPath string

	// KeyStoreDirPath is the path of the key store
	KeyStoreDirPath string

	// CertificateStoreDirPath is the path of the certificate store
	CertificateStoreDirPath string

	// PluginDirPath is the path of the plugin store
	PluginDirPath string
)

// init initialize the essential file paths
func init() {
	// init home directories
	configDir, err := os.UserConfigDir()
	if err != nil {
		panic(err)
	}
	configDir = filepath.Join(configDir, ApplicationName)
	cacheDir, err := os.UserCacheDir()
	if err != nil {
		panic(err)
	}
	cacheDir = filepath.Join(cacheDir, ApplicationName)

	// init paths
	FilePath = filepath.Join(configDir, FileName)
	SignatureStoreDirPath = filepath.Join(cacheDir, SignatureStoreDirName)
	KeyStoreDirPath = filepath.Join(configDir, KeyStoreDirName)
	CertificateStoreDirPath = filepath.Join(configDir, CertificateStoreDirName)
	PluginDirPath = filepath.Join(configDir, PluginStoreDirName)
}

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

// KeyPath returns the path of a signing key
func KeyPath(name string) string {
	return filepath.Join(KeyStoreDirPath, name+KeyExtension)
}

// CertificatePath returns the path of a certificate for verification
func CertificatePath(name string) string {
	return filepath.Join(CertificateStoreDirPath, name+CertificateExtension)
}
