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
	"reflect"
	"testing"

	"github.com/notaryproject/notation/cmd/notation/internal/display/output"
	"github.com/notaryproject/notation/cmd/notation/internal/flag"
	"github.com/spf13/pflag"
)

func TestInspectCommand_SecretsFromArgs(t *testing.T) {
	opts := &inspectOpts{}
	command := inspectCommand(opts)
	format := flag.OutputFormatFlagOpts{}
	format.ApplyFlags(&pflag.FlagSet{}, output.FormatTree, output.FormatJSON)
	format.CurrentFormat = string(output.FormatTree)
	expected := &inspectOpts{
		reference: "ref",
		SecureFlagOpts: flag.SecureFlagOpts{
			Password:         "password",
			InsecureRegistry: true,
			Username:         "user",
		},
		outputFormat:  format,
		maxSignatures: 100,
	}
	if err := command.ParseFlags([]string{
		"--password", expected.Password,
		expected.reference,
		"-u", expected.Username,
		"--insecure-registry",
		"--output", "tree"}); err != nil {
		t.Fatalf("Parse Flag failed: %v", err)
	}
	if err := command.Args(command, command.Flags().Args()); err != nil {
		t.Fatalf("Parse Args failed: %v", err)
	}
	if !reflect.DeepEqual(opts, expected) {
		t.Fatalf("Expect opts: %v, got: %v", expected, opts)
	}
}

func TestInspectCommand_SecretsFromEnv(t *testing.T) {
	t.Setenv(flag.EnvironmentUsername, "user")
	t.Setenv(flag.EnvironmentPassword, "password")
	format := flag.OutputFormatFlagOpts{}
	format.ApplyFlags(&pflag.FlagSet{}, output.FormatTree, output.FormatJSON)
	format.CurrentFormat = string(output.FormatJSON)
	expected := &inspectOpts{
		reference: "ref",
		SecureFlagOpts: flag.SecureFlagOpts{
			Password: "password",
			Username: "user",
		},
		outputFormat:  format,
		maxSignatures: 100,
	}

	opts := &inspectOpts{}
	command := inspectCommand(opts)
	if err := command.ParseFlags([]string{
		expected.reference,
		"--output", "json"}); err != nil {
		t.Fatalf("Parse Flag failed: %v", err)
	}
	if err := command.Args(command, command.Flags().Args()); err != nil {
		t.Fatalf("Parse Args failed: %v", err)
	}
	if !reflect.DeepEqual(opts, expected) {
		t.Fatalf("Expect opts: %v, got: %v", expected, opts)
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

func TestInspectCommand_Invalid_Output(t *testing.T) {
	opts := &inspectOpts{}
	command := inspectCommand(opts)
	if err := command.ParseFlags([]string{
		"ref",
		"--output", "invalidFormat"}); err != nil {
		t.Fatalf("Parse Flag failed: %v", err)
	}
	if err := command.Args(command, command.Flags().Args()); err != nil {
		t.Fatalf("Parse Args failed: %v", err)
	}
	if err := command.PreRunE(command, command.Flags().Args()); err == nil || err.Error() != "invalid format type: \"invalidFormat\"" {
		t.Fatalf("PreRunE expected error 'invalid format type: \"invalidFormat\"', got: %v", err)
	}
	if err := command.RunE(command, command.Flags().Args()); err == nil || err.Error() != "unrecognized output format invalidFormat" {
		t.Fatalf("RunE expected error 'unrecognized output format invalidFormat', got: %v", err)
	}
}
