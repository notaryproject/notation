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
		"Inspecting all signatures for signed artifact",
		"└── application/vnd.cncf.notary.signature",
		"└── sha256:",
		"├── media type:",
		"├── signature algorithm:",
		"├── signed attributes",
		"signing time:",
		"signing scheme:",
		"├── user defined attributes",
		"│   └── (empty)",
		"├── unsigned attributes",
		"│   └── signing agent: notation-go/",
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
		"Inspecting all signatures for signed artifact",
		"└── application/vnd.cncf.notary.signature",
		"└── sha256:",
		"├── media type:",
		"├── signature algorithm:",
		"├── signed attributes",
		"signing time:",
		"signing scheme:",
		"├── user defined attributes",
		"│   └── (empty)",
		"├── unsigned attributes",
		"signing agent: notation-go/",
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
			artifact := GenerateArtifact("e2e-with-timestamped-signature", "e2e-insepct-timestamped")
			expectedOutput := `Inspecting all signatures for signed artifact
localhost:5000/e2e-insepct-timestamped@sha256:f1da8cd70d6d851fa2313c8d6618f79508cf1e86877edf1c0bfe49a1b0a6467a
└── application/vnd.cncf.notary.signature
    └── sha256:e3222a9ea284789503cd2087aea775b73e049cb2c51e636a3980658e55577d18
        ├── signature algorithm: RSASSA-PSS-SHA-256
        ├── signed attributes
        │   ├── signing scheme: notary.x509
        │   └── signing time: Tue Jan 21 08:41:17 2025
        ├── user defined attributes
        │   └── purpose: e2e
        ├── unsigned attributes
        │   ├── signing agent: notation-go/1.3.0+unreleased
        │   └── timestamp signature
        │       ├── timestamp: [Tue Jan 21 08:41:16 2025, Tue Jan 21 08:41:17 2025]
        │       └── certificates
        │           ├── SHA256 fingerprint: 36e731cfa9bfd69dafb643809f6dec500902f7197daeaad86ea0159a2268a2b8
        │           │   ├── issued to: CN=Microsoft Public RSA Timestamping CA 2020,O=Microsoft Corporation,C=US
        │           │   ├── issued by: CN=Microsoft Identity Verification Root Certificate Authority 2020,O=Microsoft Corporation,C=US
        │           │   └── expiry: Mon Nov 19 20:42:31 2035
        │           └── SHA256 fingerprint: b804553ac8c88a3f71e32fe6b84f1ccef488cf45d2ebca41150e7e21dfd26e71
        │               ├── issued to: CN=Microsoft Public RSA Time Stamping Authority,OU=Microsoft America Operations+OU=Thales TSS ESN:BB73-96FD-77EF,O=Microsoft Corporation,L=Redmond,ST=Washington,C=US
        │               ├── issued by: CN=Microsoft Public RSA Timestamping CA 2020,O=Microsoft Corporation,C=US
        │               └── expiry: Wed Nov 19 18:48:47 2025
        ├── certificates
        │   └── SHA256 fingerprint: 1717fa9d18f7e9c0f609499474adfe2b8e44172454f1d6e2183d5d04f79af475
        │       ├── issued to: CN=testcert7,O=Notary,L=Seattle,ST=WA,C=US
        │       ├── issued by: CN=testcert7,O=Notary,L=Seattle,ST=WA,C=US
        │       └── expiry: Wed Jan 22 08:36:26 2025
        └── signed artifact
            ├── media type: application/vnd.oci.image.manifest.v1+json
            ├── digest: sha256:f1da8cd70d6d851fa2313c8d6618f79508cf1e86877edf1c0bfe49a1b0a6467a
            └── size: 582
`

			notation.Exec("inspect", artifact.ReferenceWithDigest()).
				MatchContent(expectedOutput)
		})
	})

	It("with timestamped oci layout and output in JSON", func() {
		Host(BaseOptions(), func(notation *utils.ExecOpts, _ *Artifact, vhost *utils.VirtualHost) {
			artifact := GenerateArtifact("e2e-with-timestamped-signature", "e2e-inspect-timestamped-json")
			expectedOutput := `{
  "mediaType": "application/vnd.oci.image.manifest.v1+json",
  "signatures": [
    {
      "digest": "sha256:e3222a9ea284789503cd2087aea775b73e049cb2c51e636a3980658e55577d18",
      "signatureAlgorithm": "RSASSA-PSS-SHA-256",
      "signedAttributes": {
        "contentType": "application/vnd.cncf.notary.payload.v1+json",
        "signingScheme": "notary.x509",
        "signingTime": "2025-01-21T08:41:17Z"
      },
      "userDefinedAttributes": {
        "purpose": "e2e"
      },
      "unsignedAttributes": {
        "signingAgent": "notation-go/1.3.0+unreleased",
        "timestampSignature": {
          "timestamp": "[2025-01-21T08:41:16.915Z, 2025-01-21T08:41:17.915Z]",
          "certificates": [
            {
              "SHA256Fingerprint": "36e731cfa9bfd69dafb643809f6dec500902f7197daeaad86ea0159a2268a2b8",
              "issuedTo": "CN=Microsoft Public RSA Timestamping CA 2020,O=Microsoft Corporation,C=US",
              "issuedBy": "CN=Microsoft Identity Verification Root Certificate Authority 2020,O=Microsoft Corporation,C=US",
              "expiry": "2035-11-19T20:42:31Z"
            },
            {
              "SHA256Fingerprint": "b804553ac8c88a3f71e32fe6b84f1ccef488cf45d2ebca41150e7e21dfd26e71",
              "issuedTo": "CN=Microsoft Public RSA Time Stamping Authority,OU=Microsoft America Operations+OU=Thales TSS ESN:BB73-96FD-77EF,O=Microsoft Corporation,L=Redmond,ST=Washington,C=US",
              "issuedBy": "CN=Microsoft Public RSA Timestamping CA 2020,O=Microsoft Corporation,C=US",
              "expiry": "2025-11-19T18:48:47Z"
            }
          ]
        }
      },
      "certificates": [
        {
          "SHA256Fingerprint": "1717fa9d18f7e9c0f609499474adfe2b8e44172454f1d6e2183d5d04f79af475",
          "issuedTo": "CN=testcert7,O=Notary,L=Seattle,ST=WA,C=US",
          "issuedBy": "CN=testcert7,O=Notary,L=Seattle,ST=WA,C=US",
          "expiry": "2025-01-22T08:36:26Z"
        }
      ],
      "signedArtifact": {
        "mediaType": "application/vnd.oci.image.manifest.v1+json",
        "digest": "sha256:f1da8cd70d6d851fa2313c8d6618f79508cf1e86877edf1c0bfe49a1b0a6467a",
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

	It("with no signature in text format", func() {
		Host(BaseOptions(), func(notation *utils.ExecOpts, _ *Artifact, vhost *utils.VirtualHost) {
			artifact := GenerateArtifact("e2e", "e2e-inspect-no-signature")
			expectedOutput := "localhost:5000/e2e-inspect-no-signature@sha256:b8479de3f88fb259a0a9ea82a5b2a052a1ef3c4ebbcfc61482d5ae4c831f8af9 has no associated signature\n"
			notation.Exec("inspect", artifact.ReferenceWithDigest()).
				MatchContent(expectedOutput)
		})
	})

	It("with no signature in JSON format", func() {
		Host(BaseOptions(), func(notation *utils.ExecOpts, _ *Artifact, vhost *utils.VirtualHost) {
			artifact := GenerateArtifact("e2e", "e2e-inspect-no-signature-json")
			expectedOutput := `{
  "mediaType": "application/vnd.oci.image.manifest.v1+json",
  "signatures": []
}
`
			notation.Exec("inspect", "--output", "json", artifact.ReferenceWithDigest()).
				MatchContent(expectedOutput)
		})
	})

	It("with invalid timestamp signature in text format", func() {
		Host(BaseOptions(), func(notation *utils.ExecOpts, _ *Artifact, vhost *utils.VirtualHost) {
			artifact := GenerateArtifact("e2e-with-invalid-timestamped-signature", "e2e-inspect-invalid-timstamped")
			expectedOutput := `Inspecting all signatures for signed artifact
localhost:5000/e2e-inspect-invalid-timstamped@sha256:f1da8cd70d6d851fa2313c8d6618f79508cf1e86877edf1c0bfe49a1b0a6467a
└── application/vnd.cncf.notary.signature
    └── sha256:eee3eec7d2947f77713484753bea67879ff62c08a73a49a41151ed18c4d1c000
        ├── signature algorithm: RSASSA-PSS-SHA-256
        ├── signed attributes
        │   ├── signing scheme: notary.x509
        │   └── signing time: Tue Jan 21 08:41:17 2025
        ├── user defined attributes
        │   └── purpose: e2e
        ├── unsigned attributes
        │   ├── signing agent: notation-go/1.3.0+unreleased
        │   └── timestamp signature
        │       └── error: failed to parse timestamp countersignature: cms: syntax error: invalid signed data: failed to convert from BER to DER: asn1: syntax error: decoding BER length octets: short form length octets value should be less or equal to the subsequent octets length
        ├── certificates
        │   └── SHA256 fingerprint: 1717fa9d18f7e9c0f609499474adfe2b8e44172454f1d6e2183d5d04f79af475
        │       ├── issued to: CN=testcert7,O=Notary,L=Seattle,ST=WA,C=US
        │       ├── issued by: CN=testcert7,O=Notary,L=Seattle,ST=WA,C=US
        │       └── expiry: Wed Jan 22 08:36:26 2025
        └── signed artifact
            ├── media type: application/vnd.oci.image.manifest.v1+json
            ├── digest: sha256:f1da8cd70d6d851fa2313c8d6618f79508cf1e86877edf1c0bfe49a1b0a6467a
            └── size: 582
`
			notation.Exec("inspect", artifact.ReferenceWithDigest()).
				MatchContent(expectedOutput)
		})
	})

	It("with invalid timestamp signature in json format", func() {
		Host(BaseOptions(), func(notation *utils.ExecOpts, _ *Artifact, vhost *utils.VirtualHost) {
			artifact := GenerateArtifact("e2e-with-invalid-timestamped-signature", "e2e-inspect-invalid-timstamped")
			expectedOutput := `{
  "mediaType": "application/vnd.oci.image.manifest.v1+json",
  "signatures": [
    {
      "digest": "sha256:eee3eec7d2947f77713484753bea67879ff62c08a73a49a41151ed18c4d1c000",
      "signatureAlgorithm": "RSASSA-PSS-SHA-256",
      "signedAttributes": {
        "contentType": "application/vnd.cncf.notary.payload.v1+json",
        "signingScheme": "notary.x509",
        "signingTime": "2025-01-21T08:41:17Z"
      },
      "userDefinedAttributes": {
        "purpose": "e2e"
      },
      "unsignedAttributes": {
        "signingAgent": "notation-go/1.3.0+unreleased",
        "timestampSignature": {
          "error": "failed to parse timestamp countersignature: cms: syntax error: invalid signed data: failed to convert from BER to DER: asn1: syntax error: decoding BER length octets: short form length octets value should be less or equal to the subsequent octets length"
        }
      },
      "certificates": [
        {
          "SHA256Fingerprint": "1717fa9d18f7e9c0f609499474adfe2b8e44172454f1d6e2183d5d04f79af475",
          "issuedTo": "CN=testcert7,O=Notary,L=Seattle,ST=WA,C=US",
          "issuedBy": "CN=testcert7,O=Notary,L=Seattle,ST=WA,C=US",
          "expiry": "2025-01-22T08:36:26Z"
        }
      ],
      "signedArtifact": {
        "mediaType": "application/vnd.oci.image.manifest.v1+json",
        "digest": "sha256:f1da8cd70d6d851fa2313c8d6618f79508cf1e86877edf1c0bfe49a1b0a6467a",
        "size": 582
      }
    }
  ]
}
`
			notation.Exec("inspect", "--output", "json", artifact.ReferenceWithDigest()).
				MatchContent(expectedOutput)
		})
	})
})
