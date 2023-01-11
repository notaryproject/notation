package scenario_test

import (
	"fmt"

	. "github.com/notaryproject/notation/test/e2e/internal/notation"
	"github.com/notaryproject/notation/test/e2e/internal/utils"
	"github.com/notaryproject/notation/test/e2e/internal/utils/validator"
	. "github.com/onsi/ginkgo/v2"
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
		artifact = GenerateArtifact("", "")
	})

	It("list the signatures associated with the container image", func() {
		notation.Exec("ls", artifact.ReferenceWithTag()).
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

		notation.Exec("cert", "ls").
			MatchKeyWords("notation/truststore/x509/ca/wabbit-networks.io/wabbit-networks.io.crt")
	})

	It("sign the container image", func() {
		notation.Exec("sign", artifact.ReferenceWithDigest()).
			MatchContent(fmt.Sprintf("Successfully signed %s\n", artifact.ReferenceWithDigest()))

		notation.Exec("ls", artifact.ReferenceWithDigest()).
			MatchKeyWords(fmt.Sprintf("%s\n└── application/vnd.cncf.notary.signature\n    └── sha256:", artifact.ReferenceWithDigest()))
	})

	It("Create a trust policy", func() {
		vhost.SetOption(AddTrustPolicyOption("quickstart_trustpolicy.json"))
		validator.CheckFileExist(vhost.AbsolutePath(NotationDirName, TrustPolicyName))
	})

	It("Verify the container image", func() {
		notation.Exec("verify", artifact.ReferenceWithDigest()).
			MatchContent(fmt.Sprintf("Successfully verified signature for %s\n", artifact.ReferenceWithDigest()))
	})
})
