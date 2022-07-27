package main

import (
	"os"
	"testing"
)

func TestLoginCommand(t *testing.T) {
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
