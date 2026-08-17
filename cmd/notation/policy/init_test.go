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

package policy

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/notaryproject/notation-go/dir"
	"github.com/notaryproject/notation-go/verifier/trustpolicy"
	"github.com/notaryproject/notation/v2/cmd/notation/internal/display/output"
)

func newInitOpts() *initOpts {
	return &initOpts{
		printer:           output.NewPrinter(os.Stdout, os.Stderr),
		name:              "test-policy",
		registryScopes:    []string{wildcardRegistryScope},
		trustStores:       []string{"ca:test-store"},
		trustedIdentities: []string{"x509.subject: C=US, ST=WA, O=acme-rockets.io"},
	}
}

func TestRunInit(t *testing.T) {
	defer func(old string) { dir.UserConfigDir = old }(dir.UserConfigDir)

	t.Run("writes a valid starter policy", func(t *testing.T) {
		tempRoot := t.TempDir()
		dir.UserConfigDir = tempRoot

		if err := runInit(newInitOpts()); err != nil {
			t.Fatalf("runInit failed: %v", err)
		}

		policyPath := filepath.Join(tempRoot, "trustpolicy.oci.json")
		data, err := os.ReadFile(policyPath)
		if err != nil {
			t.Fatalf("expected policy file at %s: %v", policyPath, err)
		}
		var doc trustpolicy.OCIDocument
		if err := json.Unmarshal(data, &doc); err != nil {
			t.Fatalf("generated policy is not valid JSON: %v", err)
		}
		if err := doc.Validate(); err != nil {
			t.Fatalf("generated policy did not validate: %v", err)
		}
		if len(doc.TrustPolicies) != 1 || doc.TrustPolicies[0].Name != "test-policy" {
			t.Fatalf("unexpected policy content: %+v", doc)
		}
		if got := doc.TrustPolicies[0].RegistryScopes; len(got) != 1 || got[0] != wildcardRegistryScope {
			t.Fatalf("expected wildcard registry scope by default, got %v", got)
		}
	})

	t.Run("refuses invalid input", func(t *testing.T) {
		tempRoot := t.TempDir()
		dir.UserConfigDir = tempRoot
		opts := newInitOpts()
		// wildcard scope cannot be combined with another scope
		opts.registryScopes = []string{wildcardRegistryScope, "registry.acme-rockets.io/software/net-monitor"}
		if err := runInit(opts); err == nil {
			t.Fatal("expected validation error for wildcard combined with another scope, got nil")
		}
	})

	t.Run("force overwrites an existing policy", func(t *testing.T) {
		tempRoot := t.TempDir()
		dir.UserConfigDir = tempRoot
		policyPath := filepath.Join(tempRoot, "trustpolicy.oci.json")
		if err := os.WriteFile(policyPath, []byte("existing junk"), 0600); err != nil {
			t.Fatalf("seeding existing policy failed: %v", err)
		}

		opts := newInitOpts()
		opts.force = true
		if err := runInit(opts); err != nil {
			t.Fatalf("runInit with --force failed: %v", err)
		}
		var doc trustpolicy.OCIDocument
		data, err := os.ReadFile(policyPath)
		if err != nil {
			t.Fatalf("reading overwritten policy failed: %v", err)
		}
		if err := json.Unmarshal(data, &doc); err != nil {
			t.Fatalf("overwritten policy is not valid JSON: %v", err)
		}
		if err := doc.Validate(); err != nil {
			t.Fatalf("overwritten policy did not validate: %v", err)
		}
	})
}
