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
	"os"
	"path/filepath"

	. "github.com/notaryproject/notation/test/e2e/internal/notation"
	"github.com/notaryproject/notation/test/e2e/internal/utils"
	. "github.com/notaryproject/notation/test/e2e/suite/common"
	. "github.com/onsi/ginkgo/v2"
)

const (
	jwsBlobSig  = "blob.jws.sig"
	coseBlobSig = "blob.cose.sig"
)

var (
	testSignatureDir = filepath.Join(NotationE2ETestDataPath, "signatures")
	jwsBlobSigPath   = filepath.Join(testSignatureDir, jwsBlobSig)
	coseBlobSigPath  = filepath.Join(testSignatureDir, coseBlobSig)
)

var _ = Describe("notation blob inspect", func() {
	It("missing required arguments", func() {
		Host(BaseOptions(), func(notation *utils.ExecOpts, _ *Artifact, vhost *utils.VirtualHost) {
			notation.ExpectFailure().Exec("blob", "inspect").
				MatchErrKeyWords("missing signature path")
		})
	})

	It("unknown output format", func() {
		Host(BaseOptions(), func(notation *utils.ExecOpts, _ *Artifact, vhost *utils.VirtualHost) {
			notation.ExpectFailure().Exec("blob", "inspect", "--output", "unknown", filepath.Join(testSignatureDir, jwsBlobSig)).
				MatchErrKeyWords("unrecognized output format")
		})
	})

	It("unknown signature file name", func() {
		Host(BaseOptions(), func(notation *utils.ExecOpts, _ *Artifact, vhost *utils.VirtualHost) {
			notation.ExpectFailure().Exec("blob", "inspect", "unknown.sig").
				MatchErrKeyWords("invalid signature filename")

			notation.ExpectFailure().Exec("blob", "inspect", "hello.unknown.sig").
				MatchErrKeyWords("signature format \"unknown\" not supported")
		})
	})

	It("missing permission to read signature file", func() {
		Host(BaseOptions(), func(notation *utils.ExecOpts, _ *Artifact, vhost *utils.VirtualHost) {
			err := os.Chmod(vhost.AbsolutePath(), 0000)
			if err != nil {
				Fail("failed to change permission of the directory")
			}
			defer os.Chmod(vhost.AbsolutePath(), 0755)

			notation.ExpectFailure().Exec("blob", "inspect", vhost.AbsolutePath("blob.jws.sig")).
				MatchErrKeyWords(
					"failed to read signature file",
					"permission denied",
				)
		})
	})

	It("malformed signature file", func() {
		Host(BaseOptions(), func(notation *utils.ExecOpts, _ *Artifact, vhost *utils.VirtualHost) {
			notation.ExpectFailure().Exec("blob", "inspect", filepath.Join(testSignatureDir, "malformed.jws.sig")).
				MatchErrKeyWords("failed to parse signature")
		})
	})

	It("with timestamping", func() {
		Host(BaseOptions(), func(notation *utils.ExecOpts, _ *Artifact, vhost *utils.VirtualHost) {
			expectedContent := jwsBlobSigPath + `
├── signature algorithm: RSASSA-PSS-SHA-256
├── signature envelope type: application/jose+json
├── signed attributes
│   ├── content type: application/vnd.cncf.notary.payload.v1+json
│   ├── signing scheme: notary.x509
│   └── signing time: Tue Dec 31 08:05:29 2024
├── user defined attributes
│   └── (empty)
├── unsigned attributes
│   ├── signing agent: notation-go/1.3.0+unreleased
│   └── timestamp signature
│       ├── timestamp: [Tue Dec 31 08:05:29 2024, Tue Dec 31 08:05:30 2024]
│       └── certificates
│           ├── SHA256 fingerprint: 36e731cfa9bfd69dafb643809f6dec500902f7197daeaad86ea0159a2268a2b8
│           │   ├── issued to: CN=Microsoft Public RSA Timestamping CA 2020,O=Microsoft Corporation,C=US
│           │   ├── issued by: CN=Microsoft Identity Verification Root Certificate Authority 2020,O=Microsoft Corporation,C=US
│           │   └── expiry: Mon Nov 19 20:42:31 2035
│           └── SHA256 fingerprint: 93db2732c49e2624cf0a5cc03ad04acc0927fcaf5e7afdd4a3e23b6fc196aedc
│               ├── issued to: CN=Microsoft Public RSA Time Stamping Authority,OU=Microsoft America Operations+OU=nShield TSS ESN:7800-05E0-D947,O=Microsoft Corporation,L=Redmond,ST=Washington,C=US
│               ├── issued by: CN=Microsoft Public RSA Timestamping CA 2020,O=Microsoft Corporation,C=US
│               └── expiry: Sat Feb 15 20:36:12 2025
├── certificates
│   └── SHA256 fingerprint: dadee19c843e94b94daae9854d0de7ad93642b6075e2d1523b860b1770b64a03
│       ├── issued to: CN=testcert2,O=Notary,L=Seattle,ST=WA,C=US
│       ├── issued by: CN=testcert2,O=Notary,L=Seattle,ST=WA,C=US
│       └── expiry: Wed Jan  1 08:04:39 2025
└── signed artifact
    ├── media type: application/octet-stream
    ├── digest: sha256:c71d239df91726fc519c6eb72d318ec65820627232b2f796219e87dcf35d0ab4
    └── size: 11357
`
			notation.Exec("blob", "inspect", "--output", "tree", jwsBlobSigPath).
				MatchContent(expectedContent)
		})
	})

	It("with timestamping and output as json", func() {
		Host(BaseOptions(), func(notation *utils.ExecOpts, _ *Artifact, vhost *utils.VirtualHost) {
			expectedContent := `{
  "signatureAlgorithm": "RSASSA-PSS-SHA-256",
  "signatureEnvelopeType": "application/jose+json",
  "signedAttributes": {
    "contentType": "application/vnd.cncf.notary.payload.v1+json",
    "signingScheme": "notary.x509",
    "signingTime": "2024-12-31T08:05:29Z"
  },
  "userDefinedAttributes": null,
  "unsignedAttributes": {
    "signingAgent": "notation-go/1.3.0+unreleased",
    "timestampSignature": {
      "timestamp": "[2024-12-31T08:05:29.509Z, 2024-12-31T08:05:30.509Z]",
      "certificates": [
        {
          "SHA256Fingerprint": "36e731cfa9bfd69dafb643809f6dec500902f7197daeaad86ea0159a2268a2b8",
          "issuedTo": "CN=Microsoft Public RSA Timestamping CA 2020,O=Microsoft Corporation,C=US",
          "issuedBy": "CN=Microsoft Identity Verification Root Certificate Authority 2020,O=Microsoft Corporation,C=US",
          "expiry": "2035-11-19T20:42:31Z"
        },
        {
          "SHA256Fingerprint": "93db2732c49e2624cf0a5cc03ad04acc0927fcaf5e7afdd4a3e23b6fc196aedc",
          "issuedTo": "CN=Microsoft Public RSA Time Stamping Authority,OU=Microsoft America Operations+OU=nShield TSS ESN:7800-05E0-D947,O=Microsoft Corporation,L=Redmond,ST=Washington,C=US",
          "issuedBy": "CN=Microsoft Public RSA Timestamping CA 2020,O=Microsoft Corporation,C=US",
          "expiry": "2025-02-15T20:36:12Z"
        }
      ]
    }
  },
  "certificates": [
    {
      "SHA256Fingerprint": "dadee19c843e94b94daae9854d0de7ad93642b6075e2d1523b860b1770b64a03",
      "issuedTo": "CN=testcert2,O=Notary,L=Seattle,ST=WA,C=US",
      "issuedBy": "CN=testcert2,O=Notary,L=Seattle,ST=WA,C=US",
      "expiry": "2025-01-01T08:04:39Z"
    }
  ],
  "signedArtifact": {
    "mediaType": "application/octet-stream",
    "digest": "sha256:c71d239df91726fc519c6eb72d318ec65820627232b2f796219e87dcf35d0ab4",
    "size": 11357
  }
}
`
			notation.Exec("blob", "inspect", "--output", "json", jwsBlobSigPath).
				MatchContent(expectedContent)
		})
	})

	It("with cose signature", func() {
		Host(BaseOptions(), func(notation *utils.ExecOpts, _ *Artifact, vhost *utils.VirtualHost) {
			expectedContent := coseBlobSigPath + `
├── signature algorithm: RSASSA-PSS-SHA-256
├── signature envelope type: application/cose
├── signed attributes
│   ├── content type: application/vnd.cncf.notary.payload.v1+json
│   ├── signing scheme: notary.x509
│   └── signing time: Tue Jan  7 08:42:43 2025
├── user defined attributes
│   └── (empty)
├── unsigned attributes
│   ├── signing agent: notation-go/1.3.0+unreleased
│   └── timestamp signature
│       ├── timestamp: [Tue Jan  7 08:42:43 2025, Tue Jan  7 08:42:44 2025]
│       └── certificates
│           ├── SHA256 fingerprint: 36e731cfa9bfd69dafb643809f6dec500902f7197daeaad86ea0159a2268a2b8
│           │   ├── issued to: CN=Microsoft Public RSA Timestamping CA 2020,O=Microsoft Corporation,C=US
│           │   ├── issued by: CN=Microsoft Identity Verification Root Certificate Authority 2020,O=Microsoft Corporation,C=US
│           │   └── expiry: Mon Nov 19 20:42:31 2035
│           └── SHA256 fingerprint: 3403d75002d22e2b8c49a8a113957d9eb225c901b946837fd61ff3ce32c51f65
│               ├── issued to: CN=Microsoft Public RSA Time Stamping Authority,OU=Microsoft America Operations+OU=Thales TSS ESN:45D6-96C5-5E63,O=Microsoft Corporation,L=Redmond,ST=Washington,C=US
│               ├── issued by: CN=Microsoft Public RSA Timestamping CA 2020,O=Microsoft Corporation,C=US
│               └── expiry: Sat Feb 15 20:35:56 2025
├── certificates
│   └── SHA256 fingerprint: 3678adce9daa3a82f4f55fd65e0c87c398b3d9bcd5338c06bbf8850df8c6641d
│       ├── issued to: CN=testcert3,O=Notary,L=Seattle,ST=WA,C=US
│       ├── issued by: CN=testcert3,O=Notary,L=Seattle,ST=WA,C=US
│       └── expiry: Wed Jan  8 08:42:24 2025
└── signed artifact
    ├── media type: application/octet-stream
    ├── digest: sha256:c71d239df91726fc519c6eb72d318ec65820627232b2f796219e87dcf35d0ab4
    └── size: 11357
`

			notation.Exec("blob", "inspect", coseBlobSigPath).
				MatchContent(expectedContent)
		})
	})

	It("with cose signature and output as json", func() {
		Host(BaseOptions(), func(notation *utils.ExecOpts, _ *Artifact, vhost *utils.VirtualHost) {
			expectedContent := `{
  "signatureAlgorithm": "RSASSA-PSS-SHA-256",
  "signatureEnvelopeType": "application/cose",
  "signedAttributes": {
    "contentType": "application/vnd.cncf.notary.payload.v1+json",
    "signingScheme": "notary.x509",
    "signingTime": "2025-01-07T08:42:43Z"
  },
  "userDefinedAttributes": null,
  "unsignedAttributes": {
    "signingAgent": "notation-go/1.3.0+unreleased",
    "timestampSignature": {
      "timestamp": "[2025-01-07T08:42:43.582Z, 2025-01-07T08:42:44.582Z]",
      "certificates": [
        {
          "SHA256Fingerprint": "36e731cfa9bfd69dafb643809f6dec500902f7197daeaad86ea0159a2268a2b8",
          "issuedTo": "CN=Microsoft Public RSA Timestamping CA 2020,O=Microsoft Corporation,C=US",
          "issuedBy": "CN=Microsoft Identity Verification Root Certificate Authority 2020,O=Microsoft Corporation,C=US",
          "expiry": "2035-11-19T20:42:31Z"
        },
        {
          "SHA256Fingerprint": "3403d75002d22e2b8c49a8a113957d9eb225c901b946837fd61ff3ce32c51f65",
          "issuedTo": "CN=Microsoft Public RSA Time Stamping Authority,OU=Microsoft America Operations+OU=Thales TSS ESN:45D6-96C5-5E63,O=Microsoft Corporation,L=Redmond,ST=Washington,C=US",
          "issuedBy": "CN=Microsoft Public RSA Timestamping CA 2020,O=Microsoft Corporation,C=US",
          "expiry": "2025-02-15T20:35:56Z"
        }
      ]
    }
  },
  "certificates": [
    {
      "SHA256Fingerprint": "3678adce9daa3a82f4f55fd65e0c87c398b3d9bcd5338c06bbf8850df8c6641d",
      "issuedTo": "CN=testcert3,O=Notary,L=Seattle,ST=WA,C=US",
      "issuedBy": "CN=testcert3,O=Notary,L=Seattle,ST=WA,C=US",
      "expiry": "2025-01-08T08:42:24Z"
    }
  ],
  "signedArtifact": {
    "mediaType": "application/octet-stream",
    "digest": "sha256:c71d239df91726fc519c6eb72d318ec65820627232b2f796219e87dcf35d0ab4",
    "size": 11357
  }
}
`
			notation.Exec("blob", "inspect", "--output", "json", coseBlobSigPath).
				MatchContent(expectedContent)
		})
	})

	It("with blob sign in jws", func() {
		HostWithBlob(BaseOptions(), func(notation *utils.ExecOpts, blobPath string, vhost *utils.VirtualHost) {
			notation.WithWorkDir(vhost.AbsolutePath()).Exec("blob", "sign", blobPath).
				MatchKeyWords(SignSuccessfully).
				MatchKeyWords("Signature file written to")

			sigName := filepath.Base(blobPath) + ".jws.sig"
			notation.Exec("blob", "inspect", vhost.AbsolutePath(sigName)).
				MatchKeyWords(
					"signature algorithm",
					"signature envelope type",
					"signed attributes",
					"user defined attributes",
					"unsigned attributes",
					"certificates",
					"signed artifact")
		})
	})

	It("with blob sign in cose", func() {
		HostWithBlob(BaseOptions(), func(notation *utils.ExecOpts, blobPath string, vhost *utils.VirtualHost) {
			notation.WithWorkDir(vhost.AbsolutePath()).Exec("blob", "sign", "--signature-format", "cose", blobPath).
				MatchKeyWords(SignSuccessfully).
				MatchKeyWords("Signature file written to")

			sigName := filepath.Base(blobPath) + ".cose.sig"
			notation.Exec("blob", "inspect", vhost.AbsolutePath(sigName)).
				MatchKeyWords(
					"signature algorithm",
					"signature envelope type",
					"signed attributes",
					"user defined attributes",
					"unsigned attributes",
					"certificates",
					"signed artifact")
		})
	})
})
