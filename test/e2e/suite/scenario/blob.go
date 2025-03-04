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
	"os"
	"path/filepath"

	. "github.com/notaryproject/notation/test/e2e/internal/notation"
	"github.com/notaryproject/notation/test/e2e/internal/utils"
	. "github.com/notaryproject/notation/test/e2e/suite/common"
	. "github.com/onsi/ginkgo/v2"
)

var _ = Describe("notation blob", Serial, func() {
	It("signing and verifying with policy init command", func() {
		Host(Opts(), func(notation *utils.ExecOpts, _ *Artifact, vhost *utils.VirtualHost) {
			workDir := vhost.AbsolutePath()

			// create a file to be signed
			content := "hello, world"
			blobPath := filepath.Join(workDir, "hello.txt")
			if err := os.WriteFile(blobPath, []byte(content), 0644); err != nil {
				Fail(err.Error())
			}

			// generate a testing key pair
			notation.Exec("cert", "generate-test", "--default", "testcert").
				MatchKeyWords(
					"Successfully added testcert.crt to named store testcert of type ca",
					"testcert: added to the key list",
				)

			// sign the file
			notation.WithWorkDir(workDir).Exec("blob", "sign", blobPath).
				MatchKeyWords(SignSuccessfully)

			// policy init
			notation.Exec("blob", "policy", "init",
				"--name", "testpolicy",
				"--trust-store", "ca:testcert",
				"--trusted-identity", "x509.subject: CN=testcert,O=Notary,L=Seattle,ST=WA,C=US").
				MatchKeyWords(
					"Successfully initialized blob trust policy file to",
				)

			notation.Exec("blob", "policy", "show").
				MatchContent(`{
  "version": "1.0",
  "trustPolicies": [
    {
      "name": "testpolicy",
      "signatureVerification": {
        "level": "strict"
      },
      "trustStores": [
        "ca:testcert"
      ],
      "trustedIdentities": [
        "x509.subject: CN=testcert,O=Notary,L=Seattle,ST=WA,C=US"
      ]
    }
  ]
}`)

			// verify the blob signature hello.txt.jws.sig
			sigPath := blobPath + ".jws.sig"
			notation.Exec("blob", "verify",
				"--signature", sigPath,
				"--policy-name", "testpolicy",
				blobPath).
				MatchKeyWords(VerifySuccessfully)
		})
	})
})
