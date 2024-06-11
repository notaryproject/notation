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
	"fmt"

	. "github.com/notaryproject/notation/test/e2e/internal/notation"
	"github.com/notaryproject/notation/test/e2e/internal/utils"
	. "github.com/notaryproject/notation/test/e2e/suite/common"
	. "github.com/onsi/ginkgo/v2"
)

var _ = Describe("notation verify", func() {
	It("by digest", func() {
		Host(BaseOptions(), func(notation *utils.ExecOpts, artifact *Artifact, vhost *utils.VirtualHost) {
			notation.Exec("sign", artifact.ReferenceWithDigest()).
				MatchKeyWords(SignSuccessfully)

			notation.Exec("verify", artifact.ReferenceWithDigest(), "-v").
				MatchKeyWords(VerifySuccessfully)
		})
	})

	It("by tag", func() {
		Host(BaseOptions(), func(notation *utils.ExecOpts, artifact *Artifact, vhost *utils.VirtualHost) {
			notation.Exec("sign", artifact.ReferenceWithDigest()).
				MatchKeyWords(SignSuccessfully)

			notation.Exec("verify", artifact.ReferenceWithTag(), "-v").
				MatchKeyWords(VerifySuccessfully)
		})
	})

	It("with debug log", func() {
		Host(BaseOptions(), func(notation *utils.ExecOpts, artifact *Artifact, vhost *utils.VirtualHost) {
			notation.Exec("sign", artifact.ReferenceWithDigest()).
				MatchKeyWords(SignSuccessfully)

			notation.Exec("verify", artifact.ReferenceWithDigest(), "-d").
				// debug log message outputs to stderr
				MatchErrKeyWords(
					"Check verification level",
					fmt.Sprintf("Verify signature against artifact %s", artifact.Digest),
					"Validating cert chain",
					"Validating trust identity",
					"Validating expiry",
					"Validating authentic timestamp",
					"Validating revocation",
				).
				MatchKeyWords(VerifySuccessfully)
		})
	})

	It("by digest with the Referrers API", func() {
		Host(BaseOptionsWithExperimental(), func(notation *utils.ExecOpts, artifact *Artifact, vhost *utils.VirtualHost) {
			notation.Exec("sign", "--allow-referrers-api", artifact.ReferenceWithDigest()).
				MatchKeyWords(SignSuccessfully)

			notation.Exec("verify", "--allow-referrers-api", artifact.ReferenceWithDigest(), "-v").
				MatchKeyWords(VerifySuccessfully)
		})
	})

	It("by digest, sign with the Referrers tag schema, verify with the Referrers API", func() {
		Host(BaseOptionsWithExperimental(), func(notation *utils.ExecOpts, artifact *Artifact, vhost *utils.VirtualHost) {
			notation.Exec("sign", artifact.ReferenceWithDigest()).
				MatchKeyWords(SignSuccessfully)

			notation.Exec("verify", "--allow-referrers-api", artifact.ReferenceWithDigest(), "-v").
				MatchKeyWords(VerifySuccessfully)
		})
	})

	It("by digest with oci layout", func() {
		HostWithOCILayout(BaseOptionsWithExperimental(), func(notation *utils.ExecOpts, ociLayout *OCILayout, vhost *utils.VirtualHost) {
			notation.Exec("sign", "--oci-layout", ociLayout.ReferenceWithDigest()).
				MatchKeyWords(SignSuccessfully)

			experimentalMsg := "Warning: This feature is experimental and may not be fully tested or completed and may be deprecated. Report any issues to \"https://github/notaryproject/notation\"\n"
			notation.Exec("verify", "--oci-layout", "--scope", "local/e2e", ociLayout.ReferenceWithDigest()).
				MatchKeyWords(VerifySuccessfully).
				MatchErrKeyWords(experimentalMsg)
		})
	})

	It("by tag with oci layout and COSE format", func() {
		HostWithOCILayout(BaseOptionsWithExperimental(), func(notation *utils.ExecOpts, ociLayout *OCILayout, vhost *utils.VirtualHost) {
			notation.Exec("sign", "--oci-layout", "--signature-format", "cose", ociLayout.ReferenceWithTag()).
				MatchKeyWords(SignSuccessfully)

			experimentalMsg := "Warning: This feature is experimental and may not be fully tested or completed and may be deprecated. Report any issues to \"https://github/notaryproject/notation\"\n"
			notation.Exec("verify", "--oci-layout", "--scope", "local/e2e", ociLayout.ReferenceWithTag()).
				MatchKeyWords(VerifySuccessfully).
				MatchErrKeyWords(experimentalMsg)
		})
	})

	It("by digest with oci layout but without experimental", func() {
		HostWithOCILayout(BaseOptions(), func(notation *utils.ExecOpts, ociLayout *OCILayout, vhost *utils.VirtualHost) {
			expectedErrMsg := "Error: flag(s) --oci-layout,--scope in \"notation verify\" is experimental and not enabled by default. To use, please set NOTATION_EXPERIMENTAL=1 environment variable\n"
			notation.ExpectFailure().Exec("verify", "--oci-layout", "--scope", "local/e2e", ociLayout.ReferenceWithDigest()).
				MatchErrContent(expectedErrMsg)
		})
	})

	It("by digest with oci layout but missing scope", func() {
		HostWithOCILayout(BaseOptionsWithExperimental(), func(notation *utils.ExecOpts, ociLayout *OCILayout, vhost *utils.VirtualHost) {
			notation.Exec("sign", "--oci-layout", ociLayout.ReferenceWithDigest()).
				MatchKeyWords(SignSuccessfully)

			experimentalMsg := "Warning: This feature is experimental and may not be fully tested or completed and may be deprecated. Report any issues to \"https://github/notaryproject/notation\"\n"
			expectedErrMsg := "Error: if any flags in the group [oci-layout scope] are set they must all be set; missing [scope]"
			notation.ExpectFailure().Exec("verify", "--oci-layout", ociLayout.ReferenceWithDigest()).
				MatchErrKeyWords(experimentalMsg).
				MatchErrKeyWords(expectedErrMsg)
		})
	})

	It("with TLS by digest", func() {
		HostInGithubAction(BaseOptions(), func(notation *utils.ExecOpts, artifact *Artifact, vhost *utils.VirtualHost) {
			notation.Exec("sign", artifact.DomainReferenceWithDigest()).
				MatchKeyWords(SignSuccessfully)

			notation.Exec("verify", "-d", artifact.DomainReferenceWithDigest()).
				MatchKeyWords(
					VerifySuccessfully,
				).
				MatchErrKeyWords(HTTPSRequest).
				NoMatchErrKeyWords(HTTPRequest)
		})
	})

	It("with --insecure-registry by digest", func() {
		HostInGithubAction(BaseOptions(), func(notation *utils.ExecOpts, artifact *Artifact, vhost *utils.VirtualHost) {
			notation.Exec("sign", artifact.DomainReferenceWithDigest()).
				MatchKeyWords(SignSuccessfully)

			notation.Exec("verify", "-d", "--insecure-registry", artifact.DomainReferenceWithDigest()).
				MatchKeyWords(
					VerifySuccessfully,
				).
				MatchErrKeyWords(HTTPRequest).
				NoMatchErrKeyWords(HTTPSRequest)
		})
	})

	It("incorrect NOTATION_CONFIG path", func() {
		Host(BaseOptions(), func(notation *utils.ExecOpts, artifact *Artifact, vhost *utils.VirtualHost) {
			notation.Exec("sign", artifact.ReferenceWithDigest()).
				MatchKeyWords(SignSuccessfully)

			vhost.UpdateEnv(map[string]string{"NOTATION_CONFIG": "/not/exist"})
			notation.ExpectFailure().Exec("verify", artifact.ReferenceWithDigest(), "-v").
				MatchErrKeyWords("trust policy is not present")
		})
	})

	It("correct NOTATION_CONFIG path", func() {
		Host(BaseOptions(), func(notation *utils.ExecOpts, artifact *Artifact, vhost *utils.VirtualHost) {
			notation.Exec("sign", artifact.ReferenceWithDigest()).
				MatchKeyWords(SignSuccessfully)

			vhost.UpdateEnv(map[string]string{"NOTATION_CONFIG": vhost.AbsolutePath(NotationDirName)})
			notation.Exec("verify", artifact.ReferenceWithDigest(), "-v").
				MatchKeyWords(VerifySuccessfully)
		})
	})
})
