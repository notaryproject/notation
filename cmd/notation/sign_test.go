package main

import (
	"fmt"
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
		RemoteFlagOpts: RemoteFlagOpts{
			SecureFlagOpts: SecureFlagOpts{
				Username: "user",
				Password: "password",
			},
			CommonFlagOpts: CommonFlagOpts{
				MediaType: defaultMediaType,
			},
		},
		SignerFlagOpts: cmd.SignerFlagOpts{
			Key:          "key",
			KeyFile:      "keyfile",
			CertFile:     "certfile",
			EnvelopeType: envelope.JWS,
		},
		push: true,
	}
	if err := command.ParseFlags([]string{
		expected.reference,
		"-u", expected.Username,
		"--password", expected.Password,
		"--key", expected.Key,
		"--key-file", expected.KeyFile,
		"--cert-file", expected.CertFile}); err != nil {
		t.Fatalf("Parse Flag failed: %v", err)
	}
	if err := command.Args(command, command.Flags().Args()); err != nil {
		t.Fatalf("Parse args failed: %v", err)
	}
	if *expected != *opts {
		t.Fatalf("Expect sign opts: %v, got: %v", expected, opts)
	}
}

func TestSignCommand_MoreArgs(t *testing.T) {
	opts := &signOpts{}
	command := signCommand(opts)
	expected := &signOpts{
		reference: "ref",
		RemoteFlagOpts: RemoteFlagOpts{
			SecureFlagOpts: SecureFlagOpts{
				Username:  "user",
				Password:  "password",
				PlainHTTP: true,
			},
			CommonFlagOpts: CommonFlagOpts{
				MediaType: "mediaT",
				Local:     true,
			},
		},
		SignerFlagOpts: cmd.SignerFlagOpts{
			Key:          "key",
			KeyFile:      "keyfile",
			CertFile:     "certfile",
			EnvelopeType: envelope.COSE,
		},
		output: "outputfile",
		push:   false,
		expiry: 24 * time.Hour,
	}
	if err := command.ParseFlags([]string{
		expected.reference,
		"-u", expected.Username,
		"-p", expected.Password,
		"--key", expected.Key,
		"--key-file", expected.KeyFile,
		"--cert-file", expected.CertFile,
		"--plain-http",
		"--push=false",
		"--media-type", expected.MediaType,
		"-l",
		"--envelope-type", expected.SignerFlagOpts.EnvelopeType,
		"--output", expected.output,
		"--expiry", expected.expiry.String()}); err != nil {
		t.Fatalf("Parse Flag failed: %v", err)
	}
	if err := command.Args(command, command.Flags().Args()); err != nil {
		t.Fatalf("Parse args failed: %v", err)
	}
	if *expected != *opts {
		t.Fatalf("Expect sign opts: %v, got: %v", expected, opts)
	}
}

func TestSignCommand_CorrectConfig(t *testing.T) {
	opts := &signOpts{}
	command := signCommand(opts)
	expected := &signOpts{
		reference: "ref",
		RemoteFlagOpts: RemoteFlagOpts{
			CommonFlagOpts: CommonFlagOpts{
				MediaType: "mediaT",
				Local:     true,
			},
		},
		SignerFlagOpts: cmd.SignerFlagOpts{
			Key:          "key",
			KeyFile:      "keyfile",
			CertFile:     "certfile",
			EnvelopeType: envelope.JWS,
		},
		push:            true,
		expiry:          365 * 24 * time.Hour,
		pluginConfig:    "key0=val0,key1=val1,key2=val2",
		originReference: "originref",
		pushReference:   "pushref",
	}
	if err := command.ParseFlags([]string{
		expected.reference,
		"--key", expected.Key,
		"--key-file", expected.KeyFile,
		"--cert-file", expected.CertFile,
		"--push",
		"--media-type", expected.MediaType,
		"--push-reference", expected.pushReference,
		"-r", expected.originReference,
		"--local",
		"--envelope-type", expected.SignerFlagOpts.EnvelopeType,
		"--expiry", expected.expiry.String(),
		"--pluginConfig", expected.pluginConfig}); err != nil {
		t.Fatalf("Parse Flag failed: %v", err)
	}
	if err := command.Args(command, command.Flags().Args()); err != nil {
		t.Fatalf("Parse args failed: %v", err)
	}
	if *expected != *opts {
		t.Fatalf("Expect sign opts: %v, got: %v", expected, opts)
	}
	config, err := cmd.ParseFlagPluginConfig(opts.pluginConfig)
	if err != nil {
		t.Fatalf("Parse plugin Config flag failed: %v", err)
	}
	if len(config) != 3 {
		t.Fatalf("Expect plugin config number: %v, got: %v ", 3, len(config))
	}
	for i := 0; i < 3; i++ {
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
