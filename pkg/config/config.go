package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// File reflects the config file.
// Specification: https://github.com/notaryproject/notation/pull/76
type File struct {
	VerificationCertificates VerificationCertificates `json:"verificationCerts"`
	SigningKeys              SigningKeys              `json:"signingKeys,omitempty"`
	InsecureRegistries       []string                 `json:"insecureRegistries"`
	KMSPlugins               KMSPlugins               `json:"kmsPlugins"`
}

// VerificationCertificates is a collection of public certs used for verification.
type VerificationCertificates struct {
	Certificates CertificateMap `json:"certs"`
	KMSCerts     KMSProfileMap  `json:"kmsCerts"`
}

// SigningKeys is a collection of signing keys.
type SigningKeys struct {
	Default string        `json:"default"`
	Keys    KeyMap        `json:"keys"`
	KMSKeys KMSProfileMap `json:"kmsKeys"`
}

// KMSPlugins is a collection of plugins.
type KMSPlugins struct {
	Plugins PluginMap `json:"plugins"`
}

// New creates a new config file
func New() *File {
	return &File{
		InsecureRegistries: []string{},
	}
}

// Save stores the config to file
func (f *File) Save() error {
	dir := filepath.Dir(FilePath)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return err
	}
	file, err := os.Create(FilePath)
	if err != nil {
		return err
	}
	defer file.Close()
	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "    ")
	return encoder.Encode(f)
}

// Load reads the config from file
func Load() (*File, error) {
	file, err := os.Open(FilePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	var config *File
	if err := json.NewDecoder(file).Decode(&config); err != nil {
		return nil, err
	}
	return config, nil
}

// LoadOrDefault reads the config from file or return a default config if not found.
func LoadOrDefault() (*File, error) {
	file, err := Load()
	if err != nil {
		if os.IsNotExist(err) {
			return New(), nil
		}
		return nil, err
	}
	return file, nil
}
