package trustpolicy

import (
	"fmt"

	. "github.com/notaryproject/notation/test/e2e/internal/notation"
	"github.com/notaryproject/notation/test/e2e/internal/utils"
	. "github.com/notaryproject/notation/test/e2e/suite/common"
	. "github.com/onsi/ginkgo/v2"
)

// trustPolicyLink is a tutorial link for creating Notation's trust policy.
const trustPolicyLink = "https://notaryproject.dev/docs/quickstart/#create-a-trust-policy"

var _ = Describe("notation trust policy registryScope test", func() {
	It("empty registryScope", func() {
		Host(BaseOptions(), func(notation *utils.ExecOpts, artifact *Artifact, vhost *utils.VirtualHost) {
			// update trustpolicy.json
			vhost.SetOption(AddTrustPolicyOption("empty_registry_scope_trustpolicy.json"))

			// test localhost:5000/test-repo
			OldNotation().Exec("sign", artifact.ReferenceWithDigest()).MatchKeyWords(SignSuccessfully)
			notation.ExpectFailure().Exec("verify", artifact.ReferenceWithDigest()).
				MatchErrKeyWords("trust policy statement \"e2e\" has zero registry scopes")
		})
	})

	It("malformed registryScope", func() {
		Host(BaseOptions(), func(notation *utils.ExecOpts, artifact *Artifact, vhost *utils.VirtualHost) {
			// update trustpolicy.json
			vhost.SetOption(AddTrustPolicyOption("malformed_registry_scope_trustpolicy.json"))

			OldNotation().Exec("sign", artifact.ReferenceWithDigest()).MatchKeyWords(SignSuccessfully)
			notation.ExpectFailure().Exec("verify", artifact.ReferenceWithDigest()).
				MatchErrKeyWords(`registry scope "localhost:5000\\test-repo" is not valid, make sure it is a fully qualified registry URL without the scheme/protocol, e.g domain.com/my/repository OR a local trust policy scope, e.g local/myOCILayout`)
		})
	})

	It("registryScope with a repository", func() {
		Host(BaseOptions(), func(notation *utils.ExecOpts, _ *Artifact, vhost *utils.VirtualHost) {
			// update trustpolicy.json
			vhost.SetOption(AddTrustPolicyOption("registry_scope_trustpolicy.json"))

			// generate an artifact with given repository name
			artifact := GenerateArtifact("", "test-repo")

			// test localhost:5000/test-repo
			notation.Exec("sign", artifact.ReferenceWithDigest()).MatchKeyWords(SignSuccessfully)
			notation.Exec("verify", artifact.ReferenceWithDigest()).MatchKeyWords(VerifySuccessfully)
		})
	})

	It("registryScope with multiple repositories", func() {
		Host(BaseOptions(), func(notation *utils.ExecOpts, _ *Artifact, vhost *utils.VirtualHost) {
			// update trustpolicy.json
			vhost.SetOption(AddTrustPolicyOption("multiple_registry_scope_trustpolicy.json"))

			// generate an artifact with given repository name
			artifact2 := GenerateArtifact("", "test-repo2")
			artifact3 := GenerateArtifact("", "test-repo3")

			// test localhost:5000/test-repo2
			notation.Exec("sign", artifact2.ReferenceWithDigest()).MatchKeyWords(SignSuccessfully)
			notation.Exec("verify", artifact2.ReferenceWithDigest()).MatchKeyWords(VerifySuccessfully)

			// test localhost:5000/test-repo3
			notation.Exec("sign", artifact3.ReferenceWithDigest()).MatchKeyWords(SignSuccessfully)
			notation.Exec("verify", artifact3.ReferenceWithDigest()).MatchKeyWords(VerifySuccessfully)
		})
	})

	It("registryScope with any(*) repository", func() {
		Host(BaseOptions(), func(notation *utils.ExecOpts, _ *Artifact, vhost *utils.VirtualHost) {
			// update trustpolicy.json
			vhost.SetOption(AddTrustPolicyOption("any_registry_scope_trust_policy.json"))

			// generate an artifact with given repository name
			artifact4 := GenerateArtifact("", "test-repo4")
			artifact5 := GenerateArtifact("", "test-repo5")

			// test localhost:5000/test-repo4
			notation.Exec("sign", artifact4.ReferenceWithDigest()).MatchKeyWords(SignSuccessfully)
			notation.Exec("verify", artifact4.ReferenceWithDigest()).MatchKeyWords(VerifySuccessfully)

			// test localhost:5000/test-repo5
			notation.Exec("sign", artifact5.ReferenceWithDigest()).MatchKeyWords(SignSuccessfully)
			notation.Exec("verify", artifact5.ReferenceWithDigest()).MatchKeyWords(VerifySuccessfully)
		})
	})

	It("overlapped registryScope", func() {
		Host(BaseOptions(), func(notation *utils.ExecOpts, _ *Artifact, vhost *utils.VirtualHost) {
			// update trustpolicy.json
			vhost.SetOption(AddTrustPolicyOption("overlapped_registry_scope_trustpolicy.json"))

			artifact := GenerateArtifact("", "test-repo6")

			// test localhost:5000/test-repo
			OldNotation().Exec("sign", artifact.ReferenceWithDigest()).MatchKeyWords(SignSuccessfully)
			notation.ExpectFailure().Exec("verify", artifact.ReferenceWithDigest()).
				MatchErrKeyWords("registry scope \"localhost:5000/test-repo6\" is present in multiple trust policy statements")
		})
	})

	It("wildcard plus specific repo registryScope", func() {
		Host(BaseOptions(), func(notation *utils.ExecOpts, _ *Artifact, vhost *utils.VirtualHost) {
			// update trustpolicy.json
			vhost.SetOption(AddTrustPolicyOption("wildcard_plus_other_registry_scope_trustpolicy.json"))

			artifact := GenerateArtifact("", "test-repo7")

			// test localhost:5000/test-repo
			OldNotation().Exec("sign", artifact.ReferenceWithDigest()).MatchKeyWords(SignSuccessfully)
			notation.ExpectFailure().Exec("verify", artifact.ReferenceWithDigest()).
				MatchErrKeyWords("trust policy statement \"e2e\" uses wildcard registry scope '*', a wildcard scope cannot be used in conjunction with other scope values")
		})
	})

	It("invalid registryScope", func() {
		Host(BaseOptions(), func(notation *utils.ExecOpts, artifact *Artifact, vhost *utils.VirtualHost) {
			// update trustpolicy.json
			vhost.SetOption(AddTrustPolicyOption("invalid_registry_scope_trustpolicy.json"))

			// test localhost:5000/test-repo
			OldNotation().Exec("sign", artifact.ReferenceWithDigest()).MatchKeyWords(SignSuccessfully)
			notation.ExpectFailure().Exec("verify", artifact.ReferenceWithDigest()).
				MatchErrContent(fmt.Sprintf("Error: signature verification failed: artifact %q has no applicable trust policy. Trust policy applicability for a given artifact is determined by registryScopes. To create a trust policy, see: %s\n", artifact.ReferenceWithDigest(), trustPolicyLink))
		})
	})
})
