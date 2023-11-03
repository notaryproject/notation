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

var _ = Describe("notation trust policy trust store test", func() {
	It("unset trust store", func() {
		Host(BaseOptions(), func(notation *utils.ExecOpts, _ *Artifact, vhost *utils.VirtualHost) {
			vhost.SetOption(AddTrustPolicyOption("unset_trust_store_trustpolicy.json"))

			artifact := GenerateArtifact("e2e-valid-signature", "")

			notation.ExpectFailure().Exec("verify", artifact.ReferenceWithDigest(), "-v").
				MatchErrKeyWords(`trust policy statement "e2e" is either missing trust stores or trusted identities, both must be specified`)
		})
	})

	It("invalid trust store", func() {
		Host(BaseOptionsWithExperimental(), func(notation *utils.ExecOpts, _ *Artifact, vhost *utils.VirtualHost) {
			vhost.SetOption(AddTrustPolicyOption("invalid_trust_store_trustpolicy.json"))

			artifact := GenerateArtifact("e2e-valid-signature", "")

			notation.ExpectFailure().Exec("verify", "--allow-referrers-api", artifact.ReferenceWithDigest(), "-v").
				MatchErrKeyWords("the trust store \"invalid_store\" of type \"ca\" does not exist")
		})
	})

	It("malformed trust store", func() {
		Host(BaseOptions(), func(notation *utils.ExecOpts, _ *Artifact, vhost *utils.VirtualHost) {
			vhost.SetOption(AddTrustPolicyOption("malformed_trust_store_trustpolicy.json"))

			artifact := GenerateArtifact("e2e-valid-signature", "")

			notation.ExpectFailure().Exec("verify", artifact.ReferenceWithDigest(), "-v").
				MatchErrKeyWords(`trust policy statement "e2e" uses an unsupported trust store name "" in trust store value "ca:". Named store name needs to follow [a-zA-Z0-9_.-]+ format`)
		})
	})

	It("wildcard (malformed) trust store", func() {
		Host(BaseOptions(), func(notation *utils.ExecOpts, _ *Artifact, vhost *utils.VirtualHost) {
			vhost.SetOption(AddTrustPolicyOption("wildcard_trust_store_trustpolicy.json"))

			artifact := GenerateArtifact("e2e-valid-signature", "")

			notation.ExpectFailure().Exec("verify", artifact.ReferenceWithDigest(), "-v").
				MatchErrKeyWords(`trust policy statement "e2e" has malformed trust store value "*". The required format is <TrustStoreType>:<TrustStoreName>`)
		})
	})

	It("multiple trust stores", func() {
		Host(nil, func(notation *utils.ExecOpts, artifact1 *Artifact, vhost *utils.VirtualHost) {
			// artifact1 signed with new_e2e.crt
			OldNotation(AuthOption("", ""), AddKeyOption("e2e.key", "new_e2e.crt")).
				Exec("sign", artifact1.ReferenceWithDigest(), "-v").
				MatchKeyWords(SignSuccessfully)

			// artifact2 signed with e2e.crt
			artifact2 := GenerateArtifact("e2e-valid-signature", "")

			// setup multiple trust store
			vhost.SetOption(AuthOption("", ""),
				AddTrustPolicyOption("multiple_trust_store_trustpolicy.json"),
				AddTrustStoreOption("e2e-new", filepath.Join(NotationE2ELocalKeysDir, "new_e2e.crt")),
				AddTrustStoreOption("e2e", filepath.Join(NotationE2ELocalKeysDir, "e2e.crt")),
				EnableExperimental())

			notation.WithDescription("verify artifact1 with trust store ca/e2e-new").
				Exec("verify", "--allow-referrers-api", artifact1.ReferenceWithDigest(), "-v").
				MatchKeyWords(VerifySuccessfully)

			notation.WithDescription("verify artifact2 with trust store ca/e2e").
				Exec("verify", "--allow-referrers-api", artifact2.ReferenceWithDigest(), "-v").
				MatchKeyWords(VerifySuccessfully)
		})
	})

	It("overlapped trust stores", func() {
		Skip("overlapped trust stores were not checked")
		Host(nil, func(notation *utils.ExecOpts, artifact *Artifact, vhost *utils.VirtualHost) {
			// artifact signed with new_e2e.crt
			notation.Exec("sign", artifact.ReferenceWithDigest(), "-v").
				MatchKeyWords(SignSuccessfully)

			// setup overlapped trust store
			vhost.SetOption(AuthOption("", ""),
				AddTrustPolicyOption("overlapped_trust_store_trustpolicy.json"),
				AddTrustStoreOption("e2e", filepath.Join(NotationE2ELocalKeysDir, "e2e.crt")))

			notation.ExpectFailure().Exec("verify", artifact.ReferenceWithDigest(), "-v").
				MatchErrKeyWords(VerifyFailed)
		})
	})
})
