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

package blob

import (
	"context"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/notaryproject/notation-core-go/testhelper"
	"github.com/notaryproject/notation-go/dir"
	"github.com/notaryproject/notation/v2/cmd/notation/internal/flag"
	"github.com/notaryproject/notation/v2/internal/envelope"
)

func TestBlobSignCommand_BasicArgs(t *testing.T) {
	opts := &blobSignOpts{}
	command := signCommand(opts)
	expected := &blobSignOpts{
		blobPath: "path",
		SignerFlagOpts: flag.SignerFlagOpts{
			Key:             "key",
			SignatureFormat: envelope.JWS,
		},
		blobMediaType: "application/octet-stream",
	}
	if err := command.ParseFlags([]string{
		expected.blobPath,
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
	opts := &blobSignOpts{}
	command := signCommand(opts)
	expected := &blobSignOpts{
		blobPath: "path",
		SignerFlagOpts: flag.SignerFlagOpts{
			Key:             "key",
			SignatureFormat: envelope.COSE,
		},
		expiry:             24 * time.Hour,
		signatureDirectory: ".",
		blobMediaType:      "application/octet-stream",
	}
	if err := command.ParseFlags([]string{
		expected.blobPath,
		"--key", expected.Key,
		"--signature-format", expected.SignerFlagOpts.SignatureFormat,
		"--expiry", expected.expiry.String(),
		"--signature-directory", "."}); err != nil {
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
	opts := &blobSignOpts{}
	command := signCommand(opts)
	expected := &blobSignOpts{
		blobPath: "path",
		SignerFlagOpts: flag.SignerFlagOpts{
			Key:             "key",
			SignatureFormat: envelope.COSE,
		},
		expiry:        365 * 24 * time.Hour,
		pluginConfig:  []string{"key0=val0", "key1=val1"},
		blobMediaType: "application/octet-stream",
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
	config, err := flag.ParseFlagMap(opts.pluginConfig, flag.PflagPluginConfig.Name)
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
	opts := &blobSignOpts{}
	command := signCommand(opts)
	expected := &blobSignOpts{
		blobPath: "path",
		SignerFlagOpts: flag.SignerFlagOpts{
			KeyID:           "keyID",
			PluginName:      "pluginName",
			SignatureFormat: envelope.JWS,
		},
		blobMediaType: "application/octet-stream",
	}
	if err := command.ParseFlags([]string{
		expected.blobPath,
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
		opts := &blobSignOpts{}
		command := signCommand(opts)
		expected := &blobSignOpts{
			blobPath: "path",
			SignerFlagOpts: flag.SignerFlagOpts{
				KeyID:           "keyID",
				PluginName:      "pluginName",
				Key:             "keyName",
				SignatureFormat: envelope.JWS,
			},
			blobMediaType: "application/octet-stream",
		}
		if err := command.ParseFlags([]string{
			expected.blobPath,
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
		opts := &blobSignOpts{}
		command := signCommand(opts)
		expected := &blobSignOpts{
			blobPath: "path",
			SignerFlagOpts: flag.SignerFlagOpts{
				KeyID:           "keyID",
				Key:             "keyName",
				SignatureFormat: envelope.JWS,
			},
			blobMediaType: "application/octet-stream",
		}
		if err := command.ParseFlags([]string{
			expected.blobPath,
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
		opts := &blobSignOpts{}
		command := signCommand(opts)
		expected := &blobSignOpts{
			blobPath: "path",
			SignerFlagOpts: flag.SignerFlagOpts{
				PluginName:      "pluginName",
				Key:             "keyName",
				SignatureFormat: envelope.JWS,
			},
			blobMediaType: "application/octet-stream",
		}
		if err := command.ParseFlags([]string{
			expected.blobPath,
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
		opts := &blobSignOpts{}
		command := signCommand(opts)
		expected := &blobSignOpts{
			blobPath: "path",
			SignerFlagOpts: flag.SignerFlagOpts{
				KeyID:           "keyID",
				SignatureFormat: envelope.JWS,
			},
			blobMediaType: "application/octet-stream",
		}
		if err := command.ParseFlags([]string{
			expected.blobPath,
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
		opts := &blobSignOpts{}
		command := signCommand(opts)
		expected := &blobSignOpts{
			blobPath: "path",
			SignerFlagOpts: flag.SignerFlagOpts{
				PluginName:      "pluginName",
				SignatureFormat: envelope.JWS,
			},
			blobMediaType: "application/octet-stream",
		}
		if err := command.ParseFlags([]string{
			expected.blobPath,
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
	cmd := signCommand(nil)
	if err := cmd.ParseFlags(nil); err != nil {
		t.Fatalf("Parse Flag failed: %v", err)
	}
	if err := cmd.Args(cmd, cmd.Flags().Args()); err == nil {
		t.Fatal("Parse Args expected error, but ok")
	}
}

func TestBlobSignCommand_ForceOutput(t *testing.T) {
	defer func(oldConfigDir string) {
		dir.UserConfigDir = oldConfigDir
	}(dir.UserConfigDir)
	dir.UserConfigDir = setUpSigningKey(t)

	blobDir := t.TempDir()
	blobPath := filepath.Join(blobDir, "blob.txt")
	if err := os.WriteFile(blobPath, []byte("test blob"), 0600); err != nil {
		t.Fatal(err)
	}
	signaturePath := filepath.Join(blobDir, "blob.txt.jws.sig")
	expectedStdout := fmt.Sprintf("Successfully signed %s\nSignature file written to %s\n", blobPath, signaturePath)
	const warning = "Warning: existing signature file will be overwritten"

	// no signature file exists yet
	stdout, stderr := runBlobSignCommand(t, "--force", blobPath)
	if stdout != expectedStdout {
		t.Fatalf("expected stdout %q, but got %q", expectedStdout, stdout)
	}
	if strings.Contains(stderr, warning) {
		t.Fatalf("expected no overwrite warning when the signature file does not exist, but got stderr %q", stderr)
	}

	// signature file exists
	stdout, stderr = runBlobSignCommand(t, "--force", blobPath)
	if stdout != expectedStdout {
		t.Fatalf("expected stdout %q, but got %q", expectedStdout, stdout)
	}
	if !strings.Contains(stderr, warning) {
		t.Fatalf("expected overwrite warning when the signature file exists, but got stderr %q", stderr)
	}
}

// setUpSigningKey creates a config directory with a default signing key and
// returns its path.
func setUpSigningKey(t *testing.T) string {
	t.Helper()
	configDir := t.TempDir()
	certTuple := testhelper.GetRSASelfSignedSigningCertTuple("test")
	keyBytes, err := x509.MarshalPKCS8PrivateKey(certTuple.PrivateKey)
	if err != nil {
		t.Fatal(err)
	}
	keyPath := filepath.Join(configDir, "test.key")
	if err := os.WriteFile(keyPath, pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: keyBytes}), 0600); err != nil {
		t.Fatal(err)
	}
	certPath := filepath.Join(configDir, "test.crt")
	if err := os.WriteFile(certPath, pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certTuple.Cert.Raw}), 0600); err != nil {
		t.Fatal(err)
	}
	signingKeys, err := json.Marshal(map[string]any{
		"default": "test",
		"keys": []map[string]string{
			{"name": "test", "keyPath": keyPath, "certPath": certPath},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(configDir, dir.PathSigningKeys), signingKeys, 0600); err != nil {
		t.Fatal(err)
	}
	return configDir
}

// runBlobSignCommand runs the blob sign command with args and returns its
// stdout and stderr.
func runBlobSignCommand(t *testing.T, args ...string) (string, string) {
	t.Helper()
	outDir := t.TempDir()
	stdoutFile, err := os.Create(filepath.Join(outDir, "stdout"))
	if err != nil {
		t.Fatal(err)
	}
	defer stdoutFile.Close()
	stderrFile, err := os.Create(filepath.Join(outDir, "stderr"))
	if err != nil {
		t.Fatal(err)
	}
	defer stderrFile.Close()

	oldStdout, oldStderr := os.Stdout, os.Stderr
	os.Stdout, os.Stderr = stdoutFile, stderrFile
	command := signCommand(nil)
	command.SetArgs(args)
	err = command.ExecuteContext(context.Background())
	os.Stdout, os.Stderr = oldStdout, oldStderr
	if err != nil {
		t.Fatalf("blob sign failed: %v", err)
	}

	stdout, err := os.ReadFile(stdoutFile.Name())
	if err != nil {
		t.Fatal(err)
	}
	stderr, err := os.ReadFile(stderrFile.Name())
	if err != nil {
		t.Fatal(err)
	}
	return string(stdout), string(stderr)
}
