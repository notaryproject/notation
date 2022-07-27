package main

import (
	"testing"
)

func TestPullCommand(t *testing.T) {
	t.Setenv(defaultUsernameEnv, "user")
	t.Setenv(defaultPasswordEnv, "password")

	opts := &pullOpts{}
	cmd := pullCommand(opts)
	expected := &pullOpts{
		reference: "ref",
		strict:    true,
		SecureFlagOpts: SecureFlagOpts{
			Username: "user2",
			Password: "password",
		},
	}
	if err := cmd.ParseFlags([]string{
		expected.reference,
		"-u", expected.Username,
		"--strict"}); err != nil {
		t.Fatalf("Parse Flag failed: %v", err)
	}
	if err := cmd.Args(cmd, cmd.Flags().Args()); err != nil {
		t.Fatalf("Parse args failed: %v", err)
	}
	if *expected != *opts {
		t.Fatalf("Expect pull opts: %v, got: %v", expected, opts)
	}
}

func TestPullCommand_MissingArgs(t *testing.T) {
	cmd := pullCommand(nil)
	if err := cmd.ParseFlags(nil); err != nil {
		t.Fatalf("Parse Flag failed: %v", err)
	}
	if err := cmd.Args(cmd, cmd.Flags().Args()); err == nil {
		t.Fatal("Parse Args expected error, but ok")
	}
}
