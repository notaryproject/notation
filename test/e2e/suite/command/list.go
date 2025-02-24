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
	. "github.com/notaryproject/notation/test/e2e/internal/notation"
	"github.com/notaryproject/notation/test/e2e/internal/utils"
	. "github.com/notaryproject/notation/test/e2e/suite/common"
	. "github.com/onsi/ginkgo/v2"
)

var _ = Describe("notation list", func() {
	It("all signatures of an image", func() {
		Host(BaseOptions(), func(notation *utils.ExecOpts, artifact *Artifact, vhost *utils.VirtualHost) {
			notation.Exec("sign", artifact.ReferenceWithDigest()).
				MatchKeyWords(SignSuccessfully)

			notation.Exec("list", "-d", artifact.ReferenceWithDigest()).
				MatchKeyWords(
					"└── application/vnd.cncf.notary.signature",
					"└── sha256:",
				)
		})
	})

	It("all signatures of an image with TLS", func() {
		HostInGithubAction(BaseOptions(), func(notation *utils.ExecOpts, artifact *Artifact, vhost *utils.VirtualHost) {
			notation.Exec("sign", artifact.DomainReferenceWithDigest()).
				MatchKeyWords(SignSuccessfully)

			notation.Exec("list", "-d", artifact.DomainReferenceWithDigest()).
				MatchKeyWords(
					"└── application/vnd.cncf.notary.signature",
					"└── sha256:",
				).
				MatchErrKeyWords(HTTPSRequest).
				NoMatchErrKeyWords(HTTPRequest)
		})
	})

	It("all signatures of an image with --insecure-registry flag", func() {
		HostInGithubAction(BaseOptions(), func(notation *utils.ExecOpts, artifact *Artifact, vhost *utils.VirtualHost) {
			notation.Exec("sign", artifact.DomainReferenceWithDigest()).
				MatchKeyWords(SignSuccessfully)

			notation.Exec("list", "-d", "--insecure-registry", artifact.DomainReferenceWithDigest()).
				MatchKeyWords(
					"└── application/vnd.cncf.notary.signature",
					"└── sha256:",
				).
				MatchErrKeyWords(HTTPRequest).
				NoMatchErrKeyWords(HTTPSRequest)
		})
	})

	It("all signatures of an oci-layout", func() {
		HostWithOCILayout(BaseOptionsWithExperimental(), func(notation *utils.ExecOpts, _ *OCILayout, vhost *utils.VirtualHost) {
			ociLayout, err := GenerateOCILayout("e2e-valid-signature")
			if err != nil {
				Fail(err.Error())
			}

			notation.Exec("list", "--oci-layout", ociLayout.ReferenceWithDigest()).
				MatchContent(ociLayout.ReferenceWithDigest() + `
└── application/vnd.cncf.notary.signature
    └── sha256:90ceaff260d657d797c408ac73564a9c7bb9d86055877c2a811f0e63b8c6524f
`)
		})
	})

	It("oci-layout with no signature", func() {
		HostWithOCILayout(BaseOptionsWithExperimental(), func(notation *utils.ExecOpts, ociLayout *OCILayout, vhost *utils.VirtualHost) {
			notation.Exec("list", "--oci-layout", ociLayout.ReferenceWithDigest()).
				MatchContent(ociLayout.ReferenceWithDigest() + " has no associated signatures\n")
		})
	})

	It("sign with --force-referrers-tag set", func() {
		Host(BaseOptions(), func(notation *utils.ExecOpts, artifact *Artifact, vhost *utils.VirtualHost) {
			notation.Exec("sign", "--force-referrers-tag", artifact.ReferenceWithDigest()).
				MatchKeyWords(SignSuccessfully)

			notation.Exec("list", artifact.ReferenceWithDigest(), "-v").
				MatchKeyWords(
					"└── application/vnd.cncf.notary.signature",
					"└── sha256:",
				)
		})
	})

	It("sign with --force-referrers-tag set to false", func() {
		Host(BaseOptions(), func(notation *utils.ExecOpts, artifact *Artifact, vhost *utils.VirtualHost) {
			notation.Exec("sign", "--force-referrers-tag=false", artifact.ReferenceWithDigest()).
				MatchKeyWords(SignSuccessfully)

			notation.Exec("list", artifact.ReferenceWithDigest(), "-v").
				MatchKeyWords(
					"└── application/vnd.cncf.notary.signature",
					"└── sha256:",
				)
		})
	})

	It("sign with --allow-referrers-api set", func() {
		Host(BaseOptionsWithExperimental(), func(notation *utils.ExecOpts, artifact *Artifact, vhost *utils.VirtualHost) {
			notation.Exec("sign", "--allow-referrers-api", artifact.ReferenceWithDigest()).
				MatchKeyWords(SignSuccessfully)

			notation.Exec("list", artifact.ReferenceWithDigest(), "-v").
				MatchKeyWords(
					"└── application/vnd.cncf.notary.signature",
					"└── sha256:",
				)

			notation.Exec("list", artifact.ReferenceWithDigest(), "--allow-referrers-api", "-v").
				MatchErrKeyWords(
					"Warning: This feature is experimental and may not be fully tested or completed and may be deprecated.",
					"Warning: flag '--allow-referrers-api' is deprecated and will be removed in future versions.",
				).
				MatchKeyWords(
					"└── application/vnd.cncf.notary.signature",
					"└── sha256:",
				)
		})
	})

	It("sign with --allow-referrers-api set to false", func() {
		Host(BaseOptionsWithExperimental(), func(notation *utils.ExecOpts, artifact *Artifact, vhost *utils.VirtualHost) {
			notation.Exec("sign", "--allow-referrers-api=false", artifact.ReferenceWithDigest()).
				MatchKeyWords(SignSuccessfully)

			notation.Exec("list", artifact.ReferenceWithDigest(), "-v").
				MatchKeyWords(
					"└── application/vnd.cncf.notary.signature",
					"└── sha256:",
				)

			notation.Exec("list", artifact.ReferenceWithDigest(), "--allow-referrers-api", "-v").
				MatchErrKeyWords(
					"Warning: This feature is experimental and may not be fully tested or completed and may be deprecated.",
					"Warning: flag '--allow-referrers-api' is deprecated and will be removed in future versions.",
				).
				MatchKeyWords(
					"└── application/vnd.cncf.notary.signature",
					"└── sha256:",
				)
		})
	})

	It("show multiple signatures", func() {
		Host(BaseOptions(), func(notation *utils.ExecOpts, _ *Artifact, vhost *utils.VirtualHost) {
			artifact := GenerateArtifact("e2e-valid-multiple-signatures", "")

			notation.Exec("list", artifact.ReferenceWithDigest()).
				MatchKeyWords(
					artifact.ReferenceWithDigest(),
					// the order of the signatures is not guaranteed
					"└── application/vnd.cncf.notary.signature",
					"    ├── ",
					"    └── ",
					"sha256:c3ebe4a20b6832328fc5078a7795ddc1114b896e13fca2add38109c3866b5fbf",
					"sha256:90ceaff260d657d797c408ac73564a9c7bb9d86055877c2a811f0e63b8c6524f",
				)
		})
	})

	It("exceed max signatures", func() {
		Host(BaseOptions(), func(notation *utils.ExecOpts, _ *Artifact, vhost *utils.VirtualHost) {
			artifact := GenerateArtifact("e2e-valid-multiple-signatures", "")

			notation.Exec("list", "--max-signatures", "1", artifact.ReferenceWithDigest()).
				MatchErrKeyWords("Warning: exceeded configured limit of max signatures 1 to examine").
				MatchContent(artifact.ReferenceWithDigest() + `
└── application/vnd.cncf.notary.signature
    └── sha256:c3ebe4a20b6832328fc5078a7795ddc1114b896e13fca2add38109c3866b5fbf
`)
		})
	})
})
