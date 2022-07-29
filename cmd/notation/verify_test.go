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
		certs:      []string{"cert0", "cert1"},
		certFiles:  []string{"certfile0", "certfile1"},
		signatures: []string{"sig0", "sig1"},
		pull:       true,
	}
	if err := command.ParseFlags([]string{
		expected.reference,
		"--username", expected.Username,
		"--password", expected.Password,
		"-c", expected.certs[0],
		"--cert", expected.certs[1],
		"--cert-file", expected.certFiles[0],
		"--cert-file", expected.certFiles[1],
		"--signature", expected.signatures[0],
		"-s", expected.signatures[1]}); err != nil {
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
		certs:      []string{},
		certFiles:  []string{},
		signatures: []string{},
		pull:       false,
	}
	if err := command.ParseFlags([]string{
		expected.reference,
		"--plain-http",
		"--pull=false",
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
