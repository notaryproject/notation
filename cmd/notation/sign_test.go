// Copyright The Notary Project Authors.
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/notaryproject/notation/cmd/notation/internal/option"
	"github.com/notaryproject/notation/internal/envelope"
)

func TestSignCommand_BasicArgs(t *testing.T) {
	opts := &signOpts{}
	command := signCommand(opts)
	expected := &signOpts{
		reference: "ref",
		Secure: option.Secure{
			Username: "user",
			Password: "password",
		},
		Signer: option.Signer{
			Key:             "key",
			SignatureFormat: envelope.JWS,
		},
		forceReferrersTag: true,
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
		Secure: option.Secure{
			Username:         "user",
			Password:         "password",
			InsecureRegistry: true,
		},
		Signer: option.Signer{
			Key:             "key",
			SignatureFormat: envelope.COSE,
			Expiry:          24 * time.Hour,
		},
		forceReferrersTag: true,
	}
	if err := command.ParseFlags([]string{
		expected.reference,
		"-u", expected.Username,
		"-p", expected.Password,
		"--key", expected.Key,
		"--insecure-registry",
		"--signature-format", expected.Signer.SignatureFormat,
		"--expiry", expected.Expiry.String(),
		"--force-referrers-tag",
	}); err != nil {
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
		Signer: option.Signer{
			Plugin: option.Plugin{
				PluginConfig: []string{"key0=val0", "key1=val1"},
			},
			Key:             "key",
			SignatureFormat: envelope.COSE,
			Expiry:          365 * 24 * time.Hour,
		},
		forceReferrersTag: false,
	}
	if err := command.ParseFlags([]string{
		expected.reference,
		"--key", expected.Key,
		"--signature-format", expected.Signer.SignatureFormat,
		"--expiry", expected.Expiry.String(),
		"--plugin-config", "key0=val0",
		"--plugin-config", "key1=val1",
		"--force-referrers-tag=false",
	}); err != nil {
		t.Fatalf("Parse Flag failed: %v", err)
	}
	if err := command.Args(command, command.Flags().Args()); err != nil {
		t.Fatalf("Parse args failed: %v", err)
	}
	if !reflect.DeepEqual(*expected, *opts) {
		t.Fatalf("Expect sign opts: %v, got: %v", expected, opts)
	}
	config, err := opts.PluginConfig.ToMap()
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

func TestSignCommmand_OnDemandKeyOptions(t *testing.T) {
	opts := &signOpts{}
	command := signCommand(opts)
	expected := &signOpts{
		reference: "ref",
		Secure: option.Secure{
			Username: "user",
			Password: "password",
		},
		Signer: option.Signer{
			Plugin: option.Plugin{
				KeyID:      "keyID",
				PluginName: "pluginName",
			},
			SignatureFormat: envelope.JWS,
		},
	}
	if err := command.ParseFlags([]string{
		expected.reference,
		"-u", expected.Username,
		"--password", expected.Password,
		"--id", expected.KeyID,
		"--plugin", expected.PluginName,
		"--force-referrers-tag=false",
	}); err != nil {
		t.Fatalf("Parse Flag failed: %v", err)
	}
	if err := command.Args(command, command.Flags().Args()); err != nil {
		t.Fatalf("Parse args failed: %v", err)
	}
	if !reflect.DeepEqual(*expected, *opts) {
		t.Fatalf("Expect sign opts: %v, got: %v", expected, opts)
	}
}

func TestSignCommmand_OnDemandKeyBadOptions(t *testing.T) {
	t.Run("error when using id and plugin options with key", func(t *testing.T) {
		opts := &signOpts{}
		command := signCommand(opts)
		expected := &signOpts{
			reference: "ref",
			Secure: option.Secure{
				Username: "user",
				Password: "password",
			},
			Signer: option.Signer{
				Plugin: option.Plugin{
					KeyID:      "keyID",
					PluginName: "pluginName",
				},
				Key:             "keyName",
				SignatureFormat: envelope.JWS,
			},
		}
		if err := command.ParseFlags([]string{
			expected.reference,
			"-u", expected.Username,
			"--password", expected.Password,
			"--id", expected.KeyID,
			"--plugin", expected.PluginName,
			"--key", expected.Key,
			"--force-referrers-tag=false",
		}); err != nil {
			t.Fatalf("Parse Flag failed: %v", err)
		}
		if err := command.Args(command, command.Flags().Args()); err != nil {
			t.Fatalf("Parse args failed: %v", err)
		}
		if !reflect.DeepEqual(*expected, *opts) {
			t.Fatalf("Expect sign opts: %v, got: %v", expected, opts)
		}
		err := command.ValidateFlagGroups()
		if err == nil || err.Error() != "if any flags in the group [key id] are set none of the others can be; [id key] were all set" {
			t.Fatalf("Didn't get the expected error, but got: %v", err)
		}
	})
	t.Run("error when using key and id options", func(t *testing.T) {
		opts := &signOpts{}
		command := signCommand(opts)
		expected := &signOpts{
			reference: "ref",
			Secure: option.Secure{
				Username: "user",
				Password: "password",
			},
			Signer: option.Signer{
				Plugin: option.Plugin{
					KeyID: "keyID",
				},
				Key:             "keyName",
				SignatureFormat: envelope.JWS,
			},
		}
		if err := command.ParseFlags([]string{
			expected.reference,
			"-u", expected.Username,
			"--password", expected.Password,
			"--id", expected.KeyID,
			"--key", expected.Key,
			"--force-referrers-tag=false",
		}); err != nil {
			t.Fatalf("Parse Flag failed: %v", err)
		}
		if err := command.Args(command, command.Flags().Args()); err != nil {
			t.Fatalf("Parse args failed: %v", err)
		}
		if !reflect.DeepEqual(*expected, *opts) {
			t.Fatalf("Expect sign opts: %v, got: %v", expected, opts)
		}
		err := command.ValidateFlagGroups()
		if err == nil || err.Error() != "if any flags in the group [id plugin] are set they must all be set; missing [plugin]" {
			t.Fatalf("Didn't get the expected error, but got: %v", err)
		}
	})
	t.Run("error when using key and plugin options", func(t *testing.T) {
		opts := &signOpts{}
		command := signCommand(opts)
		expected := &signOpts{
			reference: "ref",
			Secure: option.Secure{
				Username: "user",
				Password: "password",
			},
			Signer: option.Signer{
				Plugin: option.Plugin{
					PluginName: "pluginName",
				},
				Key:             "keyName",
				SignatureFormat: envelope.JWS,
			},
		}
		if err := command.ParseFlags([]string{
			expected.reference,
			"-u", expected.Username,
			"--password", expected.Password,
			"--plugin", expected.PluginName,
			"--key", expected.Key,
			"--force-referrers-tag=false",
		}); err != nil {
			t.Fatalf("Parse Flag failed: %v", err)
		}
		if err := command.Args(command, command.Flags().Args()); err != nil {
			t.Fatalf("Parse args failed: %v", err)
		}
		if !reflect.DeepEqual(*expected, *opts) {
			t.Fatalf("Expect sign opts: %v, got: %v", expected, opts)
		}
		err := command.ValidateFlagGroups()
		if err == nil || err.Error() != "if any flags in the group [id plugin] are set they must all be set; missing [id]" {
			t.Fatalf("Didn't get the expected error, but got: %v", err)
		}
	})
	t.Run("error when using id option and not plugin", func(t *testing.T) {
		opts := &signOpts{}
		command := signCommand(opts)
		expected := &signOpts{
			reference: "ref",
			Secure: option.Secure{
				Username: "user",
				Password: "password",
			},
			Signer: option.Signer{
				Plugin: option.Plugin{
					KeyID: "keyID",
				},
				SignatureFormat: envelope.JWS,
			},
		}
		if err := command.ParseFlags([]string{
			expected.reference,
			"-u", expected.Username,
			"--password", expected.Password,
			"--id", expected.KeyID,
			"--force-referrers-tag=false",
		}); err != nil {
			t.Fatalf("Parse Flag failed: %v", err)
		}
		if err := command.Args(command, command.Flags().Args()); err != nil {
			t.Fatalf("Parse args failed: %v", err)
		}
		if !reflect.DeepEqual(*expected, *opts) {
			t.Fatalf("Expect sign opts: %v, got: %v", expected, opts)
		}
		err := command.ValidateFlagGroups()
		if err == nil || err.Error() != "if any flags in the group [id plugin] are set they must all be set; missing [plugin]" {
			t.Fatalf("Didn't get the expected error, but got: %v", err)
		}
	})
	t.Run("error when using plugin option and not id", func(t *testing.T) {
		opts := &signOpts{}
		command := signCommand(opts)
		expected := &signOpts{
			reference: "ref",
			Secure: option.Secure{
				Username: "user",
				Password: "password",
			},
			Signer: option.Signer{
				Plugin: option.Plugin{
					PluginName: "pluginName",
				},
				SignatureFormat: envelope.JWS,
			},
		}
		if err := command.ParseFlags([]string{
			expected.reference,
			"-u", expected.Username,
			"--password", expected.Password,
			"--plugin", expected.PluginName,
			"--force-referrers-tag=false",
		}); err != nil {
			t.Fatalf("Parse Flag failed: %v", err)
		}
		if err := command.Args(command, command.Flags().Args()); err != nil {
			t.Fatalf("Parse args failed: %v", err)
		}
		if !reflect.DeepEqual(*expected, *opts) {
			t.Fatalf("Expect sign opts: %v, got: %v", expected, opts)
		}
		err := command.ValidateFlagGroups()
		if err == nil || err.Error() != "if any flags in the group [id plugin] are set they must all be set; missing [id]" {
			t.Fatalf("Didn't get the expected error, but got: %v", err)
		}
	})
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
