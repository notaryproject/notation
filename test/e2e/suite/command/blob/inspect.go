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
	"path/filepath"

	. "github.com/notaryproject/notation/test/e2e/internal/notation"
	"github.com/notaryproject/notation/test/e2e/internal/utils"
	. "github.com/onsi/ginkgo/v2"
)

const (
	jwsBlobSig = "LICENSE.jws.sig"
)

var _ = Describe("notation blob inspect", func() {
	It("with timestamping", func() {
		Host(BaseOptions(), func(notation *utils.ExecOpts, artifact *Artifact, vhost *utils.VirtualHost) {
			expectedKeyWords := `├── signature algorithm: RSASSA-PSS-SHA-256
├── signature envelope type: application/jose+json
├── signed attributes
│   ├── signingScheme: notary.x509
│   └── signingTime: Tue Dec 31 08:05:29 2024
├── user defined attributes
│   └── (empty)
├── unsigned attributes
│   ├── timestamp signature
│   │   ├── timestamp: [Tue Dec 31 08:05:29 2024, Tue Dec 31 08:05:30 2024]
│   │   └── certificates
│   │       ├── SHA256 fingerprint: 36e731cfa9bfd69dafb643809f6dec500902f7197daeaad86ea0159a2268a2b8
│   │       │   ├── issued to: CN=Microsoft Public RSA Timestamping CA 2020,O=Microsoft Corporation,C=US
│   │       │   ├── issued by: CN=Microsoft Identity Verification Root Certificate Authority 2020,O=Microsoft Corporation,C=US
│   │       │   └── expiry: Mon Nov 19 20:42:31 2035
│   │       └── SHA256 fingerprint: 93db2732c49e2624cf0a5cc03ad04acc0927fcaf5e7afdd4a3e23b6fc196aedc
│   │           ├── issued to: CN=Microsoft Public RSA Time Stamping Authority,OU=Microsoft America Operations+OU=nShield TSS ESN:7800-05E0-D947,O=Microsoft Corporation,L=Redmond,ST=Washington,C=US
│   │           ├── issued by: CN=Microsoft Public RSA Timestamping CA 2020,O=Microsoft Corporation,C=US
│   │           └── expiry: Sat Feb 15 20:36:12 2025
│   └── signingAgent: notation-go/1.3.0+unreleased
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
			notation.Exec("blob", "inspect", filepath.Join(NotationE2EConfigPath, "signatures", jwsBlobSig)).
				MatchKeyWords(expectedKeyWords)
		})
	})

	It("with timestamping and output as json", func() {
		Host(BaseOptions(), func(notation *utils.ExecOpts, artifact *Artifact, vhost *utils.VirtualHost) {
			expectedContent := `{
    "mediaType": "application/jose+json",
    "signatureAlgorithm": "RSASSA-PSS-SHA-256",
    "signedAttributes": {
        "signingScheme": "notary.x509",
        "signingTime": "2024-12-31T08:05:29Z"
    },
    "userDefinedAttributes": null,
    "unsignedAttributes": {
        "signingAgent": "notation-go/1.3.0+unreleased",
        "timestampSignature": {
            "timestamp": "[2024-12-31T08:05:29Z, 2024-12-31T08:05:30Z]",
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
			notation.Exec("blob", "inspect", "--output", "json", filepath.Join(NotationE2EConfigPath, "signatures", jwsBlobSig)).
				MatchContent(expectedContent)
		})
	})

})
