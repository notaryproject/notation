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

package command

import (
	"fmt"
	"path/filepath"

	. "github.com/notaryproject/notation/test/e2e/internal/notation"
	"github.com/notaryproject/notation/test/e2e/internal/utils"

	. "github.com/onsi/ginkgo/v2"
)

var _ = Describe("notation cert", func() {
	It("show all", func() {
		Host(BaseOptions(), func(notation *utils.ExecOpts, artifact *Artifact, vhost *utils.VirtualHost) {
			notation.Exec("cert", "list").
				MatchKeyWords(
					"STORE TYPE   STORE NAME   CERTIFICATE",
				)
		})
	})

	It("delete all", func() {
		Host(BaseOptions(), func(notation *utils.ExecOpts, artifact *Artifact, vhost *utils.VirtualHost) {
			notation.Exec("cert", "delete", "--all", "--type", "ca", "--store", "e2e", "-y").
				MatchKeyWords(
					"Successfully deleted",
				)
		})
	})

	It("delete a specfic cert", func() {
		Host(BaseOptions(), func(notation *utils.ExecOpts, artifact *Artifact, vhost *utils.VirtualHost) {
			notation.Exec("cert", "delete", "--type", "ca", "--store", "e2e", "e2e.crt", "-y").
				MatchKeyWords(
					"Successfully deleted e2e.crt",
				)
		})
	})

	It("delete a non-exist cert", func() {
		Host(BaseOptions(), func(notation *utils.ExecOpts, artifact *Artifact, vhost *utils.VirtualHost) {
			notation.ExpectFailure().Exec("cert", "delete", "--type", "ca", "--store", "e2e", "non-exist.crt", "-y").
				MatchErrKeyWords(
					"failed to delete the certificate file",
				)
		})
	})

	It("show e2e cert", func() {
		Host(BaseOptions(), func(notation *utils.ExecOpts, artifact *Artifact, vhost *utils.VirtualHost) {
			notation.Exec("cert", "show", "--type", "ca", "--store", "e2e", "e2e.crt").
				MatchKeyWords(
					"Issuer: CN=e2e,O=Notary,L=Seattle,ST=WA,C=US",
				)
		})
	})

	It("list", func() {
		Host(BaseOptions(), func(notation *utils.ExecOpts, artifact *Artifact, vhost *utils.VirtualHost) {
			notation.Exec("cert", "list").
				MatchKeyWords(
					"STORE TYPE   STORE NAME   CERTIFICATE",
				)
		})
	})

	It("list with type", func() {
		Host(BaseOptions(), func(notation *utils.ExecOpts, artifact *Artifact, vhost *utils.VirtualHost) {
			notation.Exec("cert", "list", "--type", "ca").
				MatchKeyWords(
					"STORE TYPE   STORE NAME   CERTIFICATE",
					"e2e.crt",
				)
		})
	})

	It("list with store", func() {
		Host(BaseOptions(), func(notation *utils.ExecOpts, artifact *Artifact, vhost *utils.VirtualHost) {
			notation.Exec("cert", "list", "--store", "e2e").
				MatchKeyWords(
					"STORE TYPE   STORE NAME   CERTIFICATE",
					"e2e.crt",
				)
		})
	})

	It("list with type and store", func() {
		Host(BaseOptions(), func(notation *utils.ExecOpts, artifact *Artifact, vhost *utils.VirtualHost) {
			notation.Exec("cert", "list", "--type", "ca", "--store", "e2e").
				MatchKeyWords(
					"STORE TYPE   STORE NAME   CERTIFICATE",
					"e2e.crt",
				)
		})
	})

	It("cleanup test", func() {
		Host(BaseOptions(), func(notation *utils.ExecOpts, artifact *Artifact, vhost *utils.VirtualHost) {
			notation.Exec("cert", "generate-test", "e2e-test").
				MatchKeyWords(
					"generating RSA Key with 2048 bits",
					"generated certificate expiring on",
					"wrote key:", "e2e-test.key",
					"wrote certificate:", "e2e-test.crt",
					"Successfully added e2e-test.crt to named store e2e-test of type ca",
					"e2e-test: added to the key list",
				)

			localKeyPath := filepath.Join(NotationE2ELocalKeysDir, "e2e-test.key")
			localCertPath := filepath.Join(NotationE2ELocalKeysDir, "e2e-test.crt")
			notation.Exec("cert", "cleanup-test", "e2e-test", "-y").
				MatchKeyWords(
					"Successfully deleted e2e-test.crt from trust store e2e-test of type ca",
					`Successfully removed key e2e-test from signingkeys.json`,
					fmt.Sprintf(`Successfully deleted key file: %s`, localKeyPath),
					fmt.Sprintf(`Successfully deleted certificate file: %s`, localCertPath),
					"Cleanup completed successfully",
				)
		})
	})

	It("cleanup test with key set as default", func() {
		Host(BaseOptions(), func(notation *utils.ExecOpts, artifact *Artifact, vhost *utils.VirtualHost) {
			notation.Exec("cert", "generate-test", "e2e-test", "--default").
				MatchKeyWords(
					"generating RSA Key with 2048 bits",
					"generated certificate expiring on",
					"wrote key:", "e2e-test.key",
					"wrote certificate:", "e2e-test.crt",
					"Successfully added e2e-test.crt to named store e2e-test of type ca",
					"e2e-test: added to the key list",
					"e2e-test: mark as default signing key",
				)

			localKeyPath := filepath.Join(NotationE2ELocalKeysDir, "e2e-test.key")
			localCertPath := filepath.Join(NotationE2ELocalKeysDir, "e2e-test.crt")
			notation.Exec("cert", "cleanup-test", "e2e-test", "-y").
				MatchKeyWords(
					"Successfully deleted e2e-test.crt from trust store e2e-test of type ca",
					`Successfully removed key e2e-test from signingkeys.json`,
					fmt.Sprintf(`Successfully deleted key file: %s`, localKeyPath),
					fmt.Sprintf(`Successfully deleted certificate file: %s`, localCertPath),
					"Cleanup completed successfully",
				)
		})
	})
})
