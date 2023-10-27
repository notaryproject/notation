// Copyright The Notary Project Authors.
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

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
	var artifact2 *Artifact
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
		artifact2 = GenerateArtifact("", "")
	})

	It("list the signatures associated with the container image", func() {
		notation.Exec("ls", artifact.ReferenceWithTag()).
			MatchKeyWords("has no associated signature")
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
			MatchKeyWords(
				"ca",
				"wabbit-networks.io",
				"wabbit-networks.io.crt",
			)
	})

	It("sign the container image with jws format (by default)", func() {
		notation.Exec("sign", artifact.ReferenceWithDigest()).
			MatchContent(fmt.Sprintf("Successfully signed %s\n", artifact.ReferenceWithDigest()))

		notation.Exec("ls", artifact.ReferenceWithDigest()).
			MatchKeyWords(fmt.Sprintf("%s\n└── application/vnd.cncf.notary.signature\n    └── sha256:", artifact.ReferenceWithDigest()))
	})
	It("sign the container image with cose format", func() {
		notation.Exec("sign", "--signature-format", "cose", artifact2.ReferenceWithDigest()).
			MatchContent(fmt.Sprintf("Successfully signed %s\n", artifact2.ReferenceWithDigest()))

		notation.Exec("ls", artifact2.ReferenceWithDigest()).
			MatchKeyWords(fmt.Sprintf("%s\n└── application/vnd.cncf.notary.signature\n    └── sha256:", artifact2.ReferenceWithDigest()))
	})

	It("Create a trust policy", func() {
		vhost.SetOption(AddTrustPolicyOption("quickstart_trustpolicy.json"))
		validator.CheckFileExist(vhost.AbsolutePath(NotationDirName, TrustPolicyName))
	})

	It("Verify the container image with jws format", func() {
		notation.Exec("verify", artifact.ReferenceWithDigest()).
			MatchKeyWords(fmt.Sprintf("Successfully verified signature for %s\n", artifact.ReferenceWithDigest()))
	})

	It("Verify the container image with cose format", func() {
		notation.Exec("verify", artifact2.ReferenceWithDigest()).
			MatchContent(fmt.Sprintf("Successfully verified signature for %s\n", artifact2.ReferenceWithDigest()))
	})
})
