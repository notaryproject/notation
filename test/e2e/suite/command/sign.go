package command

import (
	. "github.com/notaryproject/notation/test/e2e/internal/notation"
	"github.com/notaryproject/notation/test/e2e/internal/utils"
	. "github.com/onsi/ginkgo/v2"
)

var _ = Describe("notation sign", func() {
	It("sign in JWS signature format", func() {
		Host(BaseOptions(), func(notation *utils.ExecOpts, artifact *Artifact, vhost *utils.VirtualHost) {
			notation.
				WithDescription("sign with JWS").
				MatchKeyWords("Successfully signed").
				Exec("sign", artifact.Reference(), "--signature-format", "jws")

			notation.
				WithDescription("verify JWS signature").
				MatchKeyWords("Successfully verified").
				Exec("verify", artifact.Reference())
		})
	})

	It("sign in COSE signature format", func() {
		Host(BaseOptions(), func(notation *utils.ExecOpts, artifact *Artifact, vhost *utils.VirtualHost) {
			notation.
				WithDescription("sign with COSE").
				MatchKeyWords("Successfully signed").
				Exec("sign", artifact.Reference(), "--signature-format", "cose")

			notation.
				WithDescription("verify COSE signature").
				MatchKeyWords("Successfully verified").
				Exec("verify", artifact.Reference())
		})
	})
})
