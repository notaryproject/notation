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
	. "github.com/notaryproject/notation/test/e2e/internal/notation"
	"github.com/notaryproject/notation/test/e2e/internal/utils"
	. "github.com/notaryproject/notation/test/e2e/suite/common"
	. "github.com/onsi/ginkgo/v2"
)

var _ = Describe("notation trust policy multi-statements test", func() {
	It("multiple statements with the same registryScope", func() {
		Host(BaseOptions(), func(notation *utils.ExecOpts, _ *Artifact, vhost *utils.VirtualHost) {
			artifact := GenerateArtifact("", "test-repo8")
			// update trustpolicy.json
			vhost.SetOption(AddTrustPolicyOption("multi_statements_with_the_same_registry_scope_trustpolicy.json"))

			// test localhost:5000/test-repo
			notation.Exec("sign", artifact.ReferenceWithDigest()).MatchKeyWords(SignSuccessfully)
			notation.ExpectFailure().Exec("verify", artifact.ReferenceWithDigest()).
				MatchErrContent("Error: registry scope \"localhost:5000/test-repo8\" is present in multiple trust policy statements, one registry scope value can only be associated with one statement\n")
		})
	})

	It("multiple statements with wildcard registry scope", func() {
		Host(BaseOptions(), func(notation *utils.ExecOpts, _ *Artifact, vhost *utils.VirtualHost) {
			artifact := GenerateArtifact("e2e-valid-signature", "test-repo9")
			// update trustpolicy.json
			vhost.SetOption(AddTrustPolicyOption("multi_statements_with_wildcard_registry_scope_trustpolicy.json"))

			// test localhost:5000/test-repo
			notation.Exec("sign", artifact.ReferenceWithDigest()).MatchKeyWords(SignSuccessfully)
			notation.Exec("verify", artifact.ReferenceWithDigest()).
				MatchKeyWords(VerifySuccessfully)
		})
	})

	It("multiple statements with the same name", func() {
		Host(BaseOptions(), func(notation *utils.ExecOpts, _ *Artifact, vhost *utils.VirtualHost) {
			artifact := GenerateArtifact("", "test-repo10")
			// update trustpolicy.json
			vhost.SetOption(AddTrustPolicyOption("multi_statements_with_the_same_name_trustpolicy.json"))

			// test localhost:5000/test-repo
			notation.Exec("sign", artifact.ReferenceWithDigest()).MatchKeyWords(SignSuccessfully)
			notation.ExpectFailure().Exec("verify", artifact.ReferenceWithDigest()).
				MatchErrContent("Error: multiple trust policy statements use the same name \"e2e\", statement names must be unique\n")
		})
	})

	It("multiple statements with multi-wildcard registry scopes", func() {
		Host(BaseOptions(), func(notation *utils.ExecOpts, artifact *Artifact, vhost *utils.VirtualHost) {
			// update trustpolicy.json
			vhost.SetOption(AddTrustPolicyOption("multi_statements_with_multi_wildcard_registry_scope_trustpolicy.json"))

			// test localhost:5000/test-repo
			notation.Exec("sign", artifact.ReferenceWithDigest()).MatchKeyWords(SignSuccessfully)
			notation.ExpectFailure().Exec("verify", artifact.ReferenceWithDigest()).
				MatchErrContent("Error: registry scope \"*\" is present in multiple trust policy statements, one registry scope value can only be associated with one statement\n")
		})
	})
})
