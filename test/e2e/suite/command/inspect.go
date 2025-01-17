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
	"path/filepath"

	. "github.com/notaryproject/notation/test/e2e/internal/notation"
	"github.com/notaryproject/notation/test/e2e/internal/utils"
	. "github.com/notaryproject/notation/test/e2e/suite/common"
	. "github.com/onsi/ginkgo/v2"
)

var (
	inspectSuccessfully = []string{
		"└── application/vnd.cncf.notary.signature",
		"└── sha256:",
		"├── media type:",
		"├── signature algorithm:",
		"├── signed attributes",
		"signingTime:",
		"signingScheme:",
		"├── user defined attributes",
		"│   └── (empty)",
		"├── unsigned attributes",
		"│   └── signingAgent: notation-go/",
		"├── certificates",
		"│   └── SHA256 fingerprint:",
		"issued to:",
		"issued by:",
		"expiry:",
		"└── signed artifact",
		"media type:",
		"digest:",
		"size:",
	}

	inspectSuccessfullyWithTimestamp = []string{
		"└── application/vnd.cncf.notary.signature",
		"└── sha256:",
		"├── media type:",
		"├── signature algorithm:",
		"├── signed attributes",
		"signingTime:",
		"signingScheme:",
		"├── user defined attributes",
		"│   └── (empty)",
		"├── unsigned attributes",
		"signingAgent: notation-go/",
		"timestamp signature",
		"timestamp:",
		"certificates",
		"SHA256 fingerprint:",
		"├── certificates",
		"│   └── SHA256 fingerprint:",
		"issued to:",
		"issued by:",
		"expiry:",
		"└── signed artifact",
		"media type:",
		"digest:",
		"size:",
	}
)

var _ = Describe("notation inspect", func() {
	It("all signatures of an image", func() {
		Host(BaseOptions(), func(notation *utils.ExecOpts, artifact *Artifact, vhost *utils.VirtualHost) {
			notation.Exec("sign", artifact.ReferenceWithDigest()).
				MatchKeyWords(SignSuccessfully)

			notation.Exec("inspect", "-d", artifact.ReferenceWithDigest()).
				MatchKeyWords(inspectSuccessfully...)
		})
	})

	It("all signatures of an image with TLS", func() {
		HostInGithubAction(BaseOptions(), func(notation *utils.ExecOpts, artifact *Artifact, vhost *utils.VirtualHost) {
			notation.Exec("sign", artifact.DomainReferenceWithDigest()).
				MatchKeyWords(SignSuccessfully)

			notation.Exec("inspect", "-d", artifact.DomainReferenceWithDigest()).
				MatchKeyWords(inspectSuccessfully...).
				MatchErrKeyWords(HTTPSRequest).
				NoMatchErrKeyWords(HTTPRequest)
		})
	})

	It("all signatures of an image with --insecure-registry flag", func() {
		HostInGithubAction(BaseOptions(), func(notation *utils.ExecOpts, artifact *Artifact, vhost *utils.VirtualHost) {
			notation.Exec("sign", artifact.DomainReferenceWithDigest()).
				MatchKeyWords(SignSuccessfully)

			notation.Exec("inspect", "-d", "--insecure-registry", artifact.DomainReferenceWithDigest()).
				MatchKeyWords(inspectSuccessfully...).
				MatchErrKeyWords(HTTPRequest).
				NoMatchErrKeyWords(HTTPSRequest)
		})
	})

	It("sign with --force-referrers-tag set", func() {
		Host(BaseOptions(), func(notation *utils.ExecOpts, artifact *Artifact, vhost *utils.VirtualHost) {
			notation.Exec("sign", "--force-referrers-tag", artifact.ReferenceWithDigest()).
				MatchKeyWords(SignSuccessfully)

			notation.Exec("inspect", artifact.ReferenceWithDigest(), "-v").
				MatchKeyWords(inspectSuccessfully...)
		})
	})

	It("sign with --force-referrers-tag set to false", func() {
		Host(BaseOptions(), func(notation *utils.ExecOpts, artifact *Artifact, vhost *utils.VirtualHost) {
			notation.Exec("sign", "--force-referrers-tag=false", artifact.ReferenceWithDigest()).
				MatchKeyWords(SignSuccessfully)

			notation.Exec("inspect", artifact.ReferenceWithDigest(), "-v").
				MatchKeyWords(inspectSuccessfully...)
		})
	})

	It("sign with --allow-referrers-api set", func() {
		Host(BaseOptionsWithExperimental(), func(notation *utils.ExecOpts, artifact *Artifact, vhost *utils.VirtualHost) {
			notation.Exec("sign", "--allow-referrers-api", artifact.ReferenceWithDigest()).
				MatchKeyWords(SignSuccessfully)

			notation.Exec("inspect", artifact.ReferenceWithDigest(), "-v").
				MatchKeyWords(inspectSuccessfully...)

			notation.Exec("inspect", artifact.ReferenceWithDigest(), "--allow-referrers-api", "-v").
				MatchErrKeyWords(
					"Warning: This feature is experimental and may not be fully tested or completed and may be deprecated.",
					"Warning: flag '--allow-referrers-api' is deprecated and will be removed in future versions.",
				).
				MatchKeyWords(inspectSuccessfully...)
		})
	})

	It("sign with --allow-referrers-api set to false", func() {
		Host(BaseOptionsWithExperimental(), func(notation *utils.ExecOpts, artifact *Artifact, vhost *utils.VirtualHost) {
			notation.Exec("sign", "--allow-referrers-api=false", artifact.ReferenceWithDigest()).
				MatchKeyWords(SignSuccessfully)

			notation.Exec("inspect", artifact.ReferenceWithDigest(), "-v").
				MatchKeyWords(inspectSuccessfully...)

			notation.Exec("inspect", artifact.ReferenceWithDigest(), "--allow-referrers-api", "-v").
				MatchErrKeyWords(
					"Warning: This feature is experimental and may not be fully tested or completed and may be deprecated.",
					"Warning: flag '--allow-referrers-api' is deprecated and will be removed in future versions.",
				).
				MatchKeyWords(inspectSuccessfully...)
		})
	})

	It("with timestamping", func() {
		Host(BaseOptions(), func(notation *utils.ExecOpts, artifact *Artifact, vhost *utils.VirtualHost) {
			notation.Exec("sign", "--timestamp-url", "http://rfc3161timestamp.globalsign.com/advanced", "--timestamp-root-cert", filepath.Join(NotationE2EConfigPath, "timestamp", "globalsignTSARoot.cer"), artifact.ReferenceWithDigest()).
				MatchKeyWords(SignSuccessfully)

			notation.Exec("inspect", artifact.ReferenceWithDigest()).
				MatchKeyWords(inspectSuccessfullyWithTimestamp...)
		})
	})

	It("with timestamped oci layout", func() {
		Host(BaseOptions(), func(notation *utils.ExecOpts, _ *Artifact, vhost *utils.VirtualHost) {
			artifact := GenerateArtifact("e2e-with-timestamped-signature", "e2e")
			expectedOutput := `localhost:5000/e2e@sha256:99950868628ed79ebc295e01f8397dcacad35e17fb3b7a9f0fa77881ec3cef1c
└── application/vnd.cncf.notary.signature
    └── sha256:54eab65f9262feac4ea9f31d15b62c870bf359d912aba86622cfc735337ae4fa
        ├── signature algorithm: RSASSA-PSS-SHA-256
        ├── signature envelope type: application/jose+json
        ├── signed attributes
        │   ├── signingScheme: notary.x509
        │   └── signingTime: Fri Jan 17 06:36:19 2025
        ├── user defined attributes
        │   └── (empty)
        ├── unsigned attributes
        │   ├── signingAgent: notation-go/1.3.0+unreleased
        │   └── timestamp signature
        │       ├── timestamp: [Fri Jan 17 06:36:19 2025, Fri Jan 17 06:36:20 2025]
        │       └── certificates
        │           ├── SHA256 fingerprint: 36e731cfa9bfd69dafb643809f6dec500902f7197daeaad86ea0159a2268a2b8
        │           │   ├── issued to: CN=Microsoft Public RSA Timestamping CA 2020,O=Microsoft Corporation,C=US
        │           │   ├── issued by: CN=Microsoft Identity Verification Root Certificate Authority 2020,O=Microsoft Corporation,C=US
        │           │   └── expiry: Mon Nov 19 20:42:31 2035
        │           └── SHA256 fingerprint: 2be4c1670d176be2b0e56081a7b6523487c528a7ea092febbb84ae9db03ceb9a
        │               ├── issued to: CN=Microsoft Public RSA Time Stamping Authority,OU=Microsoft Ireland Operations Limited+OU=Thales TSS ESN:F5D6-96D6-909E,O=Microsoft Corporation,L=Redmond,ST=Washington,C=US
        │               ├── issued by: CN=Microsoft Public RSA Timestamping CA 2020,O=Microsoft Corporation,C=US
        │               └── expiry: Thu Apr 17 17:59:12 2025
        ├── certificates
        │   └── SHA256 fingerprint: 36b3e0e0fee117d33d5664a2f56b147ddbbe8b7ca3ad2ae56498703fd782a56e
        │       ├── issued to: CN=testcert5,O=Notary,L=Seattle,ST=WA,C=US
        │       ├── issued by: CN=testcert5,O=Notary,L=Seattle,ST=WA,C=US
        │       └── expiry: Sat Jan 18 06:34:29 2025
        └── signed artifact
            ├── media type: application/vnd.oci.image.manifest.v1+json
            ├── digest: sha256:99950868628ed79ebc295e01f8397dcacad35e17fb3b7a9f0fa77881ec3cef1c
            └── size: 582
`

			notation.Exec("inspect", artifact.ReferenceWithDigest()).
				MatchContent(expectedOutput)
		})
	})

	It("with timestamped oci layout and output in JSON", func() {
		Host(BaseOptions(), func(notation *utils.ExecOpts, _ *Artifact, vhost *utils.VirtualHost) {
			artifact := GenerateArtifact("e2e-with-timestamped-signature", "e2e")
			expectedOutput := `{
  "mediaType": "application/vnd.oci.image.manifest.v1+json",
  "Signatures": [
    {
      "mediaType": "application/jose+json",
      "digest": "sha256:54eab65f9262feac4ea9f31d15b62c870bf359d912aba86622cfc735337ae4fa",
      "signatureAlgorithm": "RSASSA-PSS-SHA-256",
      "signedAttributes": {
        "signingScheme": "notary.x509",
        "signingTime": "2025-01-17T06:36:19Z"
      },
      "userDefinedAttributes": null,
      "unsignedAttributes": {
        "signingAgent": "notation-go/1.3.0+unreleased",
        "timestampSignature": {
          "timestamp": "[2025-01-17T06:36:19Z, 2025-01-17T06:36:20Z]",
          "certificates": [
            {
              "SHA256Fingerprint": "36e731cfa9bfd69dafb643809f6dec500902f7197daeaad86ea0159a2268a2b8",
              "issuedTo": "CN=Microsoft Public RSA Timestamping CA 2020,O=Microsoft Corporation,C=US",
              "issuedBy": "CN=Microsoft Identity Verification Root Certificate Authority 2020,O=Microsoft Corporation,C=US",
              "expiry": "2035-11-19T20:42:31Z"
            },
            {
              "SHA256Fingerprint": "2be4c1670d176be2b0e56081a7b6523487c528a7ea092febbb84ae9db03ceb9a",
              "issuedTo": "CN=Microsoft Public RSA Time Stamping Authority,OU=Microsoft Ireland Operations Limited+OU=Thales TSS ESN:F5D6-96D6-909E,O=Microsoft Corporation,L=Redmond,ST=Washington,C=US",
              "issuedBy": "CN=Microsoft Public RSA Timestamping CA 2020,O=Microsoft Corporation,C=US",
              "expiry": "2025-04-17T17:59:12Z"
            }
          ]
        }
      },
      "certificates": [
        {
          "SHA256Fingerprint": "36b3e0e0fee117d33d5664a2f56b147ddbbe8b7ca3ad2ae56498703fd782a56e",
          "issuedTo": "CN=testcert5,O=Notary,L=Seattle,ST=WA,C=US",
          "issuedBy": "CN=testcert5,O=Notary,L=Seattle,ST=WA,C=US",
          "expiry": "2025-01-18T06:34:29Z"
        }
      ],
      "signedArtifact": {
        "mediaType": "application/vnd.oci.image.manifest.v1+json",
        "digest": "sha256:99950868628ed79ebc295e01f8397dcacad35e17fb3b7a9f0fa77881ec3cef1c",
        "size": 582
      }
    }
  ]
}
`

			notation.Exec("inspect", artifact.ReferenceWithDigest(), "--output", "json").
				MatchContent(expectedOutput)
		})
	})
})
