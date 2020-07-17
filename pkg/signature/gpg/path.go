package gpg

import (
	"os"
	"path/filepath"
)

// DefaultHomePath returns the default GPG home path
func DefaultHomePath() string {
	if path, ok := os.LookupEnv("GNUPGHOME"); ok {
		return path
	}
	home, err := os.UserHomeDir()
	if err != nil {
		home = "."
	}
	return filepath.Join(home, ".gnupg")
}

// DefaultPublicKeyRingPath returns the default GPG public key ring path
func DefaultPublicKeyRingPath() string {
	return filepath.Join(DefaultHomePath(), "pubring.gpg")
}

// DefaultSecretKeyRingPath returns the default GPG secret key ring path
func DefaultSecretKeyRingPath() string {
	return filepath.Join(DefaultHomePath(), "secring.gpg")
}
