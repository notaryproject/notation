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

			trustStorePath := vhost.AbsolutePath(NotationDirName, TrustStoreDirName, "x509", TrustStoreTypeCA, "e2e")
			if _, err := os.Stat(trustStorePath); err == nil {
				Fail(fmt.Sprintf("empty trust store directory %s should be deleted", trustStorePath))
			}
		})
	})

	It("delete a specific cert and the empty trust store directory", func() {
		Host(BaseOptions(), func(notation *utils.ExecOpts, artifact *Artifact, vhost *utils.VirtualHost) {
			notation.Exec("cert", "delete", "--type", "ca", "--store", "e2e", "e2e.crt", "-y").
				MatchKeyWords(
					"Successfully deleted e2e.crt from trust store e2e of type ca",
				)

			trustStorePath := vhost.AbsolutePath(NotationDirName, TrustStoreDirName, "x509", TrustStoreTypeCA, "e2e")
			if _, err := os.Stat(trustStorePath); err == nil {
				Fail(fmt.Sprintf("empty trust store directory %s should be deleted", trustStorePath))
			}
		})
	})

	It("delete a specific cert from trust store containing more than one certificates", func() {
		Host(BaseOptions(), func(notation *utils.ExecOpts, artifact *Artifact, vhost *utils.VirtualHost) {
			notation.Exec("cert", "add", "--type", "ca", "--store", "e2e", filepath.Join(NotationE2ELocalKeysDir, "expired_e2e.crt")).
				MatchKeyWords("Successfully added following certificates")

			notation.Exec("cert", "delete", "--type", "ca", "--store", "e2e", "expired_e2e.crt", "-y").
				MatchKeyWords(
					"Successfully deleted expired_e2e.crt from trust store e2e of type ca",
				)

			trustStorePath := vhost.AbsolutePath(NotationDirName, TrustStoreDirName, "x509", TrustStoreTypeCA, "e2e")
			if _, err := os.Stat(trustStorePath); err != nil {
				Fail(fmt.Sprintf("trust store directory %s should still exist", trustStorePath))
			}
		})
	})

	It("delete a non-exist cert", func() {
		Host(BaseOptions(), func(notation *utils.ExecOpts, artifact *Artifact, vhost *utils.VirtualHost) {
			notation.ExpectFailure().Exec("cert", "delete", "--type", "ca", "--store", "e2e", "non-exist.crt", "-y").
				MatchErrKeyWords(
					"failed to delete the certificate file",
				)

			trustStorePath := vhost.AbsolutePath(NotationDirName, TrustStoreDirName, "x509", TrustStoreTypeCA, "e2e")
			if _, err := os.Stat(trustStorePath); err != nil {
				Fail(fmt.Sprintf("trust store directory %s should still exist", trustStorePath))
			}
		})
	})

	It("delete a specific cert but failed to delete the empty trust store directory", func() {
		Host(BaseOptions(), func(notation *utils.ExecOpts, artifact *Artifact, vhost *utils.VirtualHost) {
			trustStorePath := vhost.AbsolutePath(NotationDirName, TrustStoreDirName, "x509", TrustStoreTypeCA, "e2e")

			// Remove read permission for trustStorePath
			if err := os.Chmod(trustStorePath, 0300); err != nil {
				Fail(err.Error())
			}
			defer os.Chmod(trustStorePath, 0700)

			notation.Exec("cert", "delete", "--type", "ca", "--store", "e2e", "e2e.crt", "-y").
				MatchKeyWords(
					"Successfully deleted e2e.crt from trust store e2e of type ca",
				).
				MatchErrContent(
					"failed to remove the empty trust store directory",
				)

			if _, err := os.Stat(trustStorePath); err == nil {
				Fail(fmt.Sprintf("empty trust store directory %s should be deleted", trustStorePath))
			}
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
					"Successfully removed key e2e-test from the key list",
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
					"e2e-test: marked as default signing key",
				)

			notation.Exec("cert", "cleanup-test", "e2e-test", "-y").
				MatchKeyWords(
					"Successfully deleted e2e-test.crt from trust store e2e-test of type ca",
					"Successfully removed key e2e-test from the key list",
					"Successfully deleted key file:", "e2e-test.key",
					"Successfully deleted certificate file:", "e2e-test.crt",
					"Cleanup completed successfully",
				)
		})
	})

	It("cleanup test with same name more than one time", func() {
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
					"Successfully removed key e2e-test from the key list",
					"Successfully deleted key file:", "e2e-test.key",
					"Successfully deleted certificate file:", "e2e-test.crt",
					"Cleanup completed successfully",
				)

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
					"Successfully removed key e2e-test from the key list",
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
					"Key e2e-test does not exist in the key list",
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
					"Successfully removed key e2e-test from the key list",
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
					"Successfully removed key e2e-test from the key list",
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
					"Successfully removed key e2e-test from the key list",
					"Successfully deleted key file:", "e2e-test.key",
					fmt.Sprintf("Certificate file %s does not exist", localCertPath),
					"Cleanup completed successfully",
				)
		})
	})

	It("cleanup test missing certificate common name", func() {
		Host(BaseOptions(), func(notation *utils.ExecOpts, artifact *Artifact, vhost *utils.VirtualHost) {
			notation.ExpectFailure().Exec("cert", "cleanup-test").
				MatchErrKeyWords(
					"missing certificate common name",
				)
		})
	})

	It("cleanup test with empty certificate common name", func() {
		Host(BaseOptions(), func(notation *utils.ExecOpts, artifact *Artifact, vhost *utils.VirtualHost) {
			notation.ExpectFailure().Exec("cert", "cleanup-test", "").
				MatchErrKeyWords(
					"certificate common name must follow [a-zA-Z0-9_.-]+ format",
				)
		})
	})

	It("cleanup test failed at deleting certificate from trust store", func() {
		Host(BaseOptions(), func(notation *utils.ExecOpts, artifact *Artifact, vhost *utils.VirtualHost) {
			notation.Exec("cert", "generate-test", "e2e-test")

			certPath := vhost.AbsolutePath(NotationDirName, TrustStoreDirName, "x509", TrustStoreTypeCA, "e2e-test")
			os.Chmod(certPath, 0000)
			defer os.Chmod(certPath, 0755)

			notation.ExpectFailure().Exec("cert", "cleanup-test", "e2e-test", "-y").
				MatchErrKeyWords(
					"failed to delete certificate e2e-test.crt from trust store e2e-test of type ca",
					"permission denied",
				)
		})
	})

	It("cleanup test failed at removing key from signingkeys.json", func() {
		Host(BaseOptions(), func(notation *utils.ExecOpts, artifact *Artifact, vhost *utils.VirtualHost) {
			notation.Exec("cert", "generate-test", "e2e-test")

			signingKeyPath := vhost.AbsolutePath(NotationDirName, SigningKeysFileName)
			os.Chmod(signingKeyPath, 0000)
			defer os.Chmod(signingKeyPath, 0600)

			notation.ExpectFailure().Exec("cert", "cleanup-test", "e2e-test", "-y").
				MatchErrKeyWords(
					"failed to remove key e2e-test from the key list",
					"permission denied",
				)
		})
	})

	It("cleanup test failed at deleting local key file", func() {
		Host(BaseOptions(), func(notation *utils.ExecOpts, artifact *Artifact, vhost *utils.VirtualHost) {
			notation.Exec("cert", "generate-test", "e2e-test")

			localKeysDir := vhost.AbsolutePath(NotationDirName, LocalKeysDirName)
			os.Chmod(localKeysDir, 0000)
			defer os.Chmod(localKeysDir, 0755)

			notation.ExpectFailure().Exec("cert", "cleanup-test", "e2e-test", "-y").
				MatchErrKeyWords(
					fmt.Sprintf("failed to delete key file %s", filepath.Join(localKeysDir, "e2e-test.key")),
					"permission denied",
				)
		})
	})
})
