package auth

import (
	"fmt"
	"testing"

	"github.com/notaryproject/notation/pkg/config"
)

const (
	errMsg     = "error message"
	validStore = "pass"
)

func TestLoadConfig_LoadNotationConfigFailed(t *testing.T) {
	loadOrDefault = func() (*config.File, error) {
		return nil, fmt.Errorf(errMsg)
	}
	_, err := LoadConfig()
	if err == nil || err.Error() != errMsg {
		t.Fatalf("Didn't get the expected error, but got: %v", err)
	}
}

func TestLoadConfig_NotationConfigContainsAuth(t *testing.T) {
	loadOrDefault = func() (*config.File, error) {
		return &config.File{
			CredentialsStore: validStore,
		}, nil
	}
	file, err := LoadConfig()
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if file == nil || file.CredentialsStore != validStore {
		t.Fatalf("Should contain auth")
	}
}

func TestLoadConfig_LoadDockerConfigFailed(t *testing.T) {
	loadOrDefault = func() (*config.File, error) {
		return nil, nil
	}
	loadDockerConfig = func() (*config.DockerConfigFile, error) {
		return nil, fmt.Errorf(errMsg)
	}
	_, err := LoadConfig()
	if err == nil || err.Error() != errMsg {
		t.Fatalf("Didn't get the expected error, but got: %v", err)
	}
}

func TestLoadConfig_DockerConfigContainsAuth(t *testing.T) {
	loadOrDefault = func() (*config.File, error) {
		return nil, nil
	}
	loadDockerConfig = func() (*config.DockerConfigFile, error) {
		return &config.DockerConfigFile{
			CredentialsStore: validStore,
		}, nil
	}
	file, err := LoadConfig()
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if file == nil || file.CredentialsStore != validStore {
		t.Fatalf("Should contain auth")
	}
}

func TestLoadConfig_DockerConfigEmptyAuth(t *testing.T) {
	loadOrDefault = func() (*config.File, error) {
		return nil, nil
	}
	loadDockerConfig = func() (*config.DockerConfigFile, error) {
		return &config.DockerConfigFile{}, nil
	}
	_, err := LoadConfig()
	if err == nil {
		t.Fatalf("expect error but got nil")
	}
}
