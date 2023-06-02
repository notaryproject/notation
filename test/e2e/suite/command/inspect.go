package command

import (
	. "github.com/notaryproject/notation/test/e2e/internal/notation"
	"github.com/notaryproject/notation/test/e2e/internal/utils"
	. "github.com/notaryproject/notation/test/e2e/suite/common"
	. "github.com/onsi/ginkgo/v2"
)

var inspectSuccessfully = []string{
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
	"│   └── signingAgent: Notation/",
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
		HostInGithubAction(BaseOptions(), func(notation *utils.ExecOpts, _ *Artifact, vhost *utils.VirtualHost) {
			artifact := GenerateArtifactWithDomainHost("e2e-valid-signature", "")

			notation.Exec("inspect", "-d", artifact.ReferenceWithDigest()).
				MatchKeyWords(inspectSuccessfully...).
				MatchErrKeyWords(HTTPSRequest).
				NoMatchErrKeyWords(HTTPRequest)
		})
	})

	It("all signatures of an image with --insecure-registry flag", func() {
		HostInGithubAction(BaseOptions(), func(notation *utils.ExecOpts, _ *Artifact, vhost *utils.VirtualHost) {
			artifact := GenerateArtifactWithDomainHost("e2e-valid-signature", "")

			notation.Exec("inspect", "-d", "--insecure-registry", artifact.ReferenceWithDigest()).
				MatchKeyWords(inspectSuccessfully...).
				MatchErrKeyWords(HTTPRequest).
				NoMatchErrKeyWords(HTTPSRequest)
		})
	})
})
