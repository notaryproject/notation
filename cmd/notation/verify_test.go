package main

import (
	"reflect"
	"testing"
)

func TestVerifyCommand_BasicArgs(t *testing.T) {
	opts := &verifyOpts{}
	command := verifyCommand(opts)
	expected := &verifyOpts{
		reference: "ref",
		RemoteFlagOpts: RemoteFlagOpts{
			SecureFlagOpts: SecureFlagOpts{
				Username: "user",
				Password: "password",
			},
			CommonFlagOpts: CommonFlagOpts{
				MediaType: defaultMediaType,
			},
		},
	}
	if err := command.ParseFlags([]string{
		expected.reference,
		"--username", expected.Username,
		"--password", expected.Password}); err != nil {
		t.Fatalf("Parse Flag failed: %v", err)
	}
	if err := command.Args(command, command.Flags().Args()); err != nil {
		t.Fatalf("Parse args failed: %v", err)
	}
	if !reflect.DeepEqual(*expected, *opts) {
		t.Fatalf("Expect verify opts: %v, got: %v", expected, opts)
	}
}

func TestVerifyCommand_MoreArgs(t *testing.T) {
	opts := &verifyOpts{}
	command := verifyCommand(opts)
	expected := &verifyOpts{
		reference: "ref",
		RemoteFlagOpts: RemoteFlagOpts{
			SecureFlagOpts: SecureFlagOpts{
				PlainHTTP: true,
			},
			CommonFlagOpts: CommonFlagOpts{
				MediaType: "mediaT",
			},
		},
	}
	if err := command.ParseFlags([]string{
		expected.reference,
		"--plain-http",
		"--media-type=mediaT"}); err != nil {
		t.Fatalf("Parse Flag failed: %v", err)
	}
	if err := command.Args(command, command.Flags().Args()); err != nil {
		t.Fatalf("Parse args failed: %v", err)
	}
	if !reflect.DeepEqual(*expected, *opts) {
		t.Fatalf("Expect verify opts: %v, got: %v", expected, opts)
	}
}

func TestVerifyCommand_MissingArgs(t *testing.T) {
	cmd := verifyCommand(nil)
	if err := cmd.ParseFlags(nil); err != nil {
		t.Fatalf("Parse Flag failed: %v", err)
	}
	if err := cmd.Args(cmd, cmd.Flags().Args()); err == nil {
		t.Fatal("Parse Args expected error, but ok")
	}
}
