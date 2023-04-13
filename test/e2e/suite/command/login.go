package command

import (
	"fmt"

	. "github.com/notaryproject/notation/test/e2e/internal/notation"
	"github.com/notaryproject/notation/test/e2e/internal/utils"
	. "github.com/notaryproject/notation/test/e2e/suite/common"
	. "github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega/gbytes"
)

var _ = Describe("notation login", func() {
	BeforeEach(func() {
		Skip("The login tests require setting up credential helper running in host and it is not available in Github runner. Issue to remove this skip: https://github.com/notaryproject/notation/issues/587")
	})
	It("should sign an image after successfully logging in the registry by prompt with a correct credential", func() {
		Host(TestLoginOptions(), func(notation *utils.ExecOpts, artifact *Artifact, vhost *utils.VirtualHost) {
			notation.WithInput(gbytes.BufferWithBytes([]byte(fmt.Sprintf("%s\n%s\n", TestRegistry.Username, TestRegistry.Password)))).
				Exec("login", artifact.Host).
				MatchKeyWords(LoginSuccessfully)
			notation.Exec("sign", artifact.ReferenceWithDigest()).
				MatchKeyWords(SignSuccessfully)
			notation.Exec("logout", artifact.Host).
				MatchKeyWords(LogoutSuccessfully)
		})
	})

	It("should fail to sign an image after failing to log in the registry with a wrong credential", func() {
		Host(TestLoginOptions(), func(notation *utils.ExecOpts, artifact *Artifact, vhost *utils.VirtualHost) {
			notation.WithInput(gbytes.BufferWithBytes([]byte(fmt.Sprintf("%s\n%s\n", "invalidUser", "invalidPassword")))).
				ExpectFailure().
				Exec("login", artifact.Host).
				MatchErrKeyWords("unauthorized")
			notation.ExpectFailure().
				Exec("sign", artifact.ReferenceWithDigest()).
				MatchErrKeyWords("credential required for basic auth")
		})
	})
})
