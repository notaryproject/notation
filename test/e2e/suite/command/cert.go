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
	"os"
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
					"Successfully deleted e2e.crt from trust store e2e of type ca",
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

			notation.Exec("cert", "cleanup-test", "e2e-test", "-y").
				MatchKeyWords(
					"Successfully deleted e2e-test.crt from trust store e2e-test of type ca",
					"Successfully removed key e2e-test from signingkeys.json",
					"Successfully deleted key file:", "e2e-test.key",
					"Successfully deleted certificate file:", "e2e-test.crt",
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

			notation.Exec("cert", "cleanup-test", "e2e-test", "-y").
				MatchKeyWords(
					"Successfully deleted e2e-test.crt from trust store e2e-test of type ca",
					"Successfully removed key e2e-test from signingkeys.json",
					"Successfully deleted key file:", "e2e-test.key",
					"Successfully deleted certificate file:", "e2e-test.crt",
					"Cleanup completed successfully",
				)
		})
	})

	It("cleanup test with key never generated", func() {
		Host(BaseOptions(), func(notation *utils.ExecOpts, artifact *Artifact, vhost *utils.VirtualHost) {
			localKeyPath := vhost.AbsolutePath(NotationDirName, LocalKeysDirName, "e2e-test.key")
			localCertPath := vhost.AbsolutePath(NotationDirName, LocalKeysDirName, "e2e-test.crt")
			notation.Exec("cert", "cleanup-test", "e2e-test", "-y").
				MatchKeyWords(
					"Certificate e2e-test.crt does not exist in trust store e2e-test of type ca",
					"Key e2e-test does not exist in signingkeys.json",
					fmt.Sprintf("Key file %s does not exist", localKeyPath),
					fmt.Sprintf("Certificate file %s does not exist", localCertPath),
					"Cleanup completed successfully",
				)
		})
	})

	It("cleanup test with certificate not in trust store", func() {
		Host(BaseOptions(), func(notation *utils.ExecOpts, artifact *Artifact, vhost *utils.VirtualHost) {
			notation.Exec("cert", "generate-test", "e2e-test")

			notation.Exec("cert", "delete", "--type", "ca", "--store", "e2e-test", "e2e-test.crt", "-y").
				MatchKeyWords(
					"Successfully deleted e2e-test.crt from trust store e2e-test of type ca",
				)

			notation.Exec("cert", "cleanup-test", "e2e-test", "-y").
				MatchKeyWords(
					"Certificate e2e-test.crt does not exist in trust store e2e-test of type ca",
					"Successfully removed key e2e-test from signingkeys.json",
					"Successfully deleted key file:", "e2e-test.key",
					"Successfully deleted certificate file:", "e2e-test.crt",
					"Cleanup completed successfully",
				)
		})
	})

	It("cleanup test without local key file", func() {
		Host(BaseOptions(), func(notation *utils.ExecOpts, artifact *Artifact, vhost *utils.VirtualHost) {
			notation.Exec("cert", "generate-test", "e2e-test")

			localKeyPath := vhost.AbsolutePath(NotationDirName, LocalKeysDirName, "e2e-test.key")
			if err := os.Remove(localKeyPath); err != nil {
				Fail(err.Error())
			}
			notation.Exec("cert", "cleanup-test", "e2e-test", "-y").
				MatchKeyWords(
					"Successfully deleted e2e-test.crt from trust store e2e-test of type ca",
					"Successfully removed key e2e-test from signingkeys.json",
					fmt.Sprintf("Key file %s does not exist", localKeyPath),
					"Successfully deleted certificate file:", "e2e-test.crt",
					"Cleanup completed successfully",
				)
		})
	})

	It("cleanup test without local certificate file", func() {
		Host(BaseOptions(), func(notation *utils.ExecOpts, artifact *Artifact, vhost *utils.VirtualHost) {
			notation.Exec("cert", "generate-test", "e2e-test")

			localCertPath := vhost.AbsolutePath(NotationDirName, LocalKeysDirName, "e2e-test.crt")
			if err := os.Remove(localCertPath); err != nil {
				Fail(err.Error())
			}
			notation.Exec("cert", "cleanup-test", "e2e-test", "-y").
				MatchKeyWords(
					"Successfully deleted e2e-test.crt from trust store e2e-test of type ca",
					"Successfully removed key e2e-test from signingkeys.json",
					"Successfully deleted key file:", "e2e-test.key",
					fmt.Sprintf("Certificate file %s does not exist", localCertPath),
					"Cleanup completed successfully",
				)
		})
	})

	It("cleanup test missing key name", func() {
		Host(BaseOptions(), func(notation *utils.ExecOpts, artifact *Artifact, vhost *utils.VirtualHost) {
			notation.ExpectFailure().Exec("cert", "cleanup-test").
				MatchErrKeyWords(
					"missing key name",
				)
		})
	})

	It("cleanup test with empty key name", func() {
		Host(BaseOptions(), func(notation *utils.ExecOpts, artifact *Artifact, vhost *utils.VirtualHost) {
			notation.ExpectFailure().Exec("cert", "cleanup-test", "").
				MatchErrKeyWords(
					"key name must follow [a-zA-Z0-9_.-]+ format",
				)
		})
	})

	// It("cleanup test failed at deleting certificate from trust store", func() {
	// 	Host(BaseOptions(), func(notation *utils.ExecOpts, artifact *Artifact, vhost *utils.VirtualHost) {
	// 		notation.Exec("cert", "generate-test", "e2e-test")

	// 		certPath := vhost.AbsolutePath(NotationDirName, TrustStoreDirName, "x509", TrustStoreTypeCA, "e2e-test", "e2e-test.crt")
	// 		os.Chmod(certPath, 0400)
	// 		notation.ExpectFailure().Exec("cert", "cleanup-test", "e2e-test", "-y").
	// 			MatchErrKeyWords(
	// 				"failed to delete certificate e2e-test.crt from trust store e2e-test of type ca: permission denied",
	// 			)
	// 		os.Chmod(certPath, 0600)
	// 	})
	// })

	It("cleanup test failed at deleting local key file", func() {
		Host(BaseOptions(), func(notation *utils.ExecOpts, artifact *Artifact, vhost *utils.VirtualHost) {
			notation.Exec("cert", "generate-test", "e2e-test")

			localKeysDir := vhost.AbsolutePath(NotationDirName, LocalKeysDirName)
			os.Chmod(localKeysDir, 0000)
			defer os.Chmod(localKeysDir, 0755)

			notation.ExpectFailure().Exec("cert", "cleanup-test", "e2e-test", "-y").
				MatchErrKeyWords(
					fmt.Sprintf("failed to delete key file %s: permission denied", filepath.Join(localKeysDir, "e2e-test.key")),
				)
		})
	})
})
