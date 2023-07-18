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
	"testing"
)

func TestListCommand_SecretsFromArgs(t *testing.T) {
	opts := &listOpts{}
	cmd := listCommand(opts)
	expected := &listOpts{
		reference: "ref",
		SecureFlagOpts: SecureFlagOpts{
			Password:         "password",
			InsecureRegistry: true,
			Username:         "user",
		},
		maxSignatures: 100,
	}
	if err := cmd.ParseFlags([]string{
		"--password", expected.Password,
		expected.reference,
		"-u", expected.Username,
		"--insecure-registry"}); err != nil {
		t.Fatalf("Parse Flag failed: %v", err)
	}
	if err := cmd.Args(cmd, cmd.Flags().Args()); err != nil {
		t.Fatalf("Parse Args failed: %v", err)
	}
	if *opts != *expected {
		t.Fatalf("Expect list opts: %v, got: %v", expected, opts)
	}
}

func TestListCommand_SecretsFromEnv(t *testing.T) {
	t.Setenv(defaultUsernameEnv, "user")
	t.Setenv(defaultPasswordEnv, "password")
	opts := &listOpts{}
	expected := &listOpts{
		reference: "ref",
		SecureFlagOpts: SecureFlagOpts{
			Password: "password",
			Username: "user",
		},
		maxSignatures: 100,
	}
	cmd := listCommand(opts)
	if err := cmd.ParseFlags([]string{
		expected.reference}); err != nil {
		t.Fatalf("Parse Flag failed: %v", err)
	}
	if err := cmd.Args(cmd, cmd.Flags().Args()); err != nil {
		t.Fatalf("Parse Args failed: %v", err)
	}
	if *opts != *expected {
		t.Fatalf("Expect list opts: %v, got: %v", expected, opts)
	}
}

func TestListCommand_MissingArgs(t *testing.T) {
	cmd := listCommand(nil)
	if err := cmd.ParseFlags(nil); err != nil {
		t.Fatalf("Parse Flag failed: %v", err)
	}
	if err := cmd.Args(cmd, cmd.Flags().Args()); err == nil {
		t.Fatal("Parse Args expected error, but ok")
	}
}
