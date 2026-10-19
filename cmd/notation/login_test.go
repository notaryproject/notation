// Copyright The Notary Project Authors.
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/notaryproject/notation-go/dir"
	"github.com/notaryproject/notation/v2/cmd/notation/internal/flag"
)

func TestLoginCommand_PasswordFromArgs(t *testing.T) {
	t.Setenv(flag.EnvironmentUsername, "user")
	opts := &loginOpts{}
	cmd := loginCommand(opts)
	expected := &loginOpts{
		SecureFlagOpts: flag.SecureFlagOpts{
			Username: "user",
			Password: "password",
		},
		server: "server",
	}
	if err := cmd.ParseFlags([]string{
		expected.server,
		"-u", expected.Username,
		"-p", expected.Password,
	}); err != nil {
		t.Fatalf("Parse Flag failed: %v", err)
	}
	if err := cmd.Args(cmd, cmd.Flags().Args()); err != nil {
		t.Fatalf("Parse args failed: %v", err)
	}
	if err := cmd.PreRunE(cmd, cmd.Flags().Args()); err != nil {
		t.Fatalf("Get password failed: %v", err)
	}
	if *opts != *expected {
		t.Fatalf("Expect login opts: %v, got: %v", expected, opts)
	}
}

func TestLogin_PasswordFromStdin(t *testing.T) {
	t.Setenv(flag.EnvironmentUsername, "user")
	opts := &loginOpts{}
	cmd := loginCommand(opts)
	expected := &loginOpts{
		passwordStdin: true,
		SecureFlagOpts: flag.SecureFlagOpts{
			Username: "user",
			Password: "password",
		},
		server: "server",
	}
	if err := cmd.ParseFlags([]string{
		expected.server,
		"--password-stdin",
		"-u", expected.Username,
		"-p", expected.Password,
	}); err != nil {
		t.Fatalf("Parse Flag failed: %v", err)
	}
	if err := cmd.Args(cmd, cmd.Flags().Args()); err != nil {
		t.Fatalf("Parse Args failed: %v", err)
	}

	r, w, err := os.Pipe()
	w.Write([]byte("password"))
	w.Close()
	oldStdin := os.Stdin

	defer func() {
		os.Stdin = oldStdin
	}()
	os.Stdin = r
	if err != nil {
		t.Fatalf("Create test pipe for login cmd failed: %v", err)
	}
	if err := cmd.PreRunE(cmd, cmd.Flags().Args()); err != nil {
		t.Fatalf("Read password from stdin failed: %v", err)
	}
	if *opts != *expected {
		t.Fatalf("Expect login opts: %+v, got: %+v", expected, opts)
	}
}

func TestLoginCommand_MissingArgs(t *testing.T) {
	cmd := loginCommand(nil)
	if err := cmd.ParseFlags(nil); err != nil {
		t.Fatalf("Parse Flag failed: %v", err)
	}
	if err := cmd.Args(cmd, cmd.Flags().Args()); err == nil {
		t.Fatal("Parse Args expected error, but ok")
	}
}

// setTestConfigDir points the notation config directory to a temporary directory
// with the given config.json content.
func setTestConfigDir(t *testing.T, config string) {
	t.Helper()
	oldConfigDir := dir.UserConfigDir
	t.Cleanup(func() { dir.UserConfigDir = oldConfigDir })
	dir.UserConfigDir = t.TempDir()
	if err := os.WriteFile(filepath.Join(dir.UserConfigDir, dir.PathConfigFile), []byte(config), 0o600); err != nil {
		t.Fatalf("failed to write config file: %v", err)
	}
}

// The errors returned by runLogin have to wrap the underlying error so that
// callers can inspect it with errors.Is and errors.As.
func TestRunLogin_WrapsErrors(t *testing.T) {
	tests := []struct {
		name   string
		config string
		server string
	}{
		{
			name:   "invalid credentials store config",
			config: "{",
			server: "registry.example.com",
		},
		{
			name:   "invalid registry",
			config: "{}",
			server: "invalid host name",
		},
		{
			// nothing listens on port 1, so the connection is refused
			name:   "registry not reachable",
			config: "{}",
			server: "localhost:1",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setTestConfigDir(t, tt.config)
			opts := &loginOpts{server: tt.server}
			opts.Username = "user"
			opts.Password = "password"

			err := runLogin(context.Background(), opts)
			if err == nil {
				t.Fatal("runLogin() expected an error, but got nil")
			}
			if errors.Unwrap(err) == nil {
				t.Fatalf("runLogin() error does not wrap the underlying error: %v", err)
			}
		})
	}
}
