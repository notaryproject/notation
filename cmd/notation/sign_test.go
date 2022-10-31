package main

import (
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/notaryproject/notation/internal/cmd"
	"github.com/notaryproject/notation/internal/envelope"
)

func TestSignCommand_BasicArgs(t *testing.T) {
	opts := &signOpts{}
	command := signCommand(opts)
	expected := &signOpts{
		reference: "ref",
		SecureFlagOpts: SecureFlagOpts{
			Username: "user",
			Password: "password",
		},
		SignerFlagOpts: cmd.SignerFlagOpts{
			Key:          "key",
			EnvelopeType: envelope.JWS,
		},
	}
	if err := command.ParseFlags([]string{
		expected.reference,
		"-u", expected.Username,
		"--password", expected.Password,
		"--key", expected.Key}); err != nil {
		t.Fatalf("Parse Flag failed: %v", err)
	}
	if err := command.Args(command, command.Flags().Args()); err != nil {
		t.Fatalf("Parse args failed: %v", err)
	}
	if !reflect.DeepEqual(*expected, *opts) {
		t.Fatalf("Expect sign opts: %v, got: %v", expected, opts)
	}
}

func TestSignCommand_MoreArgs(t *testing.T) {
	opts := &signOpts{}
	command := signCommand(opts)
	expected := &signOpts{
		reference: "ref",
		SecureFlagOpts: SecureFlagOpts{
			Username:  "user",
			Password:  "password",
			PlainHTTP: true,
		},
		SignerFlagOpts: cmd.SignerFlagOpts{
			Key:          "key",
			EnvelopeType: envelope.COSE,
		},
		expiry: 24 * time.Hour,
	}
	if err := command.ParseFlags([]string{
		expected.reference,
		"-u", expected.Username,
		"-p", expected.Password,
		"--key", expected.Key,
		"--plain-http",
		"--signature-format", expected.SignerFlagOpts.EnvelopeType,
		"--expiry", expected.expiry.String()}); err != nil {
		t.Fatalf("Parse Flag failed: %v", err)
	}
	if err := command.Args(command, command.Flags().Args()); err != nil {
		t.Fatalf("Parse args failed: %v", err)
	}
	if !reflect.DeepEqual(*expected, *opts) {
		t.Fatalf("Expect sign opts: %v, got: %v", expected, opts)
	}
}

func TestSignCommand_CorrectConfig(t *testing.T) {
	opts := &signOpts{}
	command := signCommand(opts)
	expected := &signOpts{
		reference: "ref",
		SignerFlagOpts: cmd.SignerFlagOpts{
			Key:          "key",
			EnvelopeType: envelope.JWS,
		},
		expiry:       365 * 24 * time.Hour,
		pluginConfig: []string{"key0=val0", "key1=val1"},
	}
	if err := command.ParseFlags([]string{
		expected.reference,
		"--key", expected.Key,
		"--signature-format", expected.SignerFlagOpts.EnvelopeType,
		"--expiry", expected.expiry.String(),
		"--plugin-config", "key0=val0",
		"--plugin-config", "key1=val1"}); err != nil {
		t.Fatalf("Parse Flag failed: %v", err)
	}
	if err := command.Args(command, command.Flags().Args()); err != nil {
		t.Fatalf("Parse args failed: %v", err)
	}
	if !reflect.DeepEqual(*expected, *opts) {
		t.Fatalf("Expect sign opts: %v, got: %v", expected, opts)
	}
	config, err := cmd.ParseFlagPluginConfig(opts.pluginConfig)
	if err != nil {
		t.Fatalf("Parse plugin Config flag failed: %v", err)
	}
	if len(config) != 2 {
		t.Fatalf("Expect plugin config number: %v, got: %v ", 2, len(config))
	}
	for i := 0; i < 2; i++ {
		key, val := fmt.Sprintf("key%v", i), fmt.Sprintf("val%v", i)
		configVal, ok := config[key]
		if !ok {
			t.Fatalf("Key: %v not in config", key)
		}
		if val != configVal {
			t.Fatalf("Value for key: %v error, got: %v, expect: %v", key, configVal, val)
		}
	}
}

func TestSignCommand_MissingArgs(t *testing.T) {
	cmd := signCommand(nil)
	if err := cmd.ParseFlags(nil); err != nil {
		t.Fatalf("Parse Flag failed: %v", err)
	}
	if err := cmd.Args(cmd, cmd.Flags().Args()); err == nil {
		t.Fatal("Parse Args expected error, but ok")
	}
}
