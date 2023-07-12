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

package trustpolicy

import (
	"path/filepath"

	. "github.com/notaryproject/notation/test/e2e/internal/notation"
	"github.com/notaryproject/notation/test/e2e/internal/utils"
	. "github.com/notaryproject/notation/test/e2e/suite/common"
	. "github.com/onsi/ginkgo/v2"
)

var _ = Describe("notation trust policy trusted identity test", func() {
	It("with unset trusted identity", func() {
		Host(BaseOptions(), func(notation *utils.ExecOpts, _ *Artifact, vhost *utils.VirtualHost) {
			vhost.SetOption(AddTrustPolicyOption("unset_trusted_identity_trustpolicy.json"))
			artifact := GenerateArtifact("e2e-valid-signature", "")

			notation.ExpectFailure().Exec("verify", artifact.ReferenceWithDigest(), "-v").
				MatchErrKeyWords(`trust policy statement "e2e" is either missing trust stores or trusted identities, both must be specified`)
		})
	})

	It("with valid trusted identity", func() {
		Host(BaseOptionsWithExperimental(), func(notation *utils.ExecOpts, _ *Artifact, vhost *utils.VirtualHost) {
			vhost.SetOption(AddTrustPolicyOption("valid_trusted_identity_trustpolicy.json"))
			artifact := GenerateArtifact("e2e-valid-signature", "")

			notation.Exec("verify", "--allow-referrers-api", artifact.ReferenceWithDigest(), "-v").
				MatchKeyWords(VerifySuccessfully)
		})
	})

	It("with invalid trusted identity", func() {
		Host(BaseOptionsWithExperimental(), func(notation *utils.ExecOpts, _ *Artifact, vhost *utils.VirtualHost) {
			vhost.SetOption(AddTrustPolicyOption("invalid_trusted_identity_trustpolicy.json"))
			artifact := GenerateArtifact("e2e-valid-signature", "")

			notation.ExpectFailure().Exec("verify", "--allow-referrers-api", artifact.ReferenceWithDigest(), "-v").
				MatchErrKeyWords("Failure reason: signing certificate from the digital signature does not match the X.509 trusted identities",
					VerifyFailed)
		})
	})

	It("with malformed trusted identity", func() {
		Host(BaseOptions(), func(notation *utils.ExecOpts, _ *Artifact, vhost *utils.VirtualHost) {
			vhost.SetOption(AddTrustPolicyOption("malformed_trusted_identity_trustpolicy.json"))
			artifact := GenerateArtifact("e2e-valid-signature", "")

			notation.ExpectFailure().Exec("verify", artifact.ReferenceWithDigest(), "-v").
				MatchErrKeyWords(`trust policy statement "e2e" has trusted identity "x509.subject CN=e2e,O=Notary,L=Seattle,ST=WA,C=US" missing separator`)
		})
	})

	It("with empty trusted identity", func() {
		Host(BaseOptions(), func(notation *utils.ExecOpts, _ *Artifact, vhost *utils.VirtualHost) {
			vhost.SetOption(AddTrustPolicyOption("empty_trusted_identity_trustpolicy.json"))
			artifact := GenerateArtifact("e2e-valid-signature", "")

			notation.ExpectFailure().Exec("verify", artifact.ReferenceWithDigest(), "-v").
				MatchErrKeyWords(`trust policy statement "e2e" has an empty trusted identity`)
		})
	})

	It("with multiple trusted identity", func() {
		Host(nil, func(notation *utils.ExecOpts, artifact1 *Artifact, vhost *utils.VirtualHost) {
			// artifact1 signed with new_e2e.crt
			OldNotation(AuthOption("", ""), AddKeyOption("e2e.key", "new_e2e.crt")).
				Exec("sign", artifact1.ReferenceWithDigest(), "-v").
				MatchKeyWords(SignSuccessfully)

			// artifact2 signed with e2e.crt
			artifact2 := GenerateArtifact("e2e-valid-signature", "")

			// setup multiple trusted identity
			vhost.SetOption(AuthOption("", ""),
				AddTrustPolicyOption("multiple_trusted_identity_trustpolicy.json"),
				AddTrustStoreOption("e2e", filepath.Join(NotationE2ELocalKeysDir, "new_e2e.crt")),
				AddTrustStoreOption("e2e", filepath.Join(NotationE2ELocalKeysDir, "e2e.crt")),
				EnableExperimental(),
			)

			notation.Exec("verify", "--allow-referrers-api", artifact1.ReferenceWithDigest(), "-v").
				MatchKeyWords(VerifySuccessfully)

			notation.Exec("verify", "--allow-referrers-api", artifact2.ReferenceWithDigest(), "-v").
				MatchKeyWords(VerifySuccessfully)
		})
	})

	It("with overlapped trusted identities", func() {
		Host(BaseOptions(), func(notation *utils.ExecOpts, _ *Artifact, vhost *utils.VirtualHost) {
			vhost.SetOption(AddTrustPolicyOption("overlapped_trusted_identity_trustpolicy.json"))
			artifact := GenerateArtifact("e2e-valid-signature", "")

			notation.ExpectFailure().Exec("verify", artifact.ReferenceWithDigest(), "-v").
				MatchErrKeyWords(`trust policy statement "e2e" has overlapping x509 trustedIdentities`)
		})
	})

	It("with wildcard plus other trusted identities", func() {
		Host(BaseOptions(), func(notation *utils.ExecOpts, _ *Artifact, vhost *utils.VirtualHost) {
			vhost.SetOption(AddTrustPolicyOption("wildcard_plus_other_trusted_identity_trustpolicy.json"))
			artifact := GenerateArtifact("e2e-valid-signature", "")

			notation.ExpectFailure().Exec("verify", artifact.ReferenceWithDigest(), "-v").
				MatchErrKeyWords(`trust policy statement "e2e" uses a wildcard trusted identity '*', a wildcard identity cannot be used in conjunction with other values`)
		})
	})

	It("with trusted identities missing organization", func() {
		Host(BaseOptions(), func(notation *utils.ExecOpts, _ *Artifact, vhost *utils.VirtualHost) {
			vhost.SetOption(AddTrustPolicyOption("missing_organization_trusted_identity_trustpolicy.json"))
			artifact := GenerateArtifact("e2e-valid-signature", "")

			notation.ExpectFailure().Exec("verify", artifact.ReferenceWithDigest(), "-v").
				MatchErrKeyWords(`distinguished name (DN) " CN=e2e,L=Seattle,ST=WA,C=US" has no mandatory RDN attribute for "O", it must contain 'C', 'ST', and 'O' RDN attributes at a minimum`)
		})
	})

	It("with trusted identities missing state", func() {
		Host(BaseOptions(), func(notation *utils.ExecOpts, _ *Artifact, vhost *utils.VirtualHost) {
			vhost.SetOption(AddTrustPolicyOption("missing_state_trusted_identity_trustpolicy.json"))
			artifact := GenerateArtifact("e2e-valid-signature", "")

			notation.ExpectFailure().Exec("verify", artifact.ReferenceWithDigest(), "-v").
				MatchErrKeyWords(`distinguished name (DN) " CN=e2e,O=Notary,L=Seattle,C=US" has no mandatory RDN attribute for "ST", it must contain 'C', 'ST', and 'O' RDN attributes at a minimum`)
		})
	})

	It("with trusted identities missing country", func() {
		Host(BaseOptions(), func(notation *utils.ExecOpts, _ *Artifact, vhost *utils.VirtualHost) {
			vhost.SetOption(AddTrustPolicyOption("missing_country_trusted_identity_trustpolicy.json"))
			artifact := GenerateArtifact("e2e-valid-signature", "")

			notation.ExpectFailure().Exec("verify", artifact.ReferenceWithDigest(), "-v").
				MatchErrKeyWords(`distinguished name (DN) " CN=e2e,O=Notary,L=Seattle,ST=WA" has no mandatory RDN attribute for "C", it must contain 'C', 'ST', and 'O' RDN attributes at a minimum`)
		})
	})
})
