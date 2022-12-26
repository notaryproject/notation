package scenario_test

import (
	. "github.com/notaryproject/notation/test/e2e/internal/notation"
	"github.com/notaryproject/notation/test/e2e/internal/utils"
	. "github.com/onsi/ginkgo/v2"
	// . "github.com/onsi/gomega"
)

// quickstart doc: https://notaryproject.dev/docs/quickstart/
var _ = Describe("notation quickstart E2E test", Ordered, func() {
	var vhost *utils.VirtualHost
	var artifact *Artifact
	var notation *utils.ExecOpts
	BeforeAll(func() {
		var err error
		// setup host
		vhost, err = utils.NewVirtualHost(NotationBinPath, CreateNotationDirOption())
		if err != nil {
			panic(err)
		}
		vhost.SetOption(AuthOption("", ""))
		notation = vhost.Executor

		// add an image to the OCI-compatible registry
		artifact = GenerateArtifact()
		DeferCleanup(artifact.Remove)
	})

	It("list the signatures associated with the container image", func() {
		notation.
			Exec("ls", artifact.Reference()).
			MatchContent("")
	})

	It("generate a test key and self-signed certificate", func() {
		notation.Exec("cert", "generate-test", "--default", "wabbit-networks.io").
			MatchKeyWords(
				"Successfully added wabbit-networks.io.crt",
				"wabbit-networks.io: added to the key list",
				"wabbit-networks.io: mark as default signing key")
		notation.Exec("key", "ls").
			MatchKeyWords(
				"notation/localkeys/wabbit-networks.io.key",
				"notation/localkeys/wabbit-networks.io.crt",
			)

	})

})
