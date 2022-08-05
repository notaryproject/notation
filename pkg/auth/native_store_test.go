package auth

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"testing"

	"github.com/docker/docker-credential-helpers/client"
	"github.com/docker/docker-credential-helpers/credentials"
	"github.com/notaryproject/notation-go/config"
	"oras.land/oras-go/v2/registry/remote/auth"
)

const (
	validServerAddress   = "https://index.docker.io/v1"
	validServerAddress2  = "https://example.com:5002"
	invalidServerAddress = "https://foobar.example.com"
	missingCredsAddress  = "https://missing.docker.io/v1"
	Username             = "Username"
	Secret               = "Secret"
	validUsername        = "username"
	validPassword        = "password"
	validIdentityToken   = "identityToken"
	validHelper          = "helper"
)

var (
	errCommandExited = fmt.Errorf("exited 1")
)

// mockCommand simulates interactions between the docker client and a remote
// credentials helper.
// Unit tests inject this mocked command into the remote to control execution.
type mockCommand struct {
	arg   string
	input io.Reader
}

// Output returns responses from the remote credentials helper.
// It mocks those responses based in the input in the mock.
func (m *mockCommand) Output() ([]byte, error) {
	in, err := io.ReadAll(m.input)
	if err != nil {
		return nil, err
	}
	inS := string(in)

	switch m.arg {
	case "erase":
		switch inS {
		case validServerAddress:
			return nil, nil
		default:
			return []byte("program failed"), errCommandExited
		}
	case "get":
		switch inS {
		case validServerAddress:
			return []byte(`{"Username": "username", "Secret": "password"}`), nil
		case invalidServerAddress:
			return []byte("program failed"), errCommandExited
		case validServerAddress2:
			return []byte(`{"Username": "<token>", "Secret": "identityToken"}`), nil
		}
	case "store":
		var c credentials.Credentials
		err := json.NewDecoder(strings.NewReader(inS)).Decode(&c)
		if err != nil {
			return []byte("program failed"), errCommandExited
		}
		switch c.ServerURL {
		case validServerAddress, validServerAddress2:
			return nil, nil
		default:
			return []byte("program failed"), errCommandExited
		}
	}

	return []byte(fmt.Sprintf("unknown argument %q with %q", m.arg, inS)), errCommandExited
}

// Input sets the input to send to a remote credentials helper.
func (m *mockCommand) Input(in io.Reader) {
	m.input = in
}

func mockCommandFn(args ...string) client.Program {
	return &mockCommand{
		arg: args[0],
	}
}

func TestNativeStore_StoreGetErase(t *testing.T) {
	creds := auth.Credential{
		Username: validUsername,
		Password: validPassword,
	}
	s := &nativeAuthStore{
		programFunc: mockCommandFn,
	}

	// store creds
	err := s.Store(validServerAddress, creds)
	if err != nil {
		t.Fatalf("unexpected error: %v", err.Error())
	}

	// get creds
	fetchedCreds, err := s.Get(validServerAddress)
	if err != nil {
		t.Fatalf("unexpected error: %v", err.Error())
	}
	if fetchedCreds != creds {
		t.Fatalf("expected %+v, got %+v", creds, fetchedCreds)
	}

	// erase creds
	err = s.Erase(validServerAddress)
	if err != nil {
		t.Fatalf("unexpected error: %v", err.Error())
	}
	fetchedCreds, err = s.Get(validServerAddress)
	if err != nil {
		t.Fatalf("unexpected error: %v", err.Error())
	}
	if fetchedCreds == auth.EmptyCredential {
		t.Fatalf("expect empty conf, but got: %+v", fetchedCreds)
	}
}

func TestNativeStore_StoreIdentityToken(t *testing.T) {
	creds := auth.Credential{
		RefreshToken: validIdentityToken,
	}
	s := &nativeAuthStore{
		programFunc: mockCommandFn,
	}

	// store creds
	err := s.Store(validServerAddress, creds)
	if err != nil {
		t.Fatalf("unexpected error: %v", err.Error())
	}

	// get creds
	fetchedCreds, err := s.Get(validServerAddress2)
	if err != nil {
		t.Fatalf("unexpected error: %v", err.Error())
	}
	if fetchedCreds.RefreshToken != creds.RefreshToken {
		t.Fatalf("expected %+v, got %+v", creds, fetchedCreds)
	}
}

func TestNativeStore_FailedGet(t *testing.T) {
	s := &nativeAuthStore{
		programFunc: mockCommandFn,
	}
	_, err := s.Get(invalidServerAddress)
	if err == nil {
		t.Fatalf("expect error, got nil")
	}
}

func TestNativeStore_GetCredentialsStore_LoadConfigFailed(t *testing.T) {
	loadConfig = func() (*config.Config, error) {
		return nil, fmt.Errorf("loadConfig err")
	}
	_, err := GetCredentialsStore(validServerAddress)
	if err == nil {
		t.Fatalf("expect error, got nil")
	}
}

func TestNativeStore_GetCredentialsStore_NoHelperSet(t *testing.T) {
	loadConfig = func() (*config.Config, error) {
		return &config.Config{}, nil
	}
	_, err := GetCredentialsStore(validServerAddress)
	if err == nil || err.Error() != "could not get the configured credentials store for registry: "+validServerAddress {
		t.Fatalf("Didn't get the expected error, but got: %v", err)
	}
}

func TestNativeStore_GetCredentialsStore_HelperSet(t *testing.T) {
	loadConfig = func() (*config.Config, error) {
		return &config.Config{
			CredentialHelpers: map[string]string{
				validServerAddress: validHelper,
			},
		}, nil
	}
	_, err := GetCredentialsStore(validServerAddress)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
