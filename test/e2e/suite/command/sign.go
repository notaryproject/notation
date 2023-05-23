package command

import (
	"fmt"
	"time"

	. "github.com/notaryproject/notation/test/e2e/internal/notation"
	"github.com/notaryproject/notation/test/e2e/internal/utils"
	. "github.com/notaryproject/notation/test/e2e/suite/common"
	. "github.com/onsi/ginkgo/v2"
)

var _ = Describe("notation sign", func() {
	It("by digest", func() {
		Host(BaseOptions(), func(notation *utils.ExecOpts, artifact *Artifact, vhost *utils.VirtualHost) {
			notation.Exec("sign", artifact.ReferenceWithDigest()).
				MatchKeyWords(SignSuccessfully)

			OldNotation().WithDescription("verify by digest").
				Exec("verify", artifact.ReferenceWithDigest()).
				MatchKeyWords(VerifySuccessfully)

			OldNotation().WithDescription("verify by tag").
				Exec("verify", artifact.ReferenceWithTag()).
				MatchKeyWords(VerifySuccessfully)
		})
	})

	It("by digest with COSE format", func() {
		Host(BaseOptions(), func(notation *utils.ExecOpts, artifact *Artifact, vhost *utils.VirtualHost) {
			notation.Exec("sign", "--signature-format", "cose", artifact.ReferenceWithDigest()).
				MatchKeyWords(SignSuccessfully)

			OldNotation().WithDescription("verify by digest").
				Exec("verify", artifact.ReferenceWithTag()).
				MatchKeyWords(VerifySuccessfully)

			OldNotation().WithDescription("verify by tag").
				Exec("verify", artifact.ReferenceWithTag()).
				MatchKeyWords(VerifySuccessfully)
		})
	})

	It("by tag with JWS format", func() {
		Host(BaseOptions(), func(notation *utils.ExecOpts, artifact *Artifact, vhost *utils.VirtualHost) {
			notation.WithDescription("sign with JWS").
				Exec("sign", artifact.ReferenceWithTag(), "--signature-format", "jws").
				MatchKeyWords(SignSuccessfully)

			OldNotation().WithDescription("verify JWS signature").
				Exec("verify", artifact.ReferenceWithTag()).
				MatchKeyWords(VerifySuccessfully)
		})
	})

	It("by tag with COSE signature format", func() {
		Host(BaseOptions(), func(notation *utils.ExecOpts, artifact *Artifact, vhost *utils.VirtualHost) {
			notation.WithDescription("sign with COSE").
				Exec("sign", artifact.ReferenceWithTag(), "--signature-format", "cose").
				MatchKeyWords(SignSuccessfully)

			OldNotation().WithDescription("verify COSE signature").
				Exec("verify", artifact.ReferenceWithTag()).
				MatchKeyWords(VerifySuccessfully)
		})
	})

	It("with specific key", func() {
		Host(BaseOptions(), func(notation *utils.ExecOpts, artifact *Artifact, vhost *utils.VirtualHost) {
			const keyName = "sKey"
			notation.Exec("cert", "generate-test", keyName).
				MatchKeyWords(fmt.Sprintf("notation/localkeys/%s.crt", keyName))

			notation.Exec("sign", "--key", keyName, artifact.ReferenceWithDigest()).
				MatchKeyWords(SignSuccessfully)

			// copy the generated cert file and create the new trust policy for verify signature with generated new key.
			OldNotation(AuthOption("", ""),
				AddTrustStoreOption(keyName, vhost.AbsolutePath(NotationDirName, LocalKeysDirName, keyName+".crt")),
				AddTrustPolicyOption("generate_test_trustpolicy.json"),
			).Exec("verify", artifact.ReferenceWithTag()).
				MatchKeyWords(VerifySuccessfully)
		})
	})

	It("with expiry in 24h", func() {
		Host(BaseOptions(), func(notation *utils.ExecOpts, artifact *Artifact, vhost *utils.VirtualHost) {
			notation.Exec("sign", "--expiry", "24h", artifact.ReferenceWithDigest()).
				MatchKeyWords(SignSuccessfully)

			OldNotation().Exec("verify", artifact.ReferenceWithTag()).
				MatchKeyWords(VerifySuccessfully)
		})
	})

	It("with expiry in 2s", func() {
		Host(BaseOptions(), func(notation *utils.ExecOpts, artifact *Artifact, vhost *utils.VirtualHost) {
			notation.Exec("sign", "--expiry", "2s", artifact.ReferenceWithDigest()).
				MatchKeyWords(SignSuccessfully)

			// sleep to wait for expiry
			time.Sleep(2100 * time.Millisecond)

			OldNotation().ExpectFailure().Exec("verify", artifact.ReferenceWithDigest(), "-v").
				MatchErrKeyWords("expiry validation failed.").
				MatchErrKeyWords("signature verification failed for all the signatures")
		})
	})

	It("by digest with oci layout", func() {
		GeneralHost(BaseOptionsWithExperimental(), func(notation *utils.ExecOpts, vhost *utils.VirtualHost) {
			const digest = "sha256:cc2ae4e91a31a77086edbdbf4711de48e5fa3ebdacad3403e61777a9e1a53b6f"
			ociLayoutReference := OCILayoutTestPath + "@" + digest
			notation.Exec("sign", "--oci-layout", ociLayoutReference).
				MatchKeyWords(SignSuccessfully)
		})
	})

	It("by digest with oci layout and COSE format", func() {
		GeneralHost(BaseOptionsWithExperimental(), func(notation *utils.ExecOpts, vhost *utils.VirtualHost) {
			const digest = "sha256:cc2ae4e91a31a77086edbdbf4711de48e5fa3ebdacad3403e61777a9e1a53b6f"
			ociLayoutReference := OCILayoutTestPath + "@" + digest
			notation.Exec("sign", "--oci-layout", "--signature-format", "cose", ociLayoutReference).
				MatchKeyWords(SignSuccessfully)
		})
	})

	It("by tag with oci layout", func() {
		GeneralHost(BaseOptionsWithExperimental(), func(notation *utils.ExecOpts, vhost *utils.VirtualHost) {
			ociLayoutReference := OCILayoutTestPath + ":" + TestTag
			notation.Exec("sign", "--oci-layout", ociLayoutReference).
				MatchKeyWords(SignSuccessfully)
		})
	})

	It("by tag with oci layout and COSE format", func() {
		GeneralHost(BaseOptionsWithExperimental(), func(notation *utils.ExecOpts, vhost *utils.VirtualHost) {
			ociLayoutReference := OCILayoutTestPath + ":" + TestTag
			notation.Exec("sign", "--oci-layout", "--signature-format", "cose", ociLayoutReference).
				MatchKeyWords(SignSuccessfully)
		})
	})

	It("by digest with oci layout but without experimental", func() {
		GeneralHost(BaseOptions(), func(notation *utils.ExecOpts, vhost *utils.VirtualHost) {
			const digest = "sha256:cc2ae4e91a31a77086edbdbf4711de48e5fa3ebdacad3403e61777a9e1a53b6f"
			expectedErrMsg := "Error: flag(s) --oci-layout in \"notation sign\" is experimental and not enabled by default. To use, please set NOTATION_EXPERIMENTAL=1 environment variable\n"
			ociLayoutReference := OCILayoutTestPath + "@" + digest
			notation.ExpectFailure().Exec("sign", "--oci-layout", ociLayoutReference).
				MatchErrContent(expectedErrMsg)
		})
	})

	It("with TLS by digest", func() {
		HostWithTLS(BaseOptions(), func(notation *utils.ExecOpts, artifact *Artifact, vhost *utils.VirtualHost) {
			notation.Exec("sign", artifact.DomainReferenceWithDigest()).
				MatchKeyWords(SignSuccessfully)

			OldNotation().Exec("verify", artifact.DomainReferenceWithDigest()).
				MatchKeyWords(VerifySuccessfully)
		})
	})

	It("with --insecure-registry by digest", func() {
		HostWithTLS(BaseOptions(), func(notation *utils.ExecOpts, artifact *Artifact, vhost *utils.VirtualHost) {
			notation.Exec("sign", "--insecure-registry", artifact.DomainReferenceWithDigest()).
				MatchKeyWords(SignSuccessfully)

			OldNotation().Exec("verify", artifact.DomainReferenceWithDigest()).
				MatchKeyWords(VerifySuccessfully)
		})
	})
})
