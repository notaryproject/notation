package main

import (
	"reflect"
	"testing"
)

func TestPushCommand(t *testing.T) {
	t.Setenv(defaultUsernameEnv, "user")
	t.Setenv(defaultPasswordEnv, "password")

	opts := &pushOpts{}
	cmd := pushCommand(opts)
	expected := &pushOpts{
		reference: "ref",
		SecureFlagOpts: SecureFlagOpts{
			Username: "user",
			Password: "password2",
		},
		signatures: []string{"s0", "s1"},
	}
	if err := cmd.ParseFlags([]string{
		expected.reference,
		"-p", expected.Password,
		"--signature", expected.signatures[0],
		"-s", expected.signatures[1]}); err != nil {
		t.Fatalf("Parse Flag failed: %v", err)
	}
	if err := cmd.Args(cmd, cmd.Flags().Args()); err != nil {
		t.Fatalf("Parse args failed: %v", err)
	}
	if !reflect.DeepEqual(*expected, *opts) {
		t.Fatalf("Expect key remove opts: %v, got: %v", expected, opts)
	}
}

func TestPushCommand_MissingArgs(t *testing.T) {
	cmd := pushCommand(nil)
	if err := cmd.ParseFlags(nil); err != nil {
		t.Fatalf("Parse Flag failed: %v", err)
	}
	if err := cmd.Args(cmd, cmd.Flags().Args()); err == nil {
		t.Fatal("Parse Args expected error, but ok")
	}
}
