package command

import (
	. "github.com/notaryproject/notation/test/e2e/internal/notation"
	"github.com/notaryproject/notation/test/e2e/internal/utils"
	. "github.com/onsi/ginkgo/v2"
)

var _ = Describe("notation sign", func() {
	It("with JWS signature format", func() {
		Host(BaseOptions(), func(notation *utils.ExecOpts, artifact *Artifact, vhost *utils.VirtualHost) {
			notation.Exec("sign", artifact.ReferenceWithTag(), "--signature-format", "jws").
				MatchKeyWords("Successfully signed")

			OldNotation().Exec("verify", artifact.ReferenceWithTag()).
				MatchKeyWords("Successfully verified")
		})
	})

	It("with COSE signature format", func() {
		Host(BaseOptions(), func(notation *utils.ExecOpts, artifact *Artifact, vhost *utils.VirtualHost) {
			notation.Exec("sign", artifact.ReferenceWithTag(), "--signature-format", "cose").
				MatchKeyWords("Successfully signed")

			OldNotation().Exec("verify", artifact.ReferenceWithTag()).
				MatchKeyWords("Successfully verified")
		})
	})
})
