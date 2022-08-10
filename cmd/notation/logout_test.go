package main

import "testing"

func TestLogoutCommand_BasicArgs(t *testing.T) {
	opts := &logoutOpts{}
	cmd := logoutCommand(opts)
	expected := &logoutOpts{
		server: "server",
	}
	if err := cmd.ParseFlags([]string{
		expected.server,
	}); err != nil {
		t.Fatalf("Parse Flag failed: %v", err)
	}
	if err := cmd.Args(cmd, cmd.Flags().Args()); err != nil {
		t.Fatalf("Parse Args failed: %v", err)
	}
	if *opts != *expected {
		t.Fatalf("Expect logout opts: %v, got: %v", expected, opts)
	}
}

func TestLogOutCommand_MissingArgs(t *testing.T) {
	cmd := logoutCommand(nil)
	if err := cmd.ParseFlags([]string{}); err != nil {
		t.Fatalf("Parse Flag failed: %v", err)
	}
	if err := cmd.Args(cmd, cmd.Flags().Args()); err == nil {
		t.Fatal("Parse Args expected error, but ok")
	}
}
