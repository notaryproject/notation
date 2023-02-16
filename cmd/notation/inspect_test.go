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
			Password:  "password",
			PlainHTTP: true,
			Username:  "user",
		},
		outputFormat: cmd.OutputPlaintext,
	}
	if err := command.ParseFlags([]string{
		"--password", expected.Password,
		expected.reference,
		"-u", expected.Username,
		"--plain-http",
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
		outputFormat: cmd.OutputJSON,
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
