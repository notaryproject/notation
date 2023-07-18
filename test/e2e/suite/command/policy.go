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
					MatchErrKeyWords("failed to load trust policy configuration", "notation policy import")
			})
		})

		It("should show exist policy", func() {
			content, err := os.ReadFile(filepath.Join(NotationE2ETrustPolicyDir, TrustPolicyName))
			Expect(err).NotTo(HaveOccurred())
			Host(Opts(AddTrustPolicyOption(TrustPolicyName)), func(notation *utils.ExecOpts, artifact *Artifact, vhost *utils.VirtualHost) {
				notation.Exec("policy", "show").
					MatchContent(string(content))
			})
		})

		It("should display error hint when showing invalid policy", func() {
			policyName := "invalid_format_trustpolicy.json"
			content, err := os.ReadFile(filepath.Join(NotationE2ETrustPolicyDir, policyName))
			Expect(err).NotTo(HaveOccurred())
			Host(Opts(AddTrustPolicyOption(policyName)), func(notation *utils.ExecOpts, artifact *Artifact, vhost *utils.VirtualHost) {
				notation.Exec("policy", "show").
					MatchErrKeyWords("existing trust policy configuration is invalid").
					MatchContent(string(content))
			})
		})
	})

	When("importing configuration without existing trust policy configuration", func() {
		opts := Opts()
		It("should fail if no file path is provided", func() {
			Host(opts, func(notation *utils.ExecOpts, artifact *Artifact, vhost *utils.VirtualHost) {
				notation.ExpectFailure().
					Exec("policy", "import")
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
	})

	When("importing configuration with existing trust policy configuration", func() {
		opts := Opts(AddTrustPolicyOption(TrustPolicyName))
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
			Host(Opts(AddTrustPolicyOption("invalid_format_trustpolicy.json")), func(notation *utils.ExecOpts, artifact *Artifact, vhost *utils.VirtualHost) {
				policyFileName := "skip_trustpolicy.json"
				notation.Exec("policy", "import", filepath.Join(NotationE2ETrustPolicyDir, policyFileName)).MatchKeyWords().
					MatchKeyWords("Trust policy configuration imported successfully.")
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
					MatchKeyWords("Trust policy configuration imported successfully.")
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
					MatchKeyWords("Trust policy configuration imported successfully.")
				// validate
				content, err := os.ReadFile(filepath.Join(NotationE2ETrustPolicyDir, policyFileName))
				Expect(err).NotTo(HaveOccurred())
				notation.Exec("policy", "show").MatchContent(string(content))
			})
		})
	})
})
