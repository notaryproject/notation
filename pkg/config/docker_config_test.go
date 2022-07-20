package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const (
	validJson = `{
		"credHelpers": {
			"localhost:5000": "pass"
		},
		"credsStore": "pass"
	}`
	invalidJson = `{`
)

func TestLoadDockerConfig_noErrors(t *testing.T) {
	// Create temp directory
	dockerConfigDir := t.TempDir()
	t.Setenv("DOCKER_CONFIG", dockerConfigDir)

	// Create config.json
	f, err := os.Create(filepath.Join(dockerConfigDir, dockerConfigFileName))
	if err != nil {
		t.Fatalf("Failed to mock docker config, err: %v", err)
	}
	defer f.Close()
	data := []byte(validJson)
	if _, err := f.Write(data); err != nil {
		t.Fatalf("Failed to mock docker config, err: %v", err)
	}

	// Load docker config
	config, err := LoadDockerConfig()
	if err != nil {
		t.Fatalf("Unexpected error loading config.json: %v", err)
	}
	if config.CredentialsStore != "pass" {
		t.Fatalf("Expected credentials store to be 'pass', but got %v", config.CredentialsStore)
	}
}

func TestLoadDockerConfig_noConfigFile(t *testing.T) {
	// Create temp directory
	dockerConfigDir := t.TempDir()
	t.Setenv("DOCKER_CONFIG", dockerConfigDir)

	// Load docker config
	_, err := LoadDockerConfig()
	if err == nil {
		t.Fatalf("Expected error not returned")
	}
}

func TestLoadDockerConfig_invalidConfigFile(t *testing.T) {
	// Create temp directory
	dockerConfigDir := t.TempDir()
	t.Setenv("DOCKER_CONFIG", dockerConfigDir)

	// Create config.json
	f, err := os.Create(filepath.Join(dockerConfigDir, dockerConfigFileName))
	if err != nil {
		t.Fatalf("Failed to mock docker config, err: %v", err)
	}
	defer f.Close()
	data := []byte(invalidJson)
	if _, err := f.Write(data); err != nil {
		t.Fatalf("Failed to mock docker config, err: %v", err)
	}

	// Load docker config
	_, err = LoadDockerConfig()
	if err == nil || !strings.HasSuffix(err.Error(), "unexpected EOF") {
		t.Fatalf("Expected error not returned")
	}
}
