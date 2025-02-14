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

package scenario_test

import (
	. "github.com/notaryproject/notation/test/e2e/internal/notation"
	"github.com/notaryproject/notation/test/e2e/internal/utils"
	. "github.com/notaryproject/notation/test/e2e/suite/common"
	. "github.com/onsi/ginkgo/v2"
)

var _ = Describe("notation OCSP revocation check", Serial, func() {
	It("successfully", func() {
		Host(OCSPOptions(), func(notation *utils.ExecOpts, artifact *Artifact, vhost *utils.VirtualHost) {
			notation.Exec("sign", artifact.ReferenceWithDigest()).
				MatchKeyWords(SignSuccessfully)

			utils.LeafOCSPUnrevoke()

			// verify without cache
			notation.Exec("verify", artifact.ReferenceWithDigest(), "-d").
				MatchKeyWords(
					VerifySuccessfully,
				).
				MatchErrKeyWords(
					`"Content-Type": "application/ocsp-response"`,
					"No verification impacting errors encountered while checking revocation, status is OK",
				)
		})
	})

	It("with leaf certificate revoked", func() {
		Host(OCSPOptions(), func(notation *utils.ExecOpts, artifact *Artifact, vhost *utils.VirtualHost) {
			notation.Exec("sign", artifact.ReferenceWithDigest()).
				MatchKeyWords(SignSuccessfully)

			utils.LeafOCSPRevoke()

			// verify without cache
			notation.ExpectFailure().Exec("verify", artifact.ReferenceWithDigest(), "-d").
				MatchErrKeyWords(
					VerifyFailed,
					`Certificate #1 in chain with subject "CN=LeafCert,OU=OrgUnit,O=Organization,L=City,ST=State,C=US" encountered an error for revocation method OCSP at URL "http://localhost:10087": certificate is revoked via OCSP`,
				)
		})
	})

	It("with unknown status", func() {
		Host(OCSPOptions(), func(notation *utils.ExecOpts, artifact *Artifact, vhost *utils.VirtualHost) {
			notation.Exec("sign", artifact.ReferenceWithDigest()).
				MatchKeyWords(SignSuccessfully)

			utils.LeafOCSPUnknown()

			// verify without cache
			notation.ExpectFailure().Exec("verify", artifact.ReferenceWithDigest(), "-d").
				MatchErrKeyWords(
					VerifyFailed,
					`error: signing certificate with subject "CN=LeafCert,OU=OrgUnit,O=Organization,L=City,ST=State,C=US" revocation status is unknown`,
				)
		})
	})
})
