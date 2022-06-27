package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

const (
	// dockerConfigFileName is the name of config file
	dockerConfigFileName = "config.json"
	dockerConfigFileDir  = ".docker"
)

// DockerConfigFile is the minimized configuration of the Docker daemon, only
// credentails store related configs are included
type DockerConfigFile struct {
	CredentialsStore  string            `json:"credsStore,omitempty"`
	CredentialHelpers map[string]string `json:"credHelpers,omitempty"`
}

// Load reads the configuration files in the given directory, and sets up
// the auth config information and returns values.
func LoadDockerConfig() (*DockerConfigFile, error) {
	configFile := &DockerConfigFile{}
	configDir, err := getDockerConfigDir()
	if err != nil {
		return configFile, err
	}

	filename := filepath.Join(configDir, dockerConfigFileName)

	// load latest config file
	file, err := os.Open(filename)
	if err != nil {
		return configFile, fmt.Errorf("%s: %w", filename, err)
	}

	defer file.Close()
	err = configFile.loadFromReader(file)
	if err != nil {
		err = fmt.Errorf("%s: %w", filename, err)
	}
	return configFile, err
}

func getDockerConfigDir() (string, error) {
	configDir := os.Getenv("DOCKER_CONFIG")
	if configDir == "" {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("%s, %w", "Could not get home directory", err)
		}
		configDir = filepath.Join(homeDir, dockerConfigFileDir)
	}
	return configDir, nil
}

// loadFromReader reads the configuration data given and sets up the auth config
// information with given directory and populates the receiver object
func (configFile *DockerConfigFile) loadFromReader(configData io.Reader) error {
	if err := json.NewDecoder(configData).Decode(configFile); err != nil && !errors.Is(err, io.EOF) {
		return err
	}
	return nil
}
