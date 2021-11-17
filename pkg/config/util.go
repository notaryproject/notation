package config

import (
	"errors"
	"strings"
)

var (
	// ErrKeyNotFound indicates that the signing key is not found.
	ErrKeyNotFound = errors.New("signing key not found")

	// ErrCertificateNotFound indicates that the verification certificate is not found.
	ErrCertificateNotFound = errors.New("verification certificate not found")

	// ErrKMSPluginNotFound indicates that the KMS plugin is not found.
	ErrPluginNotFound = errors.New("plugin not found")
)

// IsRegistryInsecure checks whether the registry is in the list of insecure registries.
func IsRegistryInsecure(target string) bool {
	config, err := LoadOrDefaultOnce()
	if err != nil {
		return false
	}
	for _, registry := range config.InsecureRegistries {
		if strings.EqualFold(registry, target) {
			return true
		}
	}
	return false
}

// ResolveKeyPath resolves the key path by name along with
// its corresponding certificate path.
// The default key is attempted if name is empty.
func ResolveKeyPath(name string) (string, string, error) {
	config, err := LoadOrDefaultOnce()
	if err != nil {
		return "", "", err
	}
	if name == "" {
		name = config.SigningKeys.Default
	}
	keyPath, certPath, ok := config.SigningKeys.Keys.Get(name)
	if !ok {
		return "", "", ErrKeyNotFound
	}
	return keyPath, certPath, nil
}

// ResolveCertificatePath resolves the certificate path by name.
func ResolveCertificatePath(name string) (string, error) {
	config, err := LoadOrDefaultOnce()
	if err != nil {
		return "", err
	}
	path, ok := config.VerificationCertificates.Certificates.Get(name)
	if !ok {
		return "", ErrCertificateNotFound
	}
	return path, nil
}

// ResolveKMSKeyID resolves the KMS key ID by name along
// with its corresponding plugin name.
// The default key is attempted if name is empty.
func ResolveKMSKey(name string) (KMSKeySuite, error) {
	key := KMSKeySuite{}
	config, err := LoadOrDefaultOnce()
	if err != nil {
		return key, err
	}
	if name == "" {
		name = config.SigningKeys.Default
	}
	kmsKey, ok := config.SigningKeys.KMSKeys.Get(name)
	if !ok {
		return key, ErrKeyNotFound
	}
	return kmsKey, nil
}

func ResolveKMSPluginPath(name string) (string, error) {
	config, err := LoadOrDefaultOnce()
	if err != nil {
		return "", err
	}
	path, ok := config.KMSPlugins.Plugins.Get(name)
	if !ok {
		return "", ErrPluginNotFound
	}
	return path, nil
}
