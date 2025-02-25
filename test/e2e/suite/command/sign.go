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
	"time"

	. "github.com/notaryproject/notation/test/e2e/internal/notation"
	"github.com/notaryproject/notation/test/e2e/internal/utils"
	. "github.com/notaryproject/notation/test/e2e/suite/common"
	. "github.com/onsi/ginkgo/v2"
)

var _ = Describe("notation sign", func() {
	It("by digest", func() {
		Host(BaseOptions(), func(notation *utils.ExecOpts, artifact *Artifact, vhost *utils.VirtualHost) {
			notation.Exec("sign", artifact.ReferenceWithDigest()).
				MatchKeyWords(SignSuccessfully)

			notation.WithDescription("verify by digest").
				Exec("verify", artifact.ReferenceWithDigest()).
				MatchKeyWords(VerifySuccessfully)

			notation.WithDescription("verify by tag").
				Exec("verify", artifact.ReferenceWithTag()).
				MatchKeyWords(VerifySuccessfully)
		})
	})

	It("by digest with COSE format", func() {
		Host(BaseOptions(), func(notation *utils.ExecOpts, artifact *Artifact, vhost *utils.VirtualHost) {
			notation.Exec("sign", "--signature-format", "cose", artifact.ReferenceWithDigest()).
				MatchKeyWords(SignSuccessfully)

			notation.WithDescription("verify by digest").
				Exec("verify", artifact.ReferenceWithTag()).
				MatchKeyWords(VerifySuccessfully)

			notation.WithDescription("verify by tag").
				Exec("verify", artifact.ReferenceWithTag()).
				MatchKeyWords(VerifySuccessfully)
		})
	})

	It("by tag with JWS format", func() {
		Host(BaseOptions(), func(notation *utils.ExecOpts, artifact *Artifact, vhost *utils.VirtualHost) {
			notation.WithDescription("sign with JWS").
				Exec("sign", artifact.ReferenceWithTag(), "--signature-format", "jws").
				MatchKeyWords(SignSuccessfully)

			notation.WithDescription("verify JWS signature").
				Exec("verify", artifact.ReferenceWithTag()).
				MatchKeyWords(VerifySuccessfully)
		})
	})

	It("by tag with COSE signature format", func() {
		Host(BaseOptions(), func(notation *utils.ExecOpts, artifact *Artifact, vhost *utils.VirtualHost) {
			notation.WithDescription("sign with COSE").
				Exec("sign", artifact.ReferenceWithTag(), "--signature-format", "cose").
				MatchKeyWords(SignSuccessfully)

			notation.WithDescription("verify COSE signature").
				Exec("verify", artifact.ReferenceWithTag()).
				MatchKeyWords(VerifySuccessfully)
		})
	})

	It("with force-referrers-tag set to true", func() {
		Host(BaseOptions(), func(notation *utils.ExecOpts, artifact *Artifact, vhost *utils.VirtualHost) {
			notation.WithDescription("store signature with referrers tag schema").
				Exec("sign", artifact.ReferenceWithDigest(), "--force-referrers-tag=true").
				MatchKeyWords(SignSuccessfully)

			notation.WithDescription("verify by tag schema").
				Exec("verify", artifact.ReferenceWithDigest(), "-v").
				MatchKeyWords(VerifySuccessfully)
		})
	})

	It("with force-referrers-tag set to false", func() {
		Host(BaseOptionsWithExperimental(), func(notation *utils.ExecOpts, artifact *Artifact, vhost *utils.VirtualHost) {
			notation.WithDescription("store signature with Referrers API").
				Exec("sign", artifact.ReferenceWithDigest(), "--force-referrers-tag=false").
				MatchKeyWords(SignSuccessfully)

			notation.WithDescription("verify by referrers api").
				Exec("verify", artifact.ReferenceWithDigest(), "-v").
				MatchKeyWords(VerifySuccessfully)
		})
	})

	It("with specific key", func() {
		Host(BaseOptions(), func(notation *utils.ExecOpts, artifact *Artifact, vhost *utils.VirtualHost) {
			const keyName = "sKey"
			notation.Exec("cert", "generate-test", keyName).
				MatchKeyWords(fmt.Sprintf("notation/localkeys/%s.crt", keyName))

			notation.Exec("sign", "--key", keyName, artifact.ReferenceWithDigest()).
				MatchKeyWords(SignSuccessfully)
		})
	})

	It("with expiry in 24h", func() {
		Host(BaseOptions(), func(notation *utils.ExecOpts, artifact *Artifact, vhost *utils.VirtualHost) {
			notation.Exec("sign", "--expiry", "24h", artifact.ReferenceWithDigest()).
				MatchKeyWords(SignSuccessfully)

			notation.Exec("verify", artifact.ReferenceWithTag()).
				MatchKeyWords(VerifySuccessfully)
		})
	})

	It("with expiry in 2s", func() {
		Host(BaseOptions(), func(notation *utils.ExecOpts, artifact *Artifact, vhost *utils.VirtualHost) {
			notation.Exec("sign", "--expiry", "2s", artifact.ReferenceWithDigest()).
				MatchKeyWords(SignSuccessfully)

			// sleep to wait for expiry
			time.Sleep(2100 * time.Millisecond)

			notation.ExpectFailure().Exec("verify", artifact.ReferenceWithDigest(), "-v").
				MatchErrKeyWords("expiry validation failed.").
				MatchErrKeyWords("signature verification failed for all the signatures")
		})
	})

	It("by digest with oci layout", func() {
		HostWithOCILayout(BaseOptionsWithExperimental(), func(notation *utils.ExecOpts, ociLayout *OCILayout, vhost *utils.VirtualHost) {
			notation.Exec("sign", "--oci-layout", ociLayout.ReferenceWithDigest()).
				MatchKeyWords(SignSuccessfully)
		})
	})

	It("by digest with oci layout and COSE format", func() {
		HostWithOCILayout(BaseOptionsWithExperimental(), func(notation *utils.ExecOpts, ociLayout *OCILayout, vhost *utils.VirtualHost) {
			notation.Exec("sign", "--oci-layout", "--signature-format", "cose", ociLayout.ReferenceWithDigest()).
				MatchKeyWords(SignSuccessfully)
		})
	})

	It("by tag with oci layout", func() {
		HostWithOCILayout(BaseOptionsWithExperimental(), func(notation *utils.ExecOpts, ociLayout *OCILayout, vhost *utils.VirtualHost) {
			notation.Exec("sign", "--oci-layout", ociLayout.ReferenceWithDigest()).
				MatchKeyWords(SignSuccessfully)
		})
	})

	It("by tag with oci layout and COSE format", func() {
		HostWithOCILayout(BaseOptionsWithExperimental(), func(notation *utils.ExecOpts, ociLayout *OCILayout, vhost *utils.VirtualHost) {
			notation.Exec("sign", "--oci-layout", "--signature-format", "cose", ociLayout.ReferenceWithDigest()).
				MatchKeyWords(SignSuccessfully)
		})
	})

	It("by digest with oci layout but without experimental", func() {
		HostWithOCILayout(BaseOptions(), func(notation *utils.ExecOpts, ociLayout *OCILayout, vhost *utils.VirtualHost) {
			expectedErrMsg := "Error: flag(s) --oci-layout in \"notation sign\" is experimental and not enabled by default. To use, please set NOTATION_EXPERIMENTAL=1 environment variable\n"
			notation.ExpectFailure().Exec("sign", "--oci-layout", ociLayout.ReferenceWithDigest()).
				MatchErrContent(expectedErrMsg)
		})
	})

	It("with TLS by digest", func() {
		HostInGithubAction(BaseOptions(), func(notation *utils.ExecOpts, artifact *Artifact, vhost *utils.VirtualHost) {
			notation.Exec("sign", "-d", artifact.DomainReferenceWithDigest()).
				MatchKeyWords(SignSuccessfully).
				MatchErrKeyWords(HTTPSRequest).
				NoMatchErrKeyWords(HTTPRequest)

			notation.Exec("verify", artifact.DomainReferenceWithDigest()).
				MatchKeyWords(VerifySuccessfully)
		})
	})

	It("with --insecure-registry by digest", func() {
		HostInGithubAction(BaseOptions(), func(notation *utils.ExecOpts, artifact *Artifact, vhost *utils.VirtualHost) {
			notation.Exec("sign", "-d", "--insecure-registry", artifact.DomainReferenceWithDigest()).
				MatchKeyWords(SignSuccessfully).
				MatchErrKeyWords(HTTPRequest).
				NoMatchErrKeyWords(HTTPSRequest)

			notation.Exec("verify", artifact.DomainReferenceWithDigest()).
				MatchKeyWords(VerifySuccessfully)
		})
	})

	It("with timestamping", func() {
		Host(BaseOptions(), func(notation *utils.ExecOpts, artifact *Artifact, vhost *utils.VirtualHost) {
			notation.Exec("sign", "--timestamp-url", "http://rfc3161timestamp.globalsign.com/advanced", "--timestamp-root-cert", filepath.Join(NotationE2EConfigPath, "timestamp", "globalsignTSARoot.cer"), artifact.ReferenceWithDigest()).
				MatchKeyWords(SignSuccessfully)
		})
	})

	It("with timestamp-root-cert but no timestamp-url", func() {
		Host(BaseOptions(), func(notation *utils.ExecOpts, artifact *Artifact, vhost *utils.VirtualHost) {
			notation.ExpectFailure().Exec("sign", "--timestamp-root-cert", filepath.Join(NotationE2EConfigPath, "timestamp", "globalsignTSARoot.cer"), artifact.ReferenceWithDigest()).
				MatchErrKeyWords("Error: if any flags in the group [timestamp-url timestamp-root-cert] are set they must all be set; missing [timestamp-url]")
		})
	})

	It("with timestamp-url but no timestamp-root-cert", func() {
		Host(BaseOptions(), func(notation *utils.ExecOpts, artifact *Artifact, vhost *utils.VirtualHost) {
			notation.ExpectFailure().Exec("sign", "--timestamp-url", "http://rfc3161timestamp.globalsign.com/advanced", artifact.ReferenceWithDigest()).
				MatchErrKeyWords("Error: if any flags in the group [timestamp-url timestamp-root-cert] are set they must all be set; missing [timestamp-root-cert]")
		})
	})

	It("with timestamping and empty tsa server", func() {
		Host(BaseOptions(), func(notation *utils.ExecOpts, artifact *Artifact, vhost *utils.VirtualHost) {
			notation.ExpectFailure().Exec("sign", "--timestamp-url", "", "--timestamp-root-cert", filepath.Join(NotationE2EConfigPath, "timestamp", "globalsignTSARoot.cer"), artifact.ReferenceWithDigest()).
				MatchErrKeyWords("Error: timestamping: tsa url cannot be empty")
		})
	})

	It("with timestamping and empty tsa root cert", func() {
		Host(BaseOptions(), func(notation *utils.ExecOpts, artifact *Artifact, vhost *utils.VirtualHost) {
			notation.ExpectFailure().Exec("sign", "--timestamp-url", "dummy", "--timestamp-root-cert", "", artifact.ReferenceWithDigest()).
				MatchErrKeyWords("Error: timestamping: tsa root certificate path cannot be empty")
		})
	})

	It("with timestamping and invalid tsa server", func() {
		Host(BaseOptions(), func(notation *utils.ExecOpts, artifact *Artifact, vhost *utils.VirtualHost) {
			notation.ExpectFailure().Exec("sign", "--timestamp-url", "http://tsa.invalid", "--timestamp-root-cert", filepath.Join(NotationE2EConfigPath, "timestamp", "globalsignTSARoot.cer"), artifact.ReferenceWithDigest()).
				MatchErrKeyWords("Error: timestamp: Post \"http://tsa.invalid\"")
		})
	})

	It("with timestamping and invalid tsa root certificate", func() {
		Host(BaseOptions(), func(notation *utils.ExecOpts, artifact *Artifact, vhost *utils.VirtualHost) {
			notation.ExpectFailure().Exec("sign", "--timestamp-url", "http://timestamp.digicert.com", "--timestamp-root-cert", filepath.Join(NotationE2EConfigPath, "timestamp", "invalid.crt"), artifact.ReferenceWithDigest()).
				MatchErrKeyWords("Error: x509: malformed certificate")
		})
	})

	It("with timestamping and empty tsa root certificate file", func() {
		Host(BaseOptions(), func(notation *utils.ExecOpts, artifact *Artifact, vhost *utils.VirtualHost) {
			notation.ExpectFailure().Exec("sign", "--timestamp-url", "http://timestamp.digicert.com", "--timestamp-root-cert", filepath.Join(NotationE2EConfigPath, "timestamp", "Empty.txt"), artifact.ReferenceWithDigest()).
				MatchErrKeyWords("cannot find any certificate from").
				MatchErrKeyWords("Expecting single x509 root certificate in PEM or DER format from the file")
		})
	})

	It("with timestamping and more than one certificates in tsa root certificate file", func() {
		Host(BaseOptions(), func(notation *utils.ExecOpts, artifact *Artifact, vhost *utils.VirtualHost) {
			notation.ExpectFailure().Exec("sign", "--timestamp-url", "http://timestamp.digicert.com", "--timestamp-root-cert", filepath.Join(NotationE2EConfigPath, "timestamp", "CertChain.pem"), artifact.ReferenceWithDigest()).
				MatchErrKeyWords("found more than one certificates").
				MatchErrKeyWords("Expecting single x509 root certificate in PEM or DER format from the file")
		})
	})

	It("with timestamping and intermediate certificate file", func() {
		Host(BaseOptions(), func(notation *utils.ExecOpts, artifact *Artifact, vhost *utils.VirtualHost) {
			notation.ExpectFailure().Exec("sign", "--timestamp-url", "http://timestamp.digicert.com", "--timestamp-root-cert", filepath.Join(NotationE2EConfigPath, "timestamp", "intermediate.pem"), artifact.ReferenceWithDigest()).
				MatchErrKeyWords("failed to check root certificate with error: crypto/rsa: verification error")
		})
	})

	It("with timestamping and not self-issued certificate file", func() {
		Host(BaseOptions(), func(notation *utils.ExecOpts, artifact *Artifact, vhost *utils.VirtualHost) {
			notation.ExpectFailure().Exec("sign", "--timestamp-url", "http://timestamp.digicert.com", "--timestamp-root-cert", filepath.Join(NotationE2EConfigPath, "timestamp", "notSelfIssued.crt"), artifact.ReferenceWithDigest()).
				MatchErrKeyWords("is not a root certificate. Expecting single x509 root certificate in PEM or DER format from the file")
		})
	})
})
