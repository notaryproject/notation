package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// File reflects the config file
type File struct {
	Enabled            bool     `json:"enabled"`
	VerificationCerts  []string `json:"verificationCerts"`
	InsecureRegistries []string `json:"insecureRegistries"`
}

// New creates a new config file
func New() *File {
	return &File{
		VerificationCerts: []string{},
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
	encoder.SetIndent("", "\t")
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
