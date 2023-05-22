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
