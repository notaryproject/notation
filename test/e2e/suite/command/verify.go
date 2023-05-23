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
		GeneralHost(BaseOptionsWithExperimental(), func(notation *utils.ExecOpts, vhost *utils.VirtualHost) {
			const digest = "sha256:cc2ae4e91a31a77086edbdbf4711de48e5fa3ebdacad3403e61777a9e1a53b6f"
			ociLayoutReference := OCILayoutTestPath + "@" + digest
			notation.Exec("sign", "--oci-layout", ociLayoutReference).
				MatchKeyWords(SignSuccessfully)

			experimentalMsg := "Warning: This feature is experimental and may not be fully tested or completed and may be deprecated. Report any issues to \"https://github/notaryproject/notation\"\n"
			notation.Exec("verify", "--oci-layout", "--scope", "local/e2e", ociLayoutReference).
				MatchKeyWords(VerifySuccessfully).
				MatchErrKeyWords(experimentalMsg)
		})
	})

	It("by tag with oci layout and COSE format", func() {
		GeneralHost(BaseOptionsWithExperimental(), func(notation *utils.ExecOpts, vhost *utils.VirtualHost) {
			ociLayoutReference := OCILayoutTestPath + ":" + TestTag
			notation.Exec("sign", "--oci-layout", "--signature-format", "cose", ociLayoutReference).
				MatchKeyWords(SignSuccessfully)

			experimentalMsg := "Warning: This feature is experimental and may not be fully tested or completed and may be deprecated. Report any issues to \"https://github/notaryproject/notation\"\n"
			notation.Exec("verify", "--oci-layout", "--scope", "local/e2e", ociLayoutReference).
				MatchKeyWords(VerifySuccessfully).
				MatchErrKeyWords(experimentalMsg)
		})
	})

	It("by digest with oci layout but without experimental", func() {
		GeneralHost(BaseOptions(), func(notation *utils.ExecOpts, vhost *utils.VirtualHost) {
			const digest = "sha256:cc2ae4e91a31a77086edbdbf4711de48e5fa3ebdacad3403e61777a9e1a53b6f"
			expectedErrMsg := "Error: flag(s) --oci-layout,--scope in \"notation verify\" is experimental and not enabled by default. To use, please set NOTATION_EXPERIMENTAL=1 environment variable\n"
			ociLayoutReference := OCILayoutTestPath + "@" + digest
			notation.ExpectFailure().Exec("verify", "--oci-layout", "--scope", "local/e2e", ociLayoutReference).
				MatchErrContent(expectedErrMsg)
		})
	})

	It("by digest with oci layout but missing scope", func() {
		GeneralHost(BaseOptionsWithExperimental(), func(notation *utils.ExecOpts, vhost *utils.VirtualHost) {
			const digest = "sha256:cc2ae4e91a31a77086edbdbf4711de48e5fa3ebdacad3403e61777a9e1a53b6f"
			ociLayoutReference := OCILayoutTestPath + "@" + digest
			notation.Exec("sign", "--oci-layout", ociLayoutReference).
				MatchKeyWords(SignSuccessfully)

			experimentalMsg := "Warning: This feature is experimental and may not be fully tested or completed and may be deprecated. Report any issues to \"https://github/notaryproject/notation\"\n"
			expectedErrMsg := "Error: if any flags in the group [oci-layout scope] are set they must all be set; missing [scope]"
			notation.ExpectFailure().Exec("verify", "--oci-layout", ociLayoutReference).
				MatchErrKeyWords(experimentalMsg).
				MatchErrKeyWords(expectedErrMsg)
		})
	})

	It("with TLS by digest", func() {
		HostWithTLS(BaseOptions(), func(notation *utils.ExecOpts, artifact *Artifact, vhost *utils.VirtualHost) {
			notation.Exec("sign", artifact.DomainReferenceWithDigest()).
				MatchKeyWords(SignSuccessfully)

			notation.Exec("verify", artifact.DomainReferenceWithDigest(), "-v").
				MatchKeyWords(VerifySuccessfully)
		})
	})

	It("with --insecure-registry by digest", func() {
		HostWithTLS(BaseOptions(), func(notation *utils.ExecOpts, artifact *Artifact, vhost *utils.VirtualHost) {
			notation.Exec("sign", artifact.DomainReferenceWithDigest()).
				MatchKeyWords(SignSuccessfully)

			notation.Exec("verify", "--insecure-registry", artifact.DomainReferenceWithDigest(), "-v").
				MatchKeyWords(VerifySuccessfully)
		})
	})
})
