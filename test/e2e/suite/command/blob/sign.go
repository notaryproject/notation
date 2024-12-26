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
	. "github.com/notaryproject/notation/test/e2e/internal/notation"
	"github.com/notaryproject/notation/test/e2e/internal/utils"
	. "github.com/notaryproject/notation/test/e2e/suite/common"
	. "github.com/onsi/ginkgo/v2"
)

var _ = Describe("notation blob sign", func() {
	It("blob sign", func() {
		HostWithBlob(BaseOptions(), func(notation *utils.ExecOpts, blobPath string, vhost *utils.VirtualHost) {
			notation.Exec("blob", "sign", blobPath).
				MatchKeyWords(SignSuccessfully)
		})
	})

	It("with COSE format", func() {
		HostWithBlob(BaseOptions(), func(notation *utils.ExecOpts, blobPath string, vhost *utils.VirtualHost) {
			notation.Exec("blob", "sign", "--signature-format", "cose", blobPath).
				MatchKeyWords(SignSuccessfully)
		})
	})

	// It("with specific key", func() {
	// 	Host(BaseOptions(), func(notation *utils.ExecOpts, artifact *Artifact, vhost *utils.VirtualHost) {
	// 		const keyName = "sKey"
	// 		notation.Exec("cert", "generate-test", keyName).
	// 			MatchKeyWords(fmt.Sprintf("notation/localkeys/%s.crt", keyName))

	// 		notation.Exec("sign", "--key", keyName, artifact.ReferenceWithDigest()).
	// 			MatchKeyWords(SignSuccessfully)

	// 		// copy the generated cert file and create the new trust policy for verify signature with generated new key.
	// 		OldNotation(AuthOption("", ""),
	// 			AddTrustStoreOption(keyName, vhost.AbsolutePath(NotationDirName, LocalKeysDirName, keyName+".crt")),
	// 			AddTrustPolicyOption("generate_test_trustpolicy.json"),
	// 		).Exec("verify", artifact.ReferenceWithTag()).
	// 			MatchKeyWords(VerifySuccessfully)
	// 	})
	// })

	// It("with expiry in 24h", func() {
	// 	Host(BaseOptions(), func(notation *utils.ExecOpts, artifact *Artifact, vhost *utils.VirtualHost) {
	// 		notation.Exec("sign", "--expiry", "24h", artifact.ReferenceWithDigest()).
	// 			MatchKeyWords(SignSuccessfully)

	// 		OldNotation().Exec("verify", artifact.ReferenceWithTag()).
	// 			MatchKeyWords(VerifySuccessfully)
	// 	})
	// })

	// It("with expiry in 2s", func() {
	// 	Host(BaseOptions(), func(notation *utils.ExecOpts, artifact *Artifact, vhost *utils.VirtualHost) {
	// 		notation.Exec("sign", "--expiry", "2s", artifact.ReferenceWithDigest()).
	// 			MatchKeyWords(SignSuccessfully)

	// 		// sleep to wait for expiry
	// 		time.Sleep(2100 * time.Millisecond)

	// 		OldNotation().ExpectFailure().Exec("verify", artifact.ReferenceWithDigest(), "-v").
	// 			MatchErrKeyWords("expiry validation failed.").
	// 			MatchErrKeyWords("signature verification failed for all the signatures")
	// 	})
	// })

	// It("with timestamping", func() {
	// 	Host(BaseOptions(), func(notation *utils.ExecOpts, artifact *Artifact, vhost *utils.VirtualHost) {
	// 		notation.Exec("sign", "--timestamp-url", "http://rfc3161timestamp.globalsign.com/advanced", "--timestamp-root-cert", filepath.Join(NotationE2EConfigPath, "timestamp", "globalsignTSARoot.cer"), artifact.ReferenceWithDigest()).
	// 			MatchKeyWords(SignSuccessfully)
	// 	})
	// })

	// It("with timestamp-root-cert but no timestamp-url", func() {
	// 	Host(BaseOptions(), func(notation *utils.ExecOpts, artifact *Artifact, vhost *utils.VirtualHost) {
	// 		notation.ExpectFailure().Exec("sign", "--timestamp-root-cert", filepath.Join(NotationE2EConfigPath, "timestamp", "globalsignTSARoot.cer"), artifact.ReferenceWithDigest()).
	// 			MatchErrKeyWords("Error: if any flags in the group [timestamp-url timestamp-root-cert] are set they must all be set; missing [timestamp-url]")
	// 	})
	// })

	// It("with timestamp-url but no timestamp-root-cert", func() {
	// 	Host(BaseOptions(), func(notation *utils.ExecOpts, artifact *Artifact, vhost *utils.VirtualHost) {
	// 		notation.ExpectFailure().Exec("sign", "--timestamp-url", "http://rfc3161timestamp.globalsign.com/advanced", artifact.ReferenceWithDigest()).
	// 			MatchErrKeyWords("Error: if any flags in the group [timestamp-url timestamp-root-cert] are set they must all be set; missing [timestamp-root-cert]")
	// 	})
	// })

	// It("with timestamping and empty tsa server", func() {
	// 	Host(BaseOptions(), func(notation *utils.ExecOpts, artifact *Artifact, vhost *utils.VirtualHost) {
	// 		notation.ExpectFailure().Exec("sign", "--timestamp-url", "", "--timestamp-root-cert", filepath.Join(NotationE2EConfigPath, "timestamp", "globalsignTSARoot.cer"), artifact.ReferenceWithDigest()).
	// 			MatchErrKeyWords("Error: timestamping: tsa url cannot be empty")
	// 	})
	// })

	// It("with timestamping and empty tsa root cert", func() {
	// 	Host(BaseOptions(), func(notation *utils.ExecOpts, artifact *Artifact, vhost *utils.VirtualHost) {
	// 		notation.ExpectFailure().Exec("sign", "--timestamp-url", "dummy", "--timestamp-root-cert", "", artifact.ReferenceWithDigest()).
	// 			MatchErrKeyWords("Error: timestamping: tsa root certificate path cannot be empty")
	// 	})
	// })

	// It("with timestamping and invalid tsa server", func() {
	// 	Host(BaseOptions(), func(notation *utils.ExecOpts, artifact *Artifact, vhost *utils.VirtualHost) {
	// 		notation.ExpectFailure().Exec("sign", "--timestamp-url", "http://invalid.com", "--timestamp-root-cert", filepath.Join(NotationE2EConfigPath, "timestamp", "globalsignTSARoot.cer"), artifact.ReferenceWithDigest()).
	// 			MatchErrKeyWords("Error: timestamp: Post \"http://invalid.com\"").
	// 			MatchErrKeyWords("server misbehaving")
	// 	})
	// })

	// It("with timestamping and invalid tsa root certificate", func() {
	// 	Host(BaseOptions(), func(notation *utils.ExecOpts, artifact *Artifact, vhost *utils.VirtualHost) {
	// 		notation.ExpectFailure().Exec("sign", "--timestamp-url", "http://timestamp.digicert.com", "--timestamp-root-cert", filepath.Join(NotationE2EConfigPath, "timestamp", "invalid.crt"), artifact.ReferenceWithDigest()).
	// 			MatchErrKeyWords("Error: x509: malformed certificate")
	// 	})
	// })

	// It("with timestamping and empty tsa root certificate file", func() {
	// 	Host(BaseOptions(), func(notation *utils.ExecOpts, artifact *Artifact, vhost *utils.VirtualHost) {
	// 		notation.ExpectFailure().Exec("sign", "--timestamp-url", "http://timestamp.digicert.com", "--timestamp-root-cert", filepath.Join(NotationE2EConfigPath, "timestamp", "Empty.txt"), artifact.ReferenceWithDigest()).
	// 			MatchErrKeyWords("cannot find any certificate from").
	// 			MatchErrKeyWords("Expecting single x509 root certificate in PEM or DER format from the file")
	// 	})
	// })

	// It("with timestamping and more than one certificates in tsa root certificate file", func() {
	// 	Host(BaseOptions(), func(notation *utils.ExecOpts, artifact *Artifact, vhost *utils.VirtualHost) {
	// 		notation.ExpectFailure().Exec("sign", "--timestamp-url", "http://timestamp.digicert.com", "--timestamp-root-cert", filepath.Join(NotationE2EConfigPath, "timestamp", "CertChain.pem"), artifact.ReferenceWithDigest()).
	// 			MatchErrKeyWords("found more than one certificates").
	// 			MatchErrKeyWords("Expecting single x509 root certificate in PEM or DER format from the file")
	// 	})
	// })

	// It("with timestamping and intermediate certificate file", func() {
	// 	Host(BaseOptions(), func(notation *utils.ExecOpts, artifact *Artifact, vhost *utils.VirtualHost) {
	// 		notation.ExpectFailure().Exec("sign", "--timestamp-url", "http://timestamp.digicert.com", "--timestamp-root-cert", filepath.Join(NotationE2EConfigPath, "timestamp", "intermediate.pem"), artifact.ReferenceWithDigest()).
	// 			MatchErrKeyWords("failed to check root certificate with error: crypto/rsa: verification error")
	// 	})
	// })

	// It("with timestamping and not self-issued certificate file", func() {
	// 	Host(BaseOptions(), func(notation *utils.ExecOpts, artifact *Artifact, vhost *utils.VirtualHost) {
	// 		notation.ExpectFailure().Exec("sign", "--timestamp-url", "http://timestamp.digicert.com", "--timestamp-root-cert", filepath.Join(NotationE2EConfigPath, "timestamp", "notSelfIssued.crt"), artifact.ReferenceWithDigest()).
	// 			MatchErrKeyWords("is not a root certificate. Expecting single x509 root certificate in PEM or DER format from the file")
	// 	})
	// })
})
