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

	. "github.com/notaryproject/notation/test/e2e/internal/notation"
	"github.com/notaryproject/notation/test/e2e/internal/utils"
	. "github.com/notaryproject/notation/test/e2e/suite/common"
	. "github.com/onsi/ginkgo/v2"
)

const tsaURL = "http://timestamp.digicert.com"

var _ = Describe("notation blob sign", func() {
	// Success cases
	It("with blob sign", func() {
		HostWithBlob(BaseOptions(), func(notation *utils.ExecOpts, blobPath string, vhost *utils.VirtualHost) {
			notation.Exec("blob", "sign", "--force", blobPath).
				MatchKeyWords(SignSuccessfully).
				MatchKeyWords("Signature file written to")
		})
	})

	It("with COSE format", func() {
		HostWithBlob(BaseOptions(), func(notation *utils.ExecOpts, blobPath string, vhost *utils.VirtualHost) {
			notation.Exec("blob", "sign", "--signature-format", "cose", "--force", blobPath).
				MatchKeyWords(SignSuccessfully).
				MatchKeyWords("Signature file written to")
		})
	})

	It("with specified media-type", func() {
		HostWithBlob(BaseOptions(), func(notation *utils.ExecOpts, blobPath string, vhost *utils.VirtualHost) {
			notation.Exec("blob", "sign", "--force", "--media-type", "other-media-type", blobPath).
				MatchKeyWords(SignSuccessfully).
				MatchKeyWords("Signature file written to")
		})
	})

	It("with specific key", func() {
		HostWithBlob(BaseOptions(), func(notation *utils.ExecOpts, blobPath string, vhost *utils.VirtualHost) {
			const keyName = "sKey"
			notation.Exec("cert", "generate-test", keyName).
				MatchKeyWords(fmt.Sprintf("notation/localkeys/%s.crt", keyName))

			notation.Exec("blob", "sign", "--force", "--key", keyName, blobPath).
				MatchKeyWords(SignSuccessfully).
				MatchKeyWords("Signature file written to")
		})
	})

	It("with expiry in 24h", func() {
		HostWithBlob(BaseOptions(), func(notation *utils.ExecOpts, blobPath string, vhost *utils.VirtualHost) {
			notation.Exec("blob", "sign", "--expiry", "24h", "--force", blobPath).
				MatchKeyWords(SignSuccessfully).
				MatchKeyWords("Signature file written to")
		})
	})

	It("with signature directory", func() {
		HostWithBlob(BaseOptions(), func(notation *utils.ExecOpts, blobPath string, vhost *utils.VirtualHost) {
			blobDir := filepath.Dir(blobPath)
			notation.Exec("blob", "sign", "--force", "--signature-directory", blobDir, blobPath).
				MatchKeyWords(SignSuccessfully).
				MatchKeyWords(fmt.Sprintf("Signature file written to %s", filepath.Join(blobDir, "blobFile.jws.sig")))
		})
	})

	It("with user metadata", func() {
		HostWithBlob(BaseOptions(), func(notation *utils.ExecOpts, blobPath string, vhost *utils.VirtualHost) {
			notation.Exec("blob", "sign", "--force", "--user-metadata", "k1=v1", "--user-metadata", "k2=v2", blobPath).
				MatchKeyWords(SignSuccessfully).
				MatchKeyWords("Signature file written to")
		})
	})

	It("with timestamping", func() {
		HostWithBlob(BaseOptions(), func(notation *utils.ExecOpts, blobPath string, vhost *utils.VirtualHost) {
			notation.Exec("blob", "sign", "--force", "--timestamp-url", tsaURL, "--timestamp-root-cert", filepath.Join(NotationE2EConfigPath, "timestamp", "DigiCertTSARootSHA384.cer"), blobPath).
				MatchKeyWords(SignSuccessfully).
				MatchKeyWords("Signature file written to")
		})
	})

	// Failure cases
	It("with undefined signature format", func() {
		HostWithBlob(BaseOptions(), func(notation *utils.ExecOpts, blobPath string, vhost *utils.VirtualHost) {
			notation.ExpectFailure().Exec("blob", "sign", "--signature-format", "invalid", blobPath).
				MatchErrKeyWords(`signature format "invalid" not supported`)
		})
	})

	It("with invalid key", func() {
		HostWithBlob(BaseOptions(), func(notation *utils.ExecOpts, blobPath string, vhost *utils.VirtualHost) {
			notation.ExpectFailure().Exec("blob", "sign", "--key", "invalid", blobPath).
				MatchErrKeyWords("signing key not found")
		})
	})

	It("with invalid plugin-config", func() {
		HostWithBlob(BaseOptions(), func(notation *utils.ExecOpts, blobPath string, vhost *utils.VirtualHost) {
			notation.ExpectFailure().Exec("blob", "sign", "--plugin-config", "invalid", blobPath).
				MatchErrKeyWords(`could not parse flag plugin-config: key-value pair requires "=" as separator`)
		})
	})

	It("with invalid user metadata", func() {
		HostWithBlob(BaseOptions(), func(notation *utils.ExecOpts, blobPath string, vhost *utils.VirtualHost) {
			notation.ExpectFailure().Exec("blob", "sign", "--user-metadata", "invalid", blobPath).
				MatchErrKeyWords(`could not parse flag user-metadata: key-value pair requires "=" as separator`)
		})
	})

	It("with no permission to read the blob file", func() {
		HostWithBlob(BaseOptions(), func(notation *utils.ExecOpts, blobPath string, vhost *utils.VirtualHost) {
			noPermissionBlobPath := filepath.Join(vhost.AbsolutePath(), "noPermissionBlob")
			newBlobFile, err := os.Create(noPermissionBlobPath)
			if err != nil {
				Fail(err.Error())
			}
			defer newBlobFile.Close()

			if err := os.Chmod(noPermissionBlobPath, 0000); err != nil {
				Fail(err.Error())
			}
			defer os.Chmod(noPermissionBlobPath, 0700)

			notation.ExpectFailure().Exec("blob", "sign", noPermissionBlobPath).
				MatchErrKeyWords("permission denied")
		})
	})

	It("with no permission to write the signature file", func() {
		HostWithBlob(BaseOptions(), func(notation *utils.ExecOpts, blobPath string, vhost *utils.VirtualHost) {
			blobDir := filepath.Dir(blobPath)
			sigDir := filepath.Join(blobDir, "signature")
			if err := os.MkdirAll(sigDir, 0000); err != nil {
				Fail(err.Error())
			}
			defer os.Chmod(sigDir, 0700)

			notation.ExpectFailure().Exec("blob", "sign", "--signature-directory", sigDir, blobPath).
				MatchErrKeyWords("permission denied")
		})
	})

	It("with timestamp-root-cert but no timestamp-url", func() {
		HostWithBlob(BaseOptions(), func(notation *utils.ExecOpts, blobPath string, vhost *utils.VirtualHost) {
			notation.ExpectFailure().Exec("blob", "sign", "--timestamp-root-cert", filepath.Join(NotationE2EConfigPath, "timestamp", "DigiCertTSARootSHA384.cer"), blobPath).
				MatchErrKeyWords("Error: if any flags in the group [timestamp-url timestamp-root-cert] are set they must all be set; missing [timestamp-url]")
		})
	})

	It("with timestamp-url but no timestamp-root-cert", func() {
		HostWithBlob(BaseOptions(), func(notation *utils.ExecOpts, blobPath string, vhost *utils.VirtualHost) {
			notation.ExpectFailure().Exec("blob", "sign", "--timestamp-url", tsaURL, blobPath).
				MatchErrKeyWords("Error: if any flags in the group [timestamp-url timestamp-root-cert] are set they must all be set; missing [timestamp-root-cert]")
		})
	})

	It("with timestamping and empty tsa server", func() {
		HostWithBlob(BaseOptions(), func(notation *utils.ExecOpts, blobPath string, vhost *utils.VirtualHost) {
			notation.ExpectFailure().Exec("blob", "sign", "--timestamp-url", "", "--timestamp-root-cert", filepath.Join(NotationE2EConfigPath, "timestamp", "DigiCertTSARootSHA384.cer"), blobPath).
				MatchErrKeyWords("Error: timestamping: tsa url cannot be empty")
		})
	})

	It("with timestamping and empty tsa root cert", func() {
		HostWithBlob(BaseOptions(), func(notation *utils.ExecOpts, blobPath string, vhost *utils.VirtualHost) {
			notation.ExpectFailure().Exec("blob", "sign", "--timestamp-url", "dummy", "--timestamp-root-cert", "", blobPath).
				MatchErrKeyWords("Error: timestamping: tsa root certificate path cannot be empty")
		})
	})

	It("with timestamping and invalid tsa server", func() {
		HostWithBlob(BaseOptions(), func(notation *utils.ExecOpts, blobPath string, vhost *utils.VirtualHost) {
			notation.ExpectFailure().Exec("blob", "sign", "--timestamp-url", "http://tsa.invalid", "--timestamp-root-cert", filepath.Join(NotationE2EConfigPath, "timestamp", "DigiCertTSARootSHA384.cer"), blobPath).
				MatchErrKeyWords("Error: timestamp: Post \"http://tsa.invalid\"")
		})
	})

	It("with timestamping and invalid tsa root certificate", func() {
		HostWithBlob(BaseOptions(), func(notation *utils.ExecOpts, blobPath string, vhost *utils.VirtualHost) {
			notation.ExpectFailure().Exec("blob", "sign", "--timestamp-url", tsaURL, "--timestamp-root-cert", filepath.Join(NotationE2EConfigPath, "timestamp", "invalid.crt"), blobPath).
				MatchErrKeyWords("Error: x509: malformed certificate")
		})
	})

	It("with timestamping and empty tsa root certificate file", func() {
		HostWithBlob(BaseOptions(), func(notation *utils.ExecOpts, blobPath string, vhost *utils.VirtualHost) {
			notation.ExpectFailure().Exec("blob", "sign", "--timestamp-url", tsaURL, "--timestamp-root-cert", filepath.Join(NotationE2EConfigPath, "timestamp", "Empty.txt"), blobPath).
				MatchErrKeyWords("cannot find any certificate from").
				MatchErrKeyWords("Expecting single x509 root certificate in PEM or DER format from the file")
		})
	})

	It("with timestamping and more than one certificates in tsa root certificate file", func() {
		HostWithBlob(BaseOptions(), func(notation *utils.ExecOpts, blobPath string, vhost *utils.VirtualHost) {
			notation.ExpectFailure().Exec("blob", "sign", "--timestamp-url", tsaURL, "--timestamp-root-cert", filepath.Join(NotationE2EConfigPath, "timestamp", "CertChain.pem"), blobPath).
				MatchErrKeyWords("found more than one certificates").
				MatchErrKeyWords("Expecting single x509 root certificate in PEM or DER format from the file")
		})
	})

	It("with timestamping and intermediate certificate file", func() {
		HostWithBlob(BaseOptions(), func(notation *utils.ExecOpts, blobPath string, vhost *utils.VirtualHost) {
			notation.ExpectFailure().Exec("blob", "sign", "--timestamp-url", tsaURL, "--timestamp-root-cert", filepath.Join(NotationE2EConfigPath, "timestamp", "intermediate.pem"), blobPath).
				MatchErrKeyWords("failed to check root certificate with error: crypto/rsa: verification error")
		})
	})

	It("with timestamping and not self-issued certificate file", func() {
		HostWithBlob(BaseOptions(), func(notation *utils.ExecOpts, blobPath string, vhost *utils.VirtualHost) {
			notation.ExpectFailure().Exec("blob", "sign", "--timestamp-url", tsaURL, "--timestamp-root-cert", filepath.Join(NotationE2EConfigPath, "timestamp", "notSelfIssued.crt"), blobPath).
				MatchErrKeyWords("is not a root certificate. Expecting single x509 root certificate in PEM or DER format from the file")
		})
	})
})
