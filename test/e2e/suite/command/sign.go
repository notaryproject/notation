package command

import (
	. "github.com/notaryproject/notation/test/e2e/internal/notation"
	"github.com/notaryproject/notation/test/e2e/internal/utils"
	. "github.com/onsi/ginkgo/v2"
)

var _ = Describe("notation sign", func() {
	It("sign in JWS signature format", func() {
		Host(BaseOptions(), func(notation *utils.ExecOpts, artifact *Artifact, vhost *utils.VirtualHost) {
			notation.WithDescription("sign with JWS").
				Exec("sign", artifact.ReferenceWithTag(), "--signature-format", "jws").
				MatchKeyWords("Successfully signed")

			notation.WithDescription("verify JWS signature").
				Exec("verify", artifact.ReferenceWithTag()).
				MatchKeyWords("Successfully verified")
		})
	})

	It("sign in COSE signature format", func() {
		Host(BaseOptions(), func(notation *utils.ExecOpts, artifact *Artifact, vhost *utils.VirtualHost) {
			notation.WithDescription("sign with COSE").
				Exec("sign", artifact.ReferenceWithTag(), "--signature-format", "cose").
				MatchKeyWords("Successfully signed")

			notation.WithDescription("verify COSE signature").
				Exec("verify", artifact.ReferenceWithTag()).
				MatchKeyWords("Successfully verified")
		})
	})
})
