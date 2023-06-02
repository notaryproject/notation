package command

import (
	. "github.com/notaryproject/notation/test/e2e/internal/notation"
	"github.com/notaryproject/notation/test/e2e/internal/utils"
	. "github.com/notaryproject/notation/test/e2e/suite/common"
	. "github.com/onsi/ginkgo/v2"
)

var _ = Describe("notation list", func() {
	It("all signatures of an image by tag schema", func() {
		Host(BaseOptions(), func(notation *utils.ExecOpts, _ *Artifact, vhost *utils.VirtualHost) {
			artifact := GenerateArtifact("e2e-valid-signature", "")

			notation.Exec("list", "-d", artifact.ReferenceWithDigest()).
				MatchKeyWords(
					"└── application/vnd.cncf.notary.signature",
					"└── sha256:273243a7a64e9312761ca0aa8f43b6ba805e677a561558143b6e92981c487339",
				).
				MatchErrKeyWords(
					"Using the referrers tag schema",
				)
		})
	})

	It("all signatures of an image with referrers api fallback to tag schema", func() {
		Host(BaseOptionsWithExperimental(), func(notation *utils.ExecOpts, _ *Artifact, vhost *utils.VirtualHost) {
			artifact := GenerateArtifact("e2e-valid-signature", "")

			notation.Exec("list", "--allow-referrers-api", "-d", artifact.ReferenceWithDigest()).
				MatchKeyWords(
					"└── application/vnd.cncf.notary.signature",
					"└── sha256:273243a7a64e9312761ca0aa8f43b6ba805e677a561558143b6e92981c487339",
				).
				MatchErrKeyWords(
					"Trying to use the referrers API",
					"404 Not Found",
					"/manifests/sha256-cc2ae4e91a31a77086edbdbf4711de48e5fa3ebdacad3403e61777a9e1a53b6f", // fallback to tag schema
				)
		})
	})

	It("by tag schema for signatures signed with referrers api", func() {
		Host(BaseOptions(), func(notation *utils.ExecOpts, _ *Artifact, vhost *utils.VirtualHost) {
			artifact := GenerateArtifactWithReferrersAPI("e2e-valid-signature", "")

			notation.Exec("list", "-d", artifact.ReferenceWithDigest()).
				MatchKeyWords("has no associated signature").
				MatchErrKeyWords(
					"Using the referrers tag schema",
				)
		})
	})

	It("all signatures of an image with referrers api", func() {
		Host(BaseOptionsWithExperimental(), func(notation *utils.ExecOpts, _ *Artifact, vhost *utils.VirtualHost) {
			artifact := GenerateArtifactWithReferrersAPI("e2e-valid-signature", "")

			notation.Exec("list", "--allow-referrers-api", "-d", artifact.ReferenceWithDigest()).
				MatchKeyWords(
					"└── application/vnd.cncf.notary.signature",
					"└── sha256:273243a7a64e9312761ca0aa8f43b6ba805e677a561558143b6e92981c487339",
				).
				MatchErrKeyWords(
					"Trying to use the referrers API",
				).
				NoMatchErrKeyWords(
					"404 Not Found",
				)
		})
	})

	It("all signatures of an image with TLS", func() {
		HostInGithubAction(BaseOptions(), func(notation *utils.ExecOpts, _ *Artifact, vhost *utils.VirtualHost) {
			artifact := GenerateArtifactWithDomainHost("e2e-valid-signature", "")

			notation.Exec("list", "-d", artifact.ReferenceWithDigest()).
				MatchKeyWords(
					"└── application/vnd.cncf.notary.signature",
					"└── sha256:273243a7a64e9312761ca0aa8f43b6ba805e677a561558143b6e92981c487339",
				).
				MatchErrKeyWords(HTTPSRequest).
				NoMatchErrKeyWords(HTTPRequest)
		})
	})

	It("all signatures of an image with --insecure-registry flag", func() {
		HostInGithubAction(BaseOptions(), func(notation *utils.ExecOpts, _ *Artifact, vhost *utils.VirtualHost) {
			artifact := GenerateArtifactWithDomainHost("e2e-valid-signature", "")

			notation.Exec("list", "-d", "--insecure-registry", artifact.ReferenceWithDigest()).
				MatchKeyWords(
					"└── application/vnd.cncf.notary.signature",
					"└── sha256:273243a7a64e9312761ca0aa8f43b6ba805e677a561558143b6e92981c487339",
				).
				MatchErrKeyWords(HTTPRequest).
				NoMatchErrKeyWords(HTTPSRequest)
		})
	})

	It("all signatures of an oci-layout", func() {
		HostWithOCILayout(BaseOptionsWithExperimental(), func(notation *utils.ExecOpts, _ *OCILayout, vhost *utils.VirtualHost) {
			ociLayout := GenerateOCILayout("e2e-valid-signature")

			notation.Exec("list", "--oci-layout", ociLayout.ReferenceWithDigest()).
				MatchKeyWords(
					"└── application/vnd.cncf.notary.signature",
					"└── sha256:273243a7a64e9312761ca0aa8f43b6ba805e677a561558143b6e92981c487339",
				)
		})
	})

	It("none signature of an oci-layout", func() {
		HostWithOCILayout(BaseOptionsWithExperimental(), func(notation *utils.ExecOpts, ociLayout *OCILayout, vhost *utils.VirtualHost) {
			notation.Exec("list", "--oci-layout", ociLayout.ReferenceWithDigest()).
				MatchKeyWords("has no associated signature")
		})
	})
})
