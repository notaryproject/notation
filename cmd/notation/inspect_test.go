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

	"github.com/notaryproject/notation/internal/cmd"
)

func TestInspectCommand_SecretsFromArgs(t *testing.T) {
	opts := &inspectOpts{}
	command := inspectCommand(opts)
	expected := &inspectOpts{
		reference: "ref",
		SecureFlagOpts: SecureFlagOpts{
			Password:         "password",
			InsecureRegistry: true,
			Username:         "user",
		},
		outputFormat:  cmd.OutputPlaintext,
		maxSignatures: 100,
	}
	if err := command.ParseFlags([]string{
		"--password", expected.Password,
		expected.reference,
		"-u", expected.Username,
		"--insecure-registry",
		"--output", "text"}); err != nil {
		t.Fatalf("Parse Flag failed: %v", err)
	}
	if err := command.Args(command, command.Flags().Args()); err != nil {
		t.Fatalf("Parse Args failed: %v", err)
	}
	if *opts != *expected {
		t.Fatalf("Expect inspect opts: %v, got: %v", expected, opts)
	}
}

func TestInspectCommand_SecretsFromEnv(t *testing.T) {
	t.Setenv(defaultUsernameEnv, "user")
	t.Setenv(defaultPasswordEnv, "password")
	opts := &inspectOpts{}
	expected := &inspectOpts{
		reference: "ref",
		SecureFlagOpts: SecureFlagOpts{
			Password: "password",
			Username: "user",
		},
		outputFormat:  cmd.OutputJSON,
		maxSignatures: 100,
	}
	command := inspectCommand(opts)
	if err := command.ParseFlags([]string{
		expected.reference,
		"--output", "json"}); err != nil {
		t.Fatalf("Parse Flag failed: %v", err)
	}
	if err := command.Args(command, command.Flags().Args()); err != nil {
		t.Fatalf("Parse Args failed: %v", err)
	}
	if *opts != *expected {
		t.Fatalf("Expect inspect opts: %v, got: %v", expected, opts)
	}
}

func TestInspectCommand_MissingArgs(t *testing.T) {
	command := inspectCommand(nil)
	if err := command.ParseFlags(nil); err != nil {
		t.Fatalf("Parse Flag failed: %v", err)
	}
	if err := command.Args(command, command.Flags().Args()); err == nil {
		t.Fatal("Parse Args expected error, but ok")
	}
}
