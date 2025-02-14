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

package blob

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	. "github.com/notaryproject/notation/test/e2e/internal/notation"
	"github.com/notaryproject/notation/test/e2e/internal/utils"
	. "github.com/notaryproject/notation/test/e2e/suite/common"
	. "github.com/onsi/ginkgo/v2"
)

var _ = Describe("notation blob verify", func() {
	// Success cases
	It("with blob verify", func() {
		HostWithBlob(BaseBlobOptions(), func(notation *utils.ExecOpts, blobPath string, vhost *utils.VirtualHost) {
			workDir := vhost.AbsolutePath()
			notation.WithWorkDir(workDir).Exec("blob", "sign", blobPath).
				MatchKeyWords(SignSuccessfully).
				MatchKeyWords("Signature file written to")

			signaturePath := signatureFilepath(workDir, blobPath, "jws")
			notation.Exec("blob", "verify", "-d", "--signature", signaturePath, blobPath).
				MatchKeyWords(VerifySuccessfully).
				// debug log message outputs to stderr
				MatchErrKeyWords(
					"Verify signature of media type application/jose+json",
					"Name:test-blob-global-statement",
				)
		})
	})

	It("with COSE signature", func() {
		HostWithBlob(BaseBlobOptions(), func(notation *utils.ExecOpts, blobPath string, vhost *utils.VirtualHost) {
			workDir := vhost.AbsolutePath()
			notation.WithWorkDir(workDir).Exec("blob", "--signature-format", "cose", "sign", blobPath).
				MatchKeyWords(SignSuccessfully).
				MatchKeyWords("Signature file written to")

			signaturePath := signatureFilepath(workDir, blobPath, "cose")
			notation.Exec("blob", "verify", "-d", "--signature", signaturePath, blobPath).
				MatchKeyWords(VerifySuccessfully).
				// debug log message outputs to stderr
				MatchErrKeyWords(
					"Verify signature of media type application/cose",
				)
		})
	})

	It("with policy name", func() {
		HostWithBlob(BaseBlobOptions(), func(notation *utils.ExecOpts, blobPath string, vhost *utils.VirtualHost) {
			workDir := vhost.AbsolutePath()
			notation.WithWorkDir(workDir).Exec("blob", "sign", blobPath).
				MatchKeyWords(SignSuccessfully).
				MatchKeyWords("Signature file written to")

			signaturePath := signatureFilepath(workDir, blobPath, "jws")
			notation.Exec("blob", "verify", "-d", "--policy-name", "test-blob-statement", "--signature", signaturePath, blobPath).
				MatchKeyWords(VerifySuccessfully).
				// debug log message outputs to stderr
				MatchErrKeyWords(
					"Name:test-blob-statement",
				)
		})
	})

	It("with media type", func() {
		HostWithBlob(BaseBlobOptions(), func(notation *utils.ExecOpts, blobPath string, vhost *utils.VirtualHost) {
			workDir := vhost.AbsolutePath()
			notation.WithWorkDir(workDir).Exec("blob", "sign", "--media-type", "image/jpeg", blobPath).
				MatchKeyWords(SignSuccessfully).
				MatchKeyWords("Signature file written to")

			signaturePath := signatureFilepath(workDir, blobPath, "jws")
			notation.Exec("blob", "verify", "--media-type", "image/jpeg", "--signature", signaturePath, blobPath).
				MatchKeyWords(VerifySuccessfully)
		})
	})

	It("with timestamping", func() {
		HostWithBlob(BaseBlobOptions(), func(notation *utils.ExecOpts, blobPath string, vhost *utils.VirtualHost) {
			workDir := vhost.AbsolutePath()
			notation.WithWorkDir(workDir).Exec("blob", "sign", "--timestamp-url", tsaURL, "--timestamp-root-cert", filepath.Join(NotationE2EConfigPath, "timestamp", "DigiCertTSARootSHA384.cer"), blobPath, "-d").
				MatchKeyWords(SignSuccessfully).
				MatchKeyWords("Signature file written to")

			signaturePath := signatureFilepath(workDir, blobPath, "jws")
			notation.Exec("blob", "verify", "-d", "--policy-name", "test-blob-with-timestamping", "--signature", signaturePath, blobPath).
				MatchKeyWords(VerifySuccessfully).
				// debug log message outputs to stderr
				MatchErrKeyWords(
					"Timestamp verification: Success",
				)
		})
	})

	It("with user metadata", func() {
		HostWithBlob(BaseBlobOptions(), func(notation *utils.ExecOpts, blobPath string, vhost *utils.VirtualHost) {
			workDir := vhost.AbsolutePath()
			notation.WithWorkDir(workDir).Exec("blob", "sign", "--user-metadata", "k1=v1", blobPath).
				MatchKeyWords(SignSuccessfully).
				MatchKeyWords("Signature file written to")

			signaturePath := signatureFilepath(workDir, blobPath, "jws")
			notation.Exec("blob", "verify", "--user-metadata", "k1=v1", "--signature", signaturePath, blobPath).
				MatchKeyWords(VerifySuccessfully)
		})
	})

	// Failure cases
	It("with missing --signature flag", func() {
		HostWithBlob(BaseBlobOptions(), func(notation *utils.ExecOpts, blobPath string, vhost *utils.VirtualHost) {
			workDir := vhost.AbsolutePath()
			notation.WithWorkDir(workDir).Exec("blob", "sign", blobPath).
				MatchKeyWords(SignSuccessfully).
				MatchKeyWords("Signature file written to")

			notation.ExpectFailure().Exec("blob", "verify", blobPath).
				MatchErrKeyWords("filepath of the signature cannot be empty")
		})
	})

	It("with no permission to read blob", func() {
		HostWithBlob(BaseBlobOptions(), func(notation *utils.ExecOpts, blobPath string, vhost *utils.VirtualHost) {
			workDir := vhost.AbsolutePath()
			noPermissionBlobPath := filepath.Join(workDir, "noPermissionBlob")
			newBlobFile, err := os.Create(noPermissionBlobPath)
			if err != nil {
				Fail(err.Error())
			}
			defer newBlobFile.Close()

			notation.WithWorkDir(workDir).Exec("blob", "sign", noPermissionBlobPath).
				MatchKeyWords(SignSuccessfully).
				MatchKeyWords("Signature file written to")
			if err := os.Chmod(noPermissionBlobPath, 0000); err != nil {
				Fail(err.Error())
			}
			defer os.Chmod(noPermissionBlobPath, 0700)

			signaturePath := signatureFilepath(workDir, noPermissionBlobPath, "jws")
			notation.ExpectFailure().Exec("blob", "verify", "--signature", signaturePath, noPermissionBlobPath).
				MatchErrKeyWords("permission denied")
		})
	})

	It("with no permission to read signature file", func() {
		HostWithBlob(BaseBlobOptions(), func(notation *utils.ExecOpts, blobPath string, vhost *utils.VirtualHost) {
			workDir := vhost.AbsolutePath()
			notation.WithWorkDir(workDir).Exec("blob", "sign", blobPath).
				MatchKeyWords(SignSuccessfully).
				MatchKeyWords("Signature file written to")
			noPermissionSignaturePath := signatureFilepath(workDir, blobPath, "jws")
			if err := os.Chmod(noPermissionSignaturePath, 0000); err != nil {
				Fail(err.Error())
			}
			defer os.Chmod(noPermissionSignaturePath, 0700)

			notation.ExpectFailure().Exec("blob", "verify", "--signature", noPermissionSignaturePath, blobPath).
				MatchErrKeyWords("permission denied")
		})
	})

	It("with invalid plugin-config", func() {
		HostWithBlob(BaseBlobOptions(), func(notation *utils.ExecOpts, blobPath string, vhost *utils.VirtualHost) {
			workDir := vhost.AbsolutePath()
			notation.WithWorkDir(workDir).Exec("blob", "sign", blobPath).
				MatchKeyWords(SignSuccessfully).
				MatchKeyWords("Signature file written to")

			signaturePath := signatureFilepath(workDir, blobPath, "jws")
			notation.ExpectFailure().Exec("blob", "verify", "--plugin-config", "invalid", "--signature", signaturePath, blobPath).
				MatchErrKeyWords(`could not parse flag plugin-config: key-value pair requires "=" as separator`)
		})
	})

	It("with invalid user metadata", func() {
		HostWithBlob(BaseBlobOptions(), func(notation *utils.ExecOpts, blobPath string, vhost *utils.VirtualHost) {
			workDir := vhost.AbsolutePath()
			notation.WithWorkDir(workDir).Exec("blob", "sign", blobPath).
				MatchKeyWords(SignSuccessfully).
				MatchKeyWords("Signature file written to")

			signaturePath := signatureFilepath(workDir, blobPath, "jws")
			notation.ExpectFailure().Exec("blob", "verify", "--user-metadata", "invalid", "--signature", signaturePath, blobPath).
				MatchErrKeyWords(`could not parse flag user-metadata: key-value pair requires "=" as separator`)
		})
	})

	It("with invalid signature file extension", func() {
		HostWithBlob(BaseBlobOptions(), func(notation *utils.ExecOpts, blobPath string, vhost *utils.VirtualHost) {
			workDir := vhost.AbsolutePath()
			notation.WithWorkDir(workDir).Exec("blob", "sign", blobPath).
				MatchKeyWords(SignSuccessfully).
				MatchKeyWords("Signature file written to")

			signaturePath := signatureFilepath(workDir, blobPath, "jws")
			invalidSignaturePath := strings.TrimSuffix(signaturePath, ".sig") + "." + "invalid"
			if err := os.Rename(signaturePath, invalidSignaturePath); err != nil {
				Fail(err.Error())
			}
			notation.ExpectFailure().Exec("blob", "verify", "--signature", invalidSignaturePath, blobPath).
				MatchErrKeyWords(`invalid signature filename blobFile.txt.jws.invalid. The file extension must be .sig`)
		})
	})

	It("with invalid signature file name", func() {
		HostWithBlob(BaseBlobOptions(), func(notation *utils.ExecOpts, blobPath string, vhost *utils.VirtualHost) {
			workDir := vhost.AbsolutePath()
			notation.WithWorkDir(workDir).Exec("blob", "sign", blobPath).
				MatchKeyWords(SignSuccessfully).
				MatchKeyWords("Signature file written to")

			signaturePath := signatureFilepath(workDir, blobPath, "jws")
			invalidSignaturePath := strings.TrimSuffix(signaturePath, ".txt.jws.sig") + ".sig"
			if err := os.Rename(signaturePath, invalidSignaturePath); err != nil {
				Fail(err.Error())
			}
			notation.ExpectFailure().Exec("blob", "verify", "--signature", invalidSignaturePath, blobPath).
				MatchErrKeyWords(`invalid signature filename blobFile.sig. A valid signature file name must contain signature format and .sig file extension`)
		})
	})

	It("with invalid signature format", func() {
		HostWithBlob(BaseBlobOptions(), func(notation *utils.ExecOpts, blobPath string, vhost *utils.VirtualHost) {
			workDir := vhost.AbsolutePath()
			notation.WithWorkDir(workDir).Exec("blob", "sign", blobPath).
				MatchKeyWords(SignSuccessfully).
				MatchKeyWords("Signature file written to")

			signaturePath := signatureFilepath(workDir, blobPath, "jws")
			invalidSignaturePath := signatureFilepath(workDir, blobPath, "invalid")
			if err := os.Rename(signaturePath, invalidSignaturePath); err != nil {
				Fail(err.Error())
			}
			notation.ExpectFailure().Exec("blob", "verify", "--signature", invalidSignaturePath, blobPath).
				MatchErrKeyWords(`signature format "invalid" not supported`).
				MatchErrKeyWords(`Supported signature envelope formats are "jws" and "cose"`)
		})
	})

	It("with mismatch media type", func() {
		HostWithBlob(BaseBlobOptions(), func(notation *utils.ExecOpts, blobPath string, vhost *utils.VirtualHost) {
			workDir := vhost.AbsolutePath()
			notation.WithWorkDir(workDir).Exec("blob", "sign", blobPath).
				MatchKeyWords(SignSuccessfully).
				MatchKeyWords("Signature file written to")

			signaturePath := signatureFilepath(workDir, blobPath, "jws")
			notation.ExpectFailure().Exec("blob", "verify", "--media-type", "image/jpeg", "--signature", signaturePath, blobPath).
				MatchErrKeyWords("integrity check failed. signature does not match the given blob")
		})
	})

	It("with no trust policy", func() {
		HostWithBlob(BaseOptions(), func(notation *utils.ExecOpts, blobPath string, vhost *utils.VirtualHost) {
			workDir := vhost.AbsolutePath()
			notation.WithWorkDir(workDir).Exec("blob", "sign", blobPath).
				MatchKeyWords(SignSuccessfully).
				MatchKeyWords("Signature file written to")

			signaturePath := signatureFilepath(workDir, blobPath, "jws")
			notation.ExpectFailure().Exec("blob", "verify", "--signature", signaturePath, blobPath).
				MatchErrKeyWords(`trust policy is not present. To create a trust policy, see: https://notaryproject.dev/docs/quickstart/#create-a-trust-policy`)
		})
	})
})

func signatureFilepath(signatureDirectory, blobPath, signatureFormat string) string {
	blobFilename := filepath.Base(blobPath)
	signatureFilename := fmt.Sprintf("%s.%s.sig", blobFilename, signatureFormat)
	return filepath.Join(signatureDirectory, signatureFilename)
}
