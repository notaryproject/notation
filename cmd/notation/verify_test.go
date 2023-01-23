package main

import (
	"reflect"
	"testing"

	"github.com/notaryproject/notation/internal/ioutil"
)

func TestVerifyCommand_BasicArgs(t *testing.T) {
	opts := &verifyOpts{}
	command := verifyCommand(opts)
	expected := &verifyOpts{
		reference: "ref",
		SecureFlagOpts: SecureFlagOpts{
			Username: "user",
			Password: "password",
		},
		pluginConfig: []string{"key1=val1"},
		outputFormat: ioutil.OutputPlaintext,
	}
	if err := command.ParseFlags([]string{
		expected.reference,
		"--username", expected.Username,
		"--password", expected.Password,
		"--plugin-config", "key1=val1"}); err != nil {
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
		SecureFlagOpts: SecureFlagOpts{
			PlainHTTP: true,
		},
		pluginConfig: []string{"key1=val1", "key2=val2"},
		outputFormat: ioutil.OutputJson,
	}
	if err := command.ParseFlags([]string{
		expected.reference,
		"--plain-http",
		"--plugin-config", "key1=val1",
		"--plugin-config", "key2=val2",
		"--output", "json"}); err != nil {
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
