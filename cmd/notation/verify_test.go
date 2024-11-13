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
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/notaryproject/notation-go/dir"
	"github.com/notaryproject/notation-go/verifier/trustpolicy"
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
		pluginConfig:         []string{"key1=val1"},
		maxSignatureAttempts: 100,
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
			InsecureRegistry: true,
		},
		pluginConfig:         []string{"key1=val1", "key2=val2"},
		maxSignatureAttempts: 100,
	}
	if err := command.ParseFlags([]string{
		expected.reference,
		"--insecure-registry",
		"--plugin-config", "key1=val1",
		"--plugin-config", "key2=val2"}); err != nil {
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

func TestGetVerifier(t *testing.T) {
	defer func(oldConfiDir, oldCacheDir string) {
		dir.UserConfigDir = oldConfiDir
		dir.UserCacheDir = oldCacheDir
	}(dir.UserConfigDir, dir.UserCacheDir)

	t.Run("success", func(t *testing.T) {
		tempRoot := t.TempDir()
		dir.UserConfigDir = tempRoot
		path := filepath.Join(tempRoot, "trustpolicy.json")
		policyJson, _ := json.Marshal(dummyOCIPolicyDocument())
		if err := os.WriteFile(path, policyJson, 0600); err != nil {
			t.Fatalf("TestLoadOCIDocument write policy file failed. Error: %v", err)
		}
		t.Cleanup(func() { os.RemoveAll(tempRoot) })

		_, err := getVerifier(context.Background())
		if err != nil {
			t.Fatal(err)
		}
	})

	t.Run("non-existing trust policy", func(t *testing.T) {
		dir.UserConfigDir = "/"
		expectedErrMsg := "trust policy is not present. To create a trust policy, see: https://notaryproject.dev/docs/quickstart/#create-a-trust-policy"
		_, err := getVerifier(context.Background())
		if err == nil || err.Error() != expectedErrMsg {
			t.Fatalf("expected %s, but got %s", expectedErrMsg, err)
		}
	})
}

func dummyOCIPolicyDocument() trustpolicy.OCIDocument {
	return trustpolicy.OCIDocument{
		Version: "1.0",
		TrustPolicies: []trustpolicy.OCITrustPolicy{
			{
				Name:                  "test-statement-name",
				RegistryScopes:        []string{"registry.acme-rockets.io/software/net-monitor"},
				SignatureVerification: trustpolicy.SignatureVerification{VerificationLevel: "strict"},
				TrustStores:           []string{"ca:valid-trust-store", "signingAuthority:valid-trust-store"},
				TrustedIdentities:     []string{"x509.subject:CN=Notation Test Root,O=Notary,L=Seattle,ST=WA,C=US"},
			},
		},
	}
}
