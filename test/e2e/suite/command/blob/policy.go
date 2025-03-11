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
	"os"
	"path/filepath"
	"strings"

	. "github.com/notaryproject/notation/test/e2e/internal/notation"
	"github.com/notaryproject/notation/test/e2e/internal/utils"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

const validBlobTrustPolicyName = "blob_trust_policy.json"

var _ = Describe("blob trust policy maintainer", func() {
	When("showing configuration", func() {
		It("should show error and hint if policy doesn't exist", func() {
			Host(Opts(), func(notation *utils.ExecOpts, artifact *Artifact, vhost *utils.VirtualHost) {
				notation.ExpectFailure().
					Exec("blob", "policy", "show").
					MatchErrKeyWords("failed to show blob trust policy", "notation blob policy import")
			})
		})

		It("should show error and hint if policy without read permission", func() {
			Host(Opts(AddBlobTrustPolicyOption(validBlobTrustPolicyName)), func(notation *utils.ExecOpts, artifact *Artifact, vhost *utils.VirtualHost) {
				trustPolicyPath := vhost.AbsolutePath(NotationDirName, BlobTrustPolicyName)
				os.Chmod(trustPolicyPath, 0200)
				notation.ExpectFailure().
					Exec("blob", "policy", "show").
					MatchErrKeyWords("failed to show trust policy", "permission denied")
			})
		})

		It("should show exist policy", func() {
			content, err := os.ReadFile(filepath.Join(NotationE2ETrustPolicyDir, validBlobTrustPolicyName))
			Expect(err).NotTo(HaveOccurred())
			Host(Opts(AddBlobTrustPolicyOption(validBlobTrustPolicyName)), func(notation *utils.ExecOpts, artifact *Artifact, vhost *utils.VirtualHost) {
				notation.Exec("blob", "policy", "show").
					MatchContent(string(content))
			})
		})

		It("should display error hint when showing invalid policy", func() {
			policyName := "invalid_format_trustpolicy.json"
			content, err := os.ReadFile(filepath.Join(NotationE2ETrustPolicyDir, policyName))
			Expect(err).NotTo(HaveOccurred())
			Host(Opts(AddBlobTrustPolicyOption(policyName)), func(notation *utils.ExecOpts, artifact *Artifact, vhost *utils.VirtualHost) {
				notation.ExpectFailure().Exec("blob", "policy", "show").
					MatchErrKeyWords("existing blob trust policy configuration is invalid").
					MatchContent(string(content))
			})
		})
	})

	When("importing configuration without existing trust policy configuration", func() {
		opts := Opts()
		It("should fail if no file path is provided", func() {
			Host(opts, func(notation *utils.ExecOpts, artifact *Artifact, vhost *utils.VirtualHost) {
				notation.ExpectFailure().
					Exec("blob", "policy", "import").
					MatchErrKeyWords("requires 1 argument but received 0")

			})
		})

		It("should fail if more than one file path is provided", func() {
			Host(opts, func(notation *utils.ExecOpts, artifact *Artifact, vhost *utils.VirtualHost) {
				notation.ExpectFailure().
					Exec("blob", "policy", "import", "a", "b").
					MatchErrKeyWords("requires 1 argument but received 2")
			})
		})

		It("should fail if provided file doesn't exist", func() {
			Host(opts, func(notation *utils.ExecOpts, artifact *Artifact, vhost *utils.VirtualHost) {
				notation.ExpectFailure().
					Exec("blob", "policy", "import", "/??/???")
			})
		})

		It("should fail if identity is malformed", func() {
			Host(opts, func(notation *utils.ExecOpts, artifact *Artifact, vhost *utils.VirtualHost) {
				notation.ExpectFailure().
					Exec("blob", "policy", "import", filepath.Join(NotationE2ETrustPolicyDir, "malformed_trusted_identity_trustpolicy.json"))
			})
		})

		It("should import successfully", func() {
			Host(opts, func(notation *utils.ExecOpts, artifact *Artifact, vhost *utils.VirtualHost) {
				notation.Exec("blob", "policy", "import", filepath.Join(NotationE2ETrustPolicyDir, validBlobTrustPolicyName))
			})
		})

		It("should import successfully by force", func() {
			Host(opts, func(notation *utils.ExecOpts, artifact *Artifact, vhost *utils.VirtualHost) {
				notation.Exec("blob", "policy", "import", filepath.Join(NotationE2ETrustPolicyDir, validBlobTrustPolicyName), "--force")
			})
		})

		It("should failed if without permission to write policy", func() {
			Host(opts, func(notation *utils.ExecOpts, artifact *Artifact, vhost *utils.VirtualHost) {
				notation.
					Exec("blob", "policy", "import", filepath.Join(NotationE2ETrustPolicyDir, validBlobTrustPolicyName))

				trustPolicyPath := vhost.AbsolutePath(NotationDirName)
				os.Chmod(trustPolicyPath, 0000)
				defer os.Chmod(trustPolicyPath, 0755)

				notation.ExpectFailure().
					Exec("blob", "policy", "import", filepath.Join(NotationE2ETrustPolicyDir, validBlobTrustPolicyName), "--force").
					MatchErrKeyWords("failed to write blob trust policy configuration")
			})
		})

		It("should failed if provide file is malformed json", func() {
			Host(opts, func(notation *utils.ExecOpts, artifact *Artifact, vhost *utils.VirtualHost) {
				notation.ExpectFailure().
					Exec("blob", "policy", "import", filepath.Join(NotationE2ETrustPolicyDir, "invalid_format_trustpolicy.json"))
			})
		})
	})

	When("importing configuration with existing trust policy configuration", func() {
		opts := Opts(AddBlobTrustPolicyOption(validBlobTrustPolicyName))
		It("should fail if no file path is provided", func() {
			Host(opts, func(notation *utils.ExecOpts, artifact *Artifact, vhost *utils.VirtualHost) {
				notation.ExpectFailure().
					Exec("blob", "policy", "import").
					MatchErrKeyWords("requires 1 argument but received 0")
			})
		})

		It("should fail if provided file doesn't exist", func() {
			Host(opts, func(notation *utils.ExecOpts, artifact *Artifact, vhost *utils.VirtualHost) {
				notation.ExpectFailure().
					Exec("blob", "policy", "import", "/??/???", "--force")
			})
		})

		It("should fail if store is malformed", func() {
			Host(opts, func(notation *utils.ExecOpts, artifact *Artifact, vhost *utils.VirtualHost) {
				notation.WithInput(strings.NewReader("Y\n")).ExpectFailure().
					Exec("blob", "policy", "import", filepath.Join(NotationE2ETrustPolicyDir, "malformed_trust_store_trustpolicy.json"))
			})
		})

		It("should fail if identity is malformed", func() {
			Host(opts, func(notation *utils.ExecOpts, artifact *Artifact, vhost *utils.VirtualHost) {
				notation.WithInput(strings.NewReader("Y\n")).ExpectFailure().
					Exec("blob", "policy", "import", filepath.Join(NotationE2ETrustPolicyDir, "malformed_trusted_identity_trustpolicy.json"))
			})
		})

		It("should cancel import with N", func() {
			Host(opts, func(notation *utils.ExecOpts, artifact *Artifact, vhost *utils.VirtualHost) {
				notation.WithInput(strings.NewReader("N\n")).Exec("blob", "policy", "import", filepath.Join(NotationE2ETrustPolicyDir, "skip_trustpolicy.json"))
				// validate
				content, err := os.ReadFile(filepath.Join(NotationE2ETrustPolicyDir, validBlobTrustPolicyName))
				Expect(err).NotTo(HaveOccurred())
				notation.Exec("blob", "policy", "show").MatchContent(string(content))
			})
		})

		It("should cancel import by default", func() {
			Host(opts, func(notation *utils.ExecOpts, artifact *Artifact, vhost *utils.VirtualHost) {
				notation.Exec("blob", "policy", "import", filepath.Join(NotationE2ETrustPolicyDir, "skip_trustpolicy.json"))
				// validate
				content, err := os.ReadFile(filepath.Join(NotationE2ETrustPolicyDir, validBlobTrustPolicyName))
				Expect(err).NotTo(HaveOccurred())
				notation.Exec("blob", "policy", "show").MatchContent(string(content))
			})
		})

		It("should skip confirmation if existing policy is malformed", func() {
			Host(Opts(AddBlobTrustPolicyOption("invalid_format_trustpolicy.json")), func(notation *utils.ExecOpts, artifact *Artifact, vhost *utils.VirtualHost) {
				policyFileName := "skip_trustpolicy.json"
				notation.Exec("blob", "policy", "import", filepath.Join(NotationE2ETrustPolicyDir, policyFileName)).MatchKeyWords().
					MatchKeyWords("Successfully imported blob trust policy configuration to")
				// validate
				content, err := os.ReadFile(filepath.Join(NotationE2ETrustPolicyDir, policyFileName))
				Expect(err).NotTo(HaveOccurred())
				notation.Exec("blob", "policy", "show").MatchContent(string(content))
			})
		})

		It("should confirm import", func() {
			Host(opts, func(notation *utils.ExecOpts, artifact *Artifact, vhost *utils.VirtualHost) {
				policyFileName := "skip_trustpolicy.json"
				notation.WithInput(strings.NewReader("Y\n")).Exec("blob", "policy", "import", filepath.Join(NotationE2ETrustPolicyDir, policyFileName)).
					MatchKeyWords("Successfully imported blob trust policy configuration to")
				// validate
				content, err := os.ReadFile(filepath.Join(NotationE2ETrustPolicyDir, policyFileName))
				Expect(err).NotTo(HaveOccurred())
				notation.Exec("blob", "policy", "show").MatchContent(string(content))
			})
		})

		It("should confirm import by force", func() {
			Host(opts, func(notation *utils.ExecOpts, artifact *Artifact, vhost *utils.VirtualHost) {
				policyFileName := "skip_trustpolicy.json"
				notation.Exec("blob", "policy", "import", filepath.Join(NotationE2ETrustPolicyDir, policyFileName), "--force").
					MatchKeyWords("Successfully imported blob trust policy configuration to")
				// validate
				content, err := os.ReadFile(filepath.Join(NotationE2ETrustPolicyDir, policyFileName))
				Expect(err).NotTo(HaveOccurred())
				notation.Exec("blob", "policy", "show").MatchContent(string(content))
			})
		})
	})

	When("initializing trust policy", func() {
		Context("without existing policy", func() {
			opts := Opts()

			It("should fail when no name flag is provided", func() {
				Host(opts, func(notation *utils.ExecOpts, artifact *Artifact, vhost *utils.VirtualHost) {
					notation.ExpectFailure().
						Exec("blob", "policy", "init", "--trust-store", "ca:example-store", "--trusted-identity", "x509.subject: CN=example").
						MatchErrKeyWords("required flag(s)", "name", "not set")
				})
			})

			It("should fail when no trust-store flag is provided", func() {
				Host(opts, func(notation *utils.ExecOpts, artifact *Artifact, vhost *utils.VirtualHost) {
					notation.ExpectFailure().
						Exec("blob", "policy", "init", "--name", "example-policy", "--trusted-identity", "x509.subject: CN=example").
						MatchErrKeyWords("required flag(s)", "trust-store", "not set")
				})
			})

			It("should fail when no trusted-identity flag is provided", func() {
				Host(opts, func(notation *utils.ExecOpts, artifact *Artifact, vhost *utils.VirtualHost) {
					notation.ExpectFailure().
						Exec("blob", "policy", "init", "--name", "example-policy", "--trust-store", "ca:example-store").
						MatchErrKeyWords("required flag(s)", "trusted-identity", "not set")
				})
			})

			It("should fail when invalid trusted-identity format is provided", func() {
				Host(opts, func(notation *utils.ExecOpts, artifact *Artifact, vhost *utils.VirtualHost) {
					notation.ExpectFailure().
						Exec("blob", "policy", "init",
							"--name", "example-policy",
							"--trust-store", "ca:example-store",
							"--trusted-identity", "invalid").
						MatchErrKeyWords("invalid blob policy")
				})
			})

			It("should fail when directory doesn't have write permission", func() {
				Host(opts, func(notation *utils.ExecOpts, artifact *Artifact, vhost *utils.VirtualHost) {
					// Create the notation config directory if it doesn't exist
					configDir := vhost.AbsolutePath(NotationDirName)
					err := os.MkdirAll(configDir, 0755)
					Expect(err).NotTo(HaveOccurred())

					// Remove write permissions from the directory
					err = os.Chmod(configDir, 0500) // r-x for owner, no write
					Expect(err).NotTo(HaveOccurred())
					defer os.Chmod(configDir, 0755) // Restore permissions after test

					notation.ExpectFailure().
						Exec("blob", "policy", "init",
							"--name", "example-policy",
							"--trust-store", "ca:example-store",
							"--trusted-identity", "x509.subject: C=example,ST=example,O=example").
						MatchErrKeyWords("failed to write blob trust policy configuration")
				})
			})

			It("should successfully initialize policy when all required flags are provided", func() {
				Host(opts, func(notation *utils.ExecOpts, artifact *Artifact, vhost *utils.VirtualHost) {
					notation.Exec("blob", "policy", "init",
						"--name", "example-policy",
						"--global",
						"--trust-store", "ca:example-store",
						"--trust-store", "ca:example-store2",
						"--trusted-identity", "x509.subject: C=example,ST=example,O=example",
						"--trusted-identity", "x509.subject: C=example2,ST=example,O=example").
						MatchKeyWords("Successfully initialized blob trust policy file")

					// Verify the policy was created
					notation.Exec("blob", "policy", "show").
						MatchContent(`{
  "version": "1.0",
  "trustPolicies": [
    {
      "name": "example-policy",
      "signatureVerification": {
        "level": "strict"
      },
      "trustStores": [
        "ca:example-store",
        "ca:example-store2"
      ],
      "trustedIdentities": [
        "x509.subject: C=example,ST=example,O=example",
        "x509.subject: C=example2,ST=example,O=example"
      ],
      "globalPolicy": true
    }
  ]
}`)
				})
			})
		})

		Context("with existing policy", func() {
			opts := Opts(AddBlobTrustPolicyOption(validBlobTrustPolicyName))

			It("should canceled when trying to initialize with existing policy", func() {
				Host(opts, func(notation *utils.ExecOpts, artifact *Artifact, vhost *utils.VirtualHost) {
					notation.Exec("blob", "policy", "init",
						"--name", "new-policy",
						"--trust-store", "ca:new-store",
						"--trusted-identity", "x509.subject: C=example,ST=example,O=example").
						MatchKeyWords("The blob trust policy configuration already exists")
				})
			})

			It("should successfully initialize policy with force flag when policy exists", func() {
				Host(opts, func(notation *utils.ExecOpts, artifact *Artifact, vhost *utils.VirtualHost) {
					notation.Exec("blob", "policy", "init",
						"--name", "new-policy",
						"--trust-store", "ca:new-store",
						"--trusted-identity", "x509.subject: C=example, ST=example, O=example",
						"--force").
						MatchKeyWords("Successfully initialized blob trust policy file")

					// Verify the new policy was created and replaced the old one
					notation.Exec("blob", "policy", "show").
						MatchKeyWords(
							"new-policy",
							"ca:new-store",
							"x509.subject: C=example, ST=example, O=example",
						)
				})
			})
		})
	})
})
