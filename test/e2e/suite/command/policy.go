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

package command

import (
	"os"
	"path/filepath"
	"strings"

	. "github.com/notaryproject/notation/test/e2e/internal/notation"
	"github.com/notaryproject/notation/test/e2e/internal/utils"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("trust policy maintainer", func() {
	When("showing configuration", func() {
		It("should show error and hint if policy doesn't exist", func() {
			Host(Opts(), func(notation *utils.ExecOpts, artifact *Artifact, vhost *utils.VirtualHost) {
				notation.ExpectFailure().
					Exec("policy", "show").
					MatchErrKeyWords("failed to show OCI trust policy", "notation policy import")
			})
		})

		It("should show error and hint if policy without read permission", func() {
			Host(Opts(AddTrustPolicyOption(TrustPolicyName, false)), func(notation *utils.ExecOpts, artifact *Artifact, vhost *utils.VirtualHost) {
				trustPolicyPath := vhost.AbsolutePath(NotationDirName, TrustPolicyName)
				os.Chmod(trustPolicyPath, 0200)
				notation.ExpectFailure().
					Exec("policy", "show").
					MatchErrKeyWords("failed to show OCI trust policy", "permission denied")
			})
		})

		It("should show exist policy", func() {
			content, err := os.ReadFile(filepath.Join(NotationE2ETrustPolicyDir, TrustPolicyName))
			Expect(err).NotTo(HaveOccurred())
			Host(Opts(AddTrustPolicyOption(TrustPolicyName, false)), func(notation *utils.ExecOpts, artifact *Artifact, vhost *utils.VirtualHost) {
				notation.Exec("policy", "show").
					MatchContent(string(content))
			})
		})

		It("should display error hint when showing invalid policy", func() {
			policyName := "invalid_format_trustpolicy.json"
			content, err := os.ReadFile(filepath.Join(NotationE2ETrustPolicyDir, policyName))
			Expect(err).NotTo(HaveOccurred())
			Host(Opts(AddTrustPolicyOption(policyName, false)), func(notation *utils.ExecOpts, artifact *Artifact, vhost *utils.VirtualHost) {
				notation.ExpectFailure().Exec("policy", "show").
					MatchErrKeyWords("Existing OCI trust policy configuration is invalid").
					MatchContent(string(content))
			})
		})
	})

	When("importing configuration without existing trust policy configuration", func() {
		opts := Opts()
		It("should fail if no file path is provided", func() {
			Host(opts, func(notation *utils.ExecOpts, artifact *Artifact, vhost *utils.VirtualHost) {
				notation.ExpectFailure().
					Exec("policy", "import").
					MatchErrKeyWords("requires 1 argument but received 0")

			})
		})

		It("should fail if more than one file path is provided", func() {
			Host(opts, func(notation *utils.ExecOpts, artifact *Artifact, vhost *utils.VirtualHost) {
				notation.ExpectFailure().
					Exec("policy", "import", "a", "b").
					MatchErrKeyWords("requires 1 argument but received 2")
			})
		})

		It("should fail if provided file doesn't exist", func() {
			Host(opts, func(notation *utils.ExecOpts, artifact *Artifact, vhost *utils.VirtualHost) {
				notation.ExpectFailure().
					Exec("policy", "import", "/??/???")
			})
		})

		It("should fail if registry scope is malformed", func() {
			Host(opts, func(notation *utils.ExecOpts, artifact *Artifact, vhost *utils.VirtualHost) {
				notation.ExpectFailure().
					Exec("policy", "import", filepath.Join(NotationE2ETrustPolicyDir, "malformed_registry_scope_trustpolicy.json"))
			})
		})

		It("should fail if store is malformed", func() {
			Host(opts, func(notation *utils.ExecOpts, artifact *Artifact, vhost *utils.VirtualHost) {
				notation.ExpectFailure().
					Exec("policy", "import", filepath.Join(NotationE2ETrustPolicyDir, "malformed_trust_store_trustpolicy.json"))
			})
		})

		It("should fail if identity is malformed", func() {
			Host(opts, func(notation *utils.ExecOpts, artifact *Artifact, vhost *utils.VirtualHost) {
				notation.ExpectFailure().
					Exec("policy", "import", filepath.Join(NotationE2ETrustPolicyDir, "malformed_trusted_identity_trustpolicy.json"))
			})
		})

		It("should import successfully", func() {
			Host(opts, func(notation *utils.ExecOpts, artifact *Artifact, vhost *utils.VirtualHost) {
				notation.Exec("policy", "import", filepath.Join(NotationE2ETrustPolicyDir, TrustPolicyName))
			})
		})

		It("should import successfully by force", func() {
			Host(opts, func(notation *utils.ExecOpts, artifact *Artifact, vhost *utils.VirtualHost) {
				notation.Exec("policy", "import", filepath.Join(NotationE2ETrustPolicyDir, TrustPolicyName), "--force")
			})
		})

		It("should failed if trust policy configuration malformed", func() {
			Host(opts, func(notation *utils.ExecOpts, artifact *Artifact, vhost *utils.VirtualHost) {
				notation.ExpectFailure().
					Exec("policy", "import", filepath.Join(NotationE2ETrustPolicyDir, "invalid_format_trustpolicy.json")).
					MatchErrKeyWords("failed to parse OCI trust policy configuration")
			})
		})

		It("should failed if cannot write the policy file", func() {
			Host(opts, func(notation *utils.ExecOpts, artifact *Artifact, vhost *utils.VirtualHost) {
				if err := os.Chmod(vhost.AbsolutePath(NotationDirName), 0400); err != nil {
					Fail(err.Error())
				}
				defer os.Chmod(vhost.AbsolutePath(NotationDirName), 0755)

				notation.ExpectFailure().
					Exec("policy", "import", filepath.Join(NotationE2ETrustPolicyDir, TrustPolicyName)).
					MatchErrKeyWords("failed to write OCI trust policy configuration")
			})
		})
	})

	When("importing configuration with existing trust policy configuration", func() {
		opts := Opts(AddTrustPolicyOption(TrustPolicyName, false))
		It("should fail if no file path is provided", func() {
			Host(opts, func(notation *utils.ExecOpts, artifact *Artifact, vhost *utils.VirtualHost) {
				notation.ExpectFailure().
					Exec("policy", "import")
			})
		})

		It("should fail if provided file doesn't exist", func() {
			Host(opts, func(notation *utils.ExecOpts, artifact *Artifact, vhost *utils.VirtualHost) {
				notation.ExpectFailure().
					Exec("policy", "import", "/??/???", "--force")
			})
		})

		It("should fail if registry scope is malformed", func() {
			Host(opts, func(notation *utils.ExecOpts, artifact *Artifact, vhost *utils.VirtualHost) {
				notation.WithInput(strings.NewReader("Y\n")).ExpectFailure().
					Exec("policy", "import", filepath.Join(NotationE2ETrustPolicyDir, "malformed_registry_scope_trustpolicy.json"))
			})
		})

		It("should fail if store is malformed", func() {
			Host(opts, func(notation *utils.ExecOpts, artifact *Artifact, vhost *utils.VirtualHost) {
				notation.WithInput(strings.NewReader("Y\n")).ExpectFailure().
					Exec("policy", "import", filepath.Join(NotationE2ETrustPolicyDir, "malformed_trust_store_trustpolicy.json"))
			})
		})

		It("should fail if identity is malformed", func() {
			Host(opts, func(notation *utils.ExecOpts, artifact *Artifact, vhost *utils.VirtualHost) {
				notation.WithInput(strings.NewReader("Y\n")).ExpectFailure().
					Exec("policy", "import", filepath.Join(NotationE2ETrustPolicyDir, "malformed_trusted_identity_trustpolicy.json"))
			})
		})

		It("should cancel import with N", func() {
			Host(opts, func(notation *utils.ExecOpts, artifact *Artifact, vhost *utils.VirtualHost) {
				notation.WithInput(strings.NewReader("N\n")).Exec("policy", "import", filepath.Join(NotationE2ETrustPolicyDir, "skip_trustpolicy.json"))
				// validate
				content, err := os.ReadFile(filepath.Join(NotationE2ETrustPolicyDir, TrustPolicyName))
				Expect(err).NotTo(HaveOccurred())
				notation.Exec("policy", "show").MatchContent(string(content))
			})
		})

		It("should cancel import by default", func() {
			Host(opts, func(notation *utils.ExecOpts, artifact *Artifact, vhost *utils.VirtualHost) {
				notation.Exec("policy", "import", filepath.Join(NotationE2ETrustPolicyDir, "skip_trustpolicy.json"))
				// validate
				content, err := os.ReadFile(filepath.Join(NotationE2ETrustPolicyDir, TrustPolicyName))
				Expect(err).NotTo(HaveOccurred())
				notation.Exec("policy", "show").MatchContent(string(content))
			})
		})

		It("should skip confirmation if existing policy is malformed", func() {
			Host(Opts(AddTrustPolicyOption("invalid_format_trustpolicy.json", false)), func(notation *utils.ExecOpts, artifact *Artifact, vhost *utils.VirtualHost) {
				policyFileName := "skip_trustpolicy.json"
				notation.Exec("policy", "import", filepath.Join(NotationE2ETrustPolicyDir, policyFileName)).MatchKeyWords().
					MatchKeyWords("Successfully imported OCI trust policy configuration to")
				// validate
				content, err := os.ReadFile(filepath.Join(NotationE2ETrustPolicyDir, policyFileName))
				Expect(err).NotTo(HaveOccurred())
				notation.Exec("policy", "show").MatchContent(string(content))
			})
		})

		It("should confirm import", func() {
			Host(opts, func(notation *utils.ExecOpts, artifact *Artifact, vhost *utils.VirtualHost) {
				policyFileName := "skip_trustpolicy.json"
				notation.WithInput(strings.NewReader("Y\n")).Exec("policy", "import", filepath.Join(NotationE2ETrustPolicyDir, policyFileName)).
					MatchKeyWords("Successfully imported OCI trust policy configuration to")
				// validate
				content, err := os.ReadFile(filepath.Join(NotationE2ETrustPolicyDir, policyFileName))
				Expect(err).NotTo(HaveOccurred())
				notation.Exec("policy", "show").MatchContent(string(content))
			})
		})

		It("should confirm import by force", func() {
			Host(opts, func(notation *utils.ExecOpts, artifact *Artifact, vhost *utils.VirtualHost) {
				policyFileName := "skip_trustpolicy.json"
				notation.Exec("policy", "import", filepath.Join(NotationE2ETrustPolicyDir, policyFileName), "--force").
					MatchKeyWords("Successfully imported OCI trust policy configuration to")
				// validate
				content, err := os.ReadFile(filepath.Join(NotationE2ETrustPolicyDir, policyFileName))
				Expect(err).NotTo(HaveOccurred())
				notation.Exec("policy", "show").MatchContent(string(content))
			})
		})

		It("should warn when failed to delete old trust policy", func() {
			Host(Opts(), func(notation *utils.ExecOpts, artifact *Artifact, vhost *utils.VirtualHost) {
				// fake a dirctory that named as trustpolicy.json
				fakePolicyPath := vhost.AbsolutePath(NotationDirName, TrustPolicyName)
				if err := os.MkdirAll(fakePolicyPath, 0755); err != nil {
					Fail(err.Error())
				}
				// write a file to create non-empty directory
				if err := os.WriteFile(filepath.Join(fakePolicyPath, "placeholder"), []byte("fake"), 0644); err != nil {
					Fail(err.Error())
				}

				policyFileName := "trustpolicy.json"
				notation.Exec("policy", "import", filepath.Join(NotationE2ETrustPolicyDir, policyFileName), "--force").
					MatchKeyWords(
						"Successfully imported OCI trust policy configuration to",
					).
					MatchErrKeyWords(
						"Warning: existing OCI trust policy configuration will be overwritten",
						"Warning: failed to clean old trust policy configuration",
					)

			})
		})
	})
})
