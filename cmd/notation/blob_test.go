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

	"github.com/notaryproject/notation/internal/cmd"
	"github.com/notaryproject/notation/internal/envelope"
)

func TestBlobSignCommand_BasicArgs(t *testing.T) {
	opts := &blobOpts{}
	command := blobSignCommand(opts)
	expected := &blobOpts{
		blobPath: "path",
		SecureFlagOpts: SecureFlagOpts{
			Username: "user",
			Password: "password",
		},
		SignerFlagOpts: cmd.SignerFlagOpts{
			Key:             "key",
			SignatureFormat: envelope.JWS,
		},
	}
	if err := command.ParseFlags([]string{
		expected.blobPath,
		"-u", expected.Username,
		"--password", expected.Password,
		"--key", expected.Key}); err != nil {
		t.Fatalf("Parse Flag failed: %v", err)
	}
	if err := command.Args(command, command.Flags().Args()); err != nil {
		t.Fatalf("Parse args failed: %v", err)
	}
	if !reflect.DeepEqual(*expected, *opts) {
		t.Fatalf("Expect blob sign opts: %v, got: %v", expected, opts)
	}
}

func TestBlobSignCommand_MoreArgs(t *testing.T) {
	opts := &blobOpts{}
	command := blobSignCommand(opts)
	expected := &blobOpts{
		blobPath: "path",
		SecureFlagOpts: SecureFlagOpts{
			Username:         "user",
			Password:         "password",
			InsecureRegistry: true,
		},
		SignerFlagOpts: cmd.SignerFlagOpts{
			Key:             "key",
			SignatureFormat: envelope.COSE,
		},
		expiry: 24 * time.Hour,
	}
	if err := command.ParseFlags([]string{
		expected.blobPath,
		"-u", expected.Username,
		"-p", expected.Password,
		"--key", expected.Key,
		"--insecure-registry",
		"--signature-format", expected.SignerFlagOpts.SignatureFormat,
		"--expiry", expected.expiry.String()}); err != nil {
		t.Fatalf("Parse Flag failed: %v", err)
	}
	if err := command.Args(command, command.Flags().Args()); err != nil {
		t.Fatalf("Parse args failed: %v", err)
	}
	if !reflect.DeepEqual(*expected, *opts) {
		t.Fatalf("Expect blob sign opts: %v, got: %v", expected, opts)
	}
}

func TestBlobSignCommand_CorrectConfig(t *testing.T) {
	opts := &blobOpts{}
	command := blobSignCommand(opts)
	expected := &blobOpts{
		blobPath: "path",
		SignerFlagOpts: cmd.SignerFlagOpts{
			Key:             "key",
			SignatureFormat: envelope.COSE,
		},
		expiry:       365 * 24 * time.Hour,
		pluginConfig: []string{"key0=val0", "key1=val1"},
	}
	if err := command.ParseFlags([]string{
		expected.blobPath,
		"--key", expected.Key,
		"--signature-format", expected.SignerFlagOpts.SignatureFormat,
		"--expiry", expected.expiry.String(),
		"--plugin-config", "key0=val0",
		"--plugin-config", "key1=val1"}); err != nil {
		t.Fatalf("Parse Flag failed: %v", err)
	}
	if err := command.Args(command, command.Flags().Args()); err != nil {
		t.Fatalf("Parse args failed: %v", err)
	}
	if !reflect.DeepEqual(*expected, *opts) {
		t.Fatalf("Expect sign blob opts: %v, got: %v", expected, opts)
	}
	config, err := cmd.ParseFlagMap(opts.pluginConfig, cmd.PflagPluginConfig.Name)
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

func TestBlobSignCommand_OnDemandKeyOptions(t *testing.T) {
	opts := &blobOpts{}
	command := blobSignCommand(opts)
	expected := &blobOpts{
		blobPath: "path",
		SecureFlagOpts: SecureFlagOpts{
			Username: "user",
			Password: "password",
		},
		SignerFlagOpts: cmd.SignerFlagOpts{
			KeyID:           "keyID",
			PluginName:      "pluginName",
			SignatureFormat: envelope.JWS,
		},
	}
	if err := command.ParseFlags([]string{
		expected.blobPath,
		"-u", expected.Username,
		"--password", expected.Password,
		"--id", expected.KeyID,
		"--plugin", expected.PluginName}); err != nil {
		t.Fatalf("Parse Flag failed: %v", err)
	}
	if err := command.Args(command, command.Flags().Args()); err != nil {
		t.Fatalf("Parse args failed: %v", err)
	}
	if !reflect.DeepEqual(*expected, *opts) {
		t.Fatalf("Expect blob sign opts: %v, got: %v", expected, opts)
	}
}

func TestBlobSignCommand_OnDemandKeyBadOptions(t *testing.T) {
	t.Run("error when using id and plugin options with key", func(t *testing.T) {
		opts := &blobOpts{}
		command := blobSignCommand(opts)
		expected := &blobOpts{
			blobPath: "path",
			SecureFlagOpts: SecureFlagOpts{
				Username: "user",
				Password: "password",
			},
			SignerFlagOpts: cmd.SignerFlagOpts{
				KeyID:           "keyID",
				PluginName:      "pluginName",
				Key:             "keyName",
				SignatureFormat: envelope.JWS,
			},
		}
		if err := command.ParseFlags([]string{
			expected.blobPath,
			"-u", expected.Username,
			"--password", expected.Password,
			"--id", expected.KeyID,
			"--plugin", expected.PluginName,
			"--key", expected.Key}); err != nil {
			t.Fatalf("Parse Flag failed: %v", err)
		}
		if err := command.Args(command, command.Flags().Args()); err != nil {
			t.Fatalf("Parse args failed: %v", err)
		}
		if !reflect.DeepEqual(*expected, *opts) {
			t.Fatalf("Expect blob sign opts: %v, got: %v", expected, opts)
		}
		err := command.ValidateFlagGroups()
		if err == nil || err.Error() != "if any flags in the group [key id] are set none of the others can be; [id key] were all set" {
			t.Fatalf("Didn't get the expected error, but got: %v", err)
		}
	})
	t.Run("error when using key and id options", func(t *testing.T) {
		opts := &blobOpts{}
		command := blobSignCommand(opts)
		expected := &blobOpts{
			blobPath: "path",
			SecureFlagOpts: SecureFlagOpts{
				Username: "user",
				Password: "password",
			},
			SignerFlagOpts: cmd.SignerFlagOpts{
				KeyID:           "keyID",
				Key:             "keyName",
				SignatureFormat: envelope.JWS,
			},
		}
		if err := command.ParseFlags([]string{
			expected.blobPath,
			"-u", expected.Username,
			"--password", expected.Password,
			"--id", expected.KeyID,
			"--key", expected.Key}); err != nil {
			t.Fatalf("Parse Flag failed: %v", err)
		}
		if err := command.Args(command, command.Flags().Args()); err != nil {
			t.Fatalf("Parse args failed: %v", err)
		}
		if !reflect.DeepEqual(*expected, *opts) {
			t.Fatalf("Expect blob sign opts: %v, got: %v", expected, opts)
		}
		err := command.ValidateFlagGroups()
		if err == nil || err.Error() != "if any flags in the group [id plugin] are set they must all be set; missing [plugin]" {
			t.Fatalf("Didn't get the expected error, but got: %v", err)
		}
	})
	t.Run("error when using key and plugin options", func(t *testing.T) {
		opts := &blobOpts{}
		command := blobSignCommand(opts)
		expected := &blobOpts{
			blobPath: "path",
			SecureFlagOpts: SecureFlagOpts{
				Username: "user",
				Password: "password",
			},
			SignerFlagOpts: cmd.SignerFlagOpts{
				PluginName:      "pluginName",
				Key:             "keyName",
				SignatureFormat: envelope.JWS,
			},
		}
		if err := command.ParseFlags([]string{
			expected.blobPath,
			"-u", expected.Username,
			"--password", expected.Password,
			"--plugin", expected.PluginName,
			"--key", expected.Key}); err != nil {
			t.Fatalf("Parse Flag failed: %v", err)
		}
		if err := command.Args(command, command.Flags().Args()); err != nil {
			t.Fatalf("Parse args failed: %v", err)
		}
		if !reflect.DeepEqual(*expected, *opts) {
			t.Fatalf("Expect blob sign opts: %v, got: %v", expected, opts)
		}
		err := command.ValidateFlagGroups()
		if err == nil || err.Error() != "if any flags in the group [id plugin] are set they must all be set; missing [id]" {
			t.Fatalf("Didn't get the expected error, but got: %v", err)
		}
	})
	t.Run("error when using id option and not plugin", func(t *testing.T) {
		opts := &blobOpts{}
		command := blobSignCommand(opts)
		expected := &blobOpts{
			blobPath: "path",
			SecureFlagOpts: SecureFlagOpts{
				Username: "user",
				Password: "password",
			},
			SignerFlagOpts: cmd.SignerFlagOpts{
				KeyID:           "keyID",
				SignatureFormat: envelope.JWS,
			},
		}
		if err := command.ParseFlags([]string{
			expected.blobPath,
			"-u", expected.Username,
			"--password", expected.Password,
			"--id", expected.KeyID}); err != nil {
			t.Fatalf("Parse Flag failed: %v", err)
		}
		if err := command.Args(command, command.Flags().Args()); err != nil {
			t.Fatalf("Parse args failed: %v", err)
		}
		if !reflect.DeepEqual(*expected, *opts) {
			t.Fatalf("Expect blob sign opts: %v, got: %v", expected, opts)
		}
		err := command.ValidateFlagGroups()
		if err == nil || err.Error() != "if any flags in the group [id plugin] are set they must all be set; missing [plugin]" {
			t.Fatalf("Didn't get the expected error, but got: %v", err)
		}
	})
	t.Run("error when using plugin option and not id", func(t *testing.T) {
		opts := &blobOpts{}
		command := blobSignCommand(opts)
		expected := &blobOpts{
			blobPath: "path",
			SecureFlagOpts: SecureFlagOpts{
				Username: "user",
				Password: "password",
			},
			SignerFlagOpts: cmd.SignerFlagOpts{
				PluginName:      "pluginName",
				SignatureFormat: envelope.JWS,
			},
		}
		if err := command.ParseFlags([]string{
			expected.blobPath,
			"-u", expected.Username,
			"--password", expected.Password,
			"--plugin", expected.PluginName}); err != nil {
			t.Fatalf("Parse Flag failed: %v", err)
		}
		if err := command.Args(command, command.Flags().Args()); err != nil {
			t.Fatalf("Parse args failed: %v", err)
		}
		if !reflect.DeepEqual(*expected, *opts) {
			t.Fatalf("Expect blob sign opts: %v, got: %v", expected, opts)
		}
		err := command.ValidateFlagGroups()
		if err == nil || err.Error() != "if any flags in the group [id plugin] are set they must all be set; missing [id]" {
			t.Fatalf("Didn't get the expected error, but got: %v", err)
		}
	})
}

func TestBlobSignCommand_MissingArgs(t *testing.T) {
	cmd := blobSignCommand(nil)
	if err := cmd.ParseFlags(nil); err != nil {
		t.Fatalf("Parse Flag failed: %v", err)
	}
	if err := cmd.Args(cmd, cmd.Flags().Args()); err == nil {
		t.Fatal("Parse Args expected error, but ok")
	}
}

func TestBlobInspectCommand_SecretsFromArgs(t *testing.T) {
	opts := &blobOpts{}
	command := blobInspectCommand(opts)
	expected := &blobOpts{
		signaturePath: "path",
		SecureFlagOpts: SecureFlagOpts{
			Password:         "password",
			InsecureRegistry: true,
			Username:         "user",
		},
		outputFormat: cmd.OutputPlaintext}
	if err := command.ParseFlags([]string{
		expected.signaturePath,
		"--password", expected.Password,
		"-u", expected.Username,
		"--insecure-registry",
		"--output", "text"}); err != nil {
		t.Fatalf("Parse Flag failed: %v", err)
	}
	if err := command.Args(command, command.Flags().Args()); err != nil {
		t.Fatalf("Parse Args failed: %v", err)
	}
	if !reflect.DeepEqual(*expected, *opts) {
		t.Fatalf("Expect blob inspect opts: %v, got: %v", expected, opts)
	}
}

func TestBlobInspectCommand_SecretsFromEnv(t *testing.T) {
	t.Setenv(defaultUsernameEnv, "user")
	t.Setenv(defaultPasswordEnv, "password")
	opts := &blobOpts{}
	expected := &blobOpts{
		signaturePath: "path",
		SecureFlagOpts: SecureFlagOpts{
			Password: "password",
			Username: "user",
		},
		outputFormat: cmd.OutputJSON,
	}
	command := blobInspectCommand(opts)
	if err := command.ParseFlags([]string{
		expected.signaturePath,
		"--output", "json"}); err != nil {
		t.Fatalf("Parse Flag failed: %v", err)
	}
	if err := command.Args(command, command.Flags().Args()); err != nil {
		t.Fatalf("Parse Args failed: %v", err)
	}
	if !reflect.DeepEqual(*expected, *opts) {
		t.Fatalf("Expect blob inspect opts: %v, got: %v", expected, opts)
	}
}

func TestBlobInspectCommand_MissingArgs(t *testing.T) {
	command := blobInspectCommand(nil)
	if err := command.ParseFlags(nil); err != nil {
		t.Fatalf("Parse Flag failed: %v", err)
	}
	if err := command.Args(command, command.Flags().Args()); err == nil {
		t.Fatal("Parse Args expected error, but ok")
	}
}
