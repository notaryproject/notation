package auth

import (
	"fmt"
	"testing"

	"github.com/notaryproject/notation-go/config"
	"github.com/notaryproject/notation/pkg/configutil"
)

const (
	errMsg     = "error message"
	validStore = "pass"
)

func TestLoadConfig_LoadNotationConfigFailed(t *testing.T) {
	loadOrDefault = func() (*config.Config, error) {
		return nil, fmt.Errorf(errMsg)
	}
	_, err := LoadConfig()
	if err == nil || err.Error() != errMsg {
		t.Fatalf("Didn't get the expected error, but got: %v", err)
	}
}

func TestLoadConfig_NotationConfigContainsAuth(t *testing.T) {
	loadOrDefault = func() (*config.Config, error) {
		return &config.Config{CredentialsStore: validStore}, nil
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
	loadOrDefault = func() (*config.Config, error) {
		return nil, nil
	}
	loadDockerConfig = func() (*configutil.DockerConfigFile, error) {
		return nil, fmt.Errorf(errMsg)
	}
	_, err := LoadConfig()
	if err == nil || err.Error() != errMsg {
		t.Fatalf("Didn't get the expected error, but got: %v", err)
	}
}

func TestLoadConfig_DockerConfigContainsAuth(t *testing.T) {
	loadOrDefault = func() (*config.Config, error) {
		return nil, nil
	}
	loadDockerConfig = func() (*configutil.DockerConfigFile, error) {
		return &configutil.DockerConfigFile{
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
	loadOrDefault = func() (*config.Config, error) {
		return nil, nil
	}
	loadDockerConfig = func() (*configutil.DockerConfigFile, error) {
		return &configutil.DockerConfigFile{}, nil
	}
	_, err := LoadConfig()
	if err == nil {
		t.Fatalf("expect error but got nil")
	}
}
