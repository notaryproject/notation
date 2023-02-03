package plugin

import (
	. "github.com/notaryproject/notation/test/e2e/internal/notation"
	"github.com/notaryproject/notation/test/e2e/internal/utils"
	. "github.com/onsi/ginkgo/v2"
)

var _ = Describe("notation plugin list", func() {
	It("with empty result", func() {
		Host(nil, func(notation *utils.ExecOpts, _ *Artifact, vhost *utils.VirtualHost) {
			notation.Exec("plugin", "list").
				MatchContent("NAME   DESCRIPTION   VERSION   CAPABILITIES   ERROR   \n")
		})
	})

	It("with e2e-plugin installed", func() {
		Host(Opts(AddPlugin(NotationE2EPluginPath)), func(notation *utils.ExecOpts, _ *Artifact, vhost *utils.VirtualHost) {
			notation.Exec("plugin", "list").
				MatchKeyWords("NAME", "e2e-plugin").
				MatchKeyWords("DESCRIPTION", "The e2e-plugin is a Notation compatible plugin for Notation E2E test").
				MatchKeyWords("VERSION", "1.0.0").
				MatchKeyWords("CAPABILITIES", "[SIGNATURE_VERIFIER.TRUSTED_IDENTITY SIGNATURE_VERIFIER.REVOCATION_CHECK]").
				MatchKeyWords("ERROR", "<nil>")
		})
	})
})
