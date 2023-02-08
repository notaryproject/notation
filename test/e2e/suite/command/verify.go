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
			OldNotation().Exec("sign", artifact.ReferenceWithDigest()).
				MatchKeyWords(SignSuccessfully)

			notation.Exec("verify", artifact.ReferenceWithDigest()).
				MatchKeyWords(VerifySuccessfully)
		})
	})

	It("by tag", func() {
		Host(BaseOptions(), func(notation *utils.ExecOpts, artifact *Artifact, vhost *utils.VirtualHost) {
			OldNotation().Exec("sign", artifact.ReferenceWithDigest()).
				MatchKeyWords(SignSuccessfully)

			notation.Exec("verify", artifact.ReferenceWithTag()).
				MatchKeyWords(VerifySuccessfully)
		})
	})

	It("with debug log", func() {
		Host(BaseOptions(), func(notation *utils.ExecOpts, artifact *Artifact, vhost *utils.VirtualHost) {
			OldNotation().Exec("sign", artifact.ReferenceWithDigest()).
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

	It("with added user metadata", func() {
		Host(BaseOptions(), func(notation *utils.ExecOpts, artifact *Artifact, vhost *utils.VirtualHost) {
			notation.Exec("sign", artifact.ReferenceWithDigest(), "--user-metadata", "io.wabbit-networks.buildId=123").
				MatchKeyWords(SignSuccessfully)

			notation.Exec("verify", artifact.ReferenceWithTag()).
				MatchKeyWords(
					VerifySuccessfully,
					"KEY",
					"VALUE",
					"io.wabbit-networks.buildId",
					"123",
				)

			notation.Exec("verify", artifact.ReferenceWithDigest(), "--user-metadata", "io.wabbit-networks.buildId=123").
				MatchKeyWords(
					VerifySuccessfully,
					"KEY",
					"VALUE",
					"io.wabbit-networks.buildId",
					"123",
				)

			notation.Exec("verify", artifact.ReferenceWithDigest(), "--user-metadata", "io.wabbit-networks.buildId=321").
				MatchErrKeyWords("unable to find specified metadata in the signature")
		})
	})

	It("with json output", func() {
		Host(BaseOptions(), func(notation *utils.ExecOpts, artifact *Artifact, vhost *utils.VirtualHost) {
			notation.Exec("sign", artifact.ReferenceWithDigest(), "--user-metadata", "io.wabbit-networks.buildId=123").
				MatchKeyWords(SignSuccessfully)

			notation.Exec("verify", artifact.ReferenceWithDigest(), "--output", "json").
				MatchContent(fmt.Sprintf("{\n    \"reference\": \"%s\",\n    \"userMetadata\": {\n        \"io.wabbit-networks.buildId\": \"123\"\n    },\n    \"result\": \"Success\"\n}\n", artifact.ReferenceWithDigest()))

			notation.Exec("verify", artifact.ReferenceWithDigest(), "--user-metadata", "io.wabbit-networks.buildId=321").
				MatchErrKeyWords("unable to find specified metadata in the signature")
		})
	})
})
