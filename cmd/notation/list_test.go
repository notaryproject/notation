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

func TestListCommand_InvalidMaxSignatures(t *testing.T) {
	opts := &listOpts{}
	cmd := listCommand(opts)
	if err := cmd.ParseFlags([]string{
		"--max-signatures", "0",
		"ref"}); err != nil {
		t.Fatalf("Parse Flag failed: %v", err)
	}
	if err := cmd.Execute(); err == nil {
		t.Fatal("Expected error for invalid max-signatures value, but got none")
	}
}

func TestListCommand_OCILayoutFlag(t *testing.T) {
	// Set the experimental flag
	t.Setenv("NOTATION_EXPERIMENTAL", "1")

	opts := &listOpts{}
	cmd := listCommand(opts)
	expected := &listOpts{
		reference: "oci-layout-ref",
		ociLayout: true,
		inputType: inputTypeOCILayout,
	}
	if err := cmd.ParseFlags([]string{
		"--oci-layout",
		expected.reference}); err != nil {
		t.Fatalf("Parse Flag failed: %v", err)
	}
	if err := cmd.Args(cmd, cmd.Flags().Args()); err != nil {
		t.Fatalf("Parse Args failed: %v", err)
	}
	if err := cmd.PreRunE(cmd, cmd.Flags().Args()); err != nil {
		t.Fatalf("PreRunE failed: %v", err)
	}
	if opts.inputType != expected.inputType {
		t.Fatalf("Expected inputType %v, got %v", expected.inputType, opts.inputType)
	}
	if opts.ociLayout != expected.ociLayout {
		t.Fatalf("Expected ociLayout %v, got %v", expected.ociLayout, opts.ociLayout)
	}
}

func TestListCommand_CustomMaxSignatures(t *testing.T) {
	opts := &listOpts{}
	cmd := listCommand(opts)
	expected := &listOpts{
		reference:     "ref",
		maxSignatures: 50,
	}
	if err := cmd.ParseFlags([]string{
		"--max-signatures", "50",
		expected.reference}); err != nil {
		t.Fatalf("Parse Flag failed: %v", err)
	}
	if err := cmd.Args(cmd, cmd.Flags().Args()); err != nil {
		t.Fatalf("Parse Args failed: %v", err)
	}
	if opts.maxSignatures != expected.maxSignatures {
		t.Fatalf("Expected maxSignatures %d, got %d", expected.maxSignatures, opts.maxSignatures)
	}
}

func TestListCommand_AllowReferrersAPIFlag(t *testing.T) {
	opts := &listOpts{}
	cmd := listCommand(opts)
	expected := &listOpts{
		reference:         "ref",
		allowReferrersAPI: true,
	}
	if err := cmd.ParseFlags([]string{
		"--allow-referrers-api",
		expected.reference}); err != nil {
		t.Fatalf("Parse Flag failed: %v", err)
	}
	if err := cmd.Args(cmd, cmd.Flags().Args()); err != nil {
		t.Fatalf("Parse Args failed: %v", err)
	}
	if opts.allowReferrersAPI != expected.allowReferrersAPI {
		t.Fatalf("Expected allowReferrersAPI %v, got %v", expected.allowReferrersAPI, opts.allowReferrersAPI)
	}
}
