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
				MatchKeyWords(
					"└── application/vnd.cncf.notary.signature",
					"└── sha256:90ceaff260d657d797c408ac73564a9c7bb9d86055877c2a811f0e63b8c6524f",
				)
		})
	})

	It("oci-layout with no signature", func() {
		HostWithOCILayout(BaseOptionsWithExperimental(), func(notation *utils.ExecOpts, ociLayout *OCILayout, vhost *utils.VirtualHost) {
			notation.Exec("list", "--oci-layout", ociLayout.ReferenceWithDigest()).
				MatchKeyWords("has no associated signature")
		})
	})
})
