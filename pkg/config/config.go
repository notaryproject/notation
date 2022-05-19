package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// X509KeyPair contains the paths of a public/private key pair files.
type X509KeyPair struct {
	KeyPath         string `json:"keyPath,omitempty"`
	CertificatePath string `json:"certPath,omitempty"`
}

// ExternalKey contains the necessary information to delegate
// the signing operation to the named plugin.
type ExternalKey struct {
	ID           string            `json:"id,omitempty"`
	PluginName   string            `json:"pluginName,omitempty"`
	PluginConfig map[string]string `json:"pluginConfig,omitempty"`
}

// KeySuite is a named key suite.
type KeySuite struct {
	Name string `json:"name"`

	*X509KeyPair
	*ExternalKey
}

func (k KeySuite) Is(name string) bool {
	return k.Name == name
}

// CertificateReference is a named file path.
type CertificateReference struct {
	Name string `json:"name"`
	Path string `json:"path"`
}

func (c CertificateReference) Is(name string) bool {
	return c.Name == name
}

// File reflects the config file.
// Specification: https://github.com/notaryproject/notation/pull/76
type File struct {
	VerificationCertificates VerificationCertificates `json:"verificationCerts"`
	SigningKeys              SigningKeys              `json:"signingKeys,omitempty"`
	InsecureRegistries       []string                 `json:"insecureRegistries"`
}

// VerificationCertificates is a collection of public certs used for verification.
type VerificationCertificates struct {
	Certificates []CertificateReference `json:"certs"`
}

// SigningKeys is a collection of signing keys.
type SigningKeys struct {
	Default string     `json:"default"`
	Keys    []KeySuite `json:"keys"`
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
