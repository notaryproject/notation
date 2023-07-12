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
	"os"
	"testing"
)

func TestLoginCommand_PasswordFromArgs(t *testing.T) {
	t.Setenv(defaultUsernameEnv, "user")
	opts := &loginOpts{}
	cmd := loginCommand(opts)
	expected := &loginOpts{
		SecureFlagOpts: SecureFlagOpts{
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
	t.Setenv(defaultUsernameEnv, "user")
	opts := &loginOpts{}
	cmd := loginCommand(opts)
	expected := &loginOpts{
		passwordStdin: true,
		SecureFlagOpts: SecureFlagOpts{
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
