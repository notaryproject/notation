package plugin

import (
	"fmt"

	. "github.com/notaryproject/notation/test/e2e/internal/notation"
	"github.com/notaryproject/notation/test/e2e/internal/utils"
	. "github.com/notaryproject/notation/test/e2e/suite/common"
	. "github.com/onsi/ginkgo/v2"
)

var _ = Describe("notation plugin verify", func() {
	It("with basic case", func() {
		Host(BaseOptions(), func(notation *utils.ExecOpts, artifact *Artifact, vhost *utils.VirtualHost) {
			// setup plugin and plugin-key
			vhost.SetOption(AddPlugin(NotationE2EPluginPath))
			notation.Exec("key", "add", "plugin-key", "--id", "key1", "--plugin", "e2e-plugin",
				// add pluginConfig to enable generating envelope capability and update extended attribute
				"-c", fmt.Sprintf("%s=true", CapabilityEnvelopeGenerator),
				// specify verification plugin is e2e-plugin
				"-c", fmt.Sprintf("%s=e2e-plugin", HeaderVerificationPlugin)).
				MatchKeyWords("plugin-key")

			// run signing
			notation.Exec("sign", artifact.ReferenceWithDigest(), "--key", "plugin-key", "-d").
				MatchErrKeyWords(
					"Plugin get-plugin-metadata request",
					"Plugin generate-envelope request",
				).
				MatchKeyWords(SignSuccessfully)

			notation.Exec("verify", artifact.ReferenceWithDigest(), "-d").
				MatchErrKeyWords(
					"Plugin verify-signature request",
					"Plugin verify-signature response",
					`{\"verificationResults\":{\"SIGNATURE_VERIFIER.REVOCATION_CHECK\":{\"success\":true},\"SIGNATURE_VERIFIER.TRUSTED_IDENTITY\":{\"success\":true}},\"processedAttributes\":null}`).
				MatchKeyWords(VerifySuccessfully)
		})
	})

	It("with plugin minimum version 1.0.0", func() {
		Host(BaseOptions(), func(notation *utils.ExecOpts, artifact *Artifact, vhost *utils.VirtualHost) {
			// setup plugin and plugin-key
			vhost.SetOption(AddPlugin(NotationE2EPluginPath))
			notation.Exec("key", "add", "plugin-key", "--id", "key1", "--plugin", "e2e-plugin",
				// add pluginConfig to enable generating envelope capability and update extended attribute
				"-c", fmt.Sprintf("%s=true", CapabilityEnvelopeGenerator),
				// specify verification plugin is e2e-plugin
				"-c", fmt.Sprintf("%s=e2e-plugin", HeaderVerificationPlugin),
				// specify verification plugin minimum version
				"-c", fmt.Sprintf("%s=1.0.0", HeaderVerificationPluginMinVersion)).
				MatchKeyWords("plugin-key")

			// run signing
			notation.Exec("sign", artifact.ReferenceWithDigest(), "--key", "plugin-key", "-d").
				MatchErrKeyWords(
					"Plugin get-plugin-metadata request",
					"Plugin generate-envelope request",
				).
				MatchKeyWords(SignSuccessfully)

			notation.Exec("verify", artifact.ReferenceWithDigest(), "-d").
				MatchErrKeyWords(
					"Plugin verify-signature request",
					"Plugin verify-signature response",
					`{\"verificationResults\":{\"SIGNATURE_VERIFIER.REVOCATION_CHECK\":{\"success\":true},\"SIGNATURE_VERIFIER.TRUSTED_IDENTITY\":{\"success\":true}},\"processedAttributes\":null}`).
				MatchKeyWords(VerifySuccessfully)
		})
	})

	It("with plugin minimum version 1.0.11", func() {
		Host(BaseOptions(), func(notation *utils.ExecOpts, artifact *Artifact, vhost *utils.VirtualHost) {
			// setup plugin and plugin-key
			vhost.SetOption(AddPlugin(NotationE2EPluginPath))
			notation.Exec("key", "add", "plugin-key", "--id", "key1", "--plugin", "e2e-plugin",
				// add pluginConfig to enable generating envelope capability and update extended attribute
				"-c", fmt.Sprintf("%s=true", CapabilityEnvelopeGenerator),
				// specify verification plugin is e2e-plugin
				"-c", fmt.Sprintf("%s=e2e-plugin", HeaderVerificationPlugin),
				// specify verification plugin minimum version
				"-c", fmt.Sprintf("%s=1.0.11", HeaderVerificationPluginMinVersion)).
				MatchKeyWords("plugin-key")

			// run signing
			notation.Exec("sign", artifact.ReferenceWithDigest(), "--key", "plugin-key", "-d").
				MatchErrKeyWords(
					"Plugin get-plugin-metadata request",
					"Plugin generate-envelope request",
				).
				MatchKeyWords(SignSuccessfully)

			notation.ExpectFailure().Exec("verify", artifact.ReferenceWithDigest(), "-d").
				MatchErrKeyWords(
					"found plugin e2e-plugin with version 1.0.0 but signature verification needs plugin version greater than or equal to 1.0.1",
					VerifyFailed,
				)
		})
	})

	It("with plugin minimum version 0.0.1", func() {
		Host(BaseOptions(), func(notation *utils.ExecOpts, artifact *Artifact, vhost *utils.VirtualHost) {
			// setup plugin and plugin-key
			vhost.SetOption(AddPlugin(NotationE2EPluginPath))
			notation.Exec("key", "add", "plugin-key", "--id", "key1", "--plugin", "e2e-plugin",
				// add pluginConfig to enable generating envelope capability and update extended attribute
				"-c", fmt.Sprintf("%s=true", CapabilityEnvelopeGenerator),
				// specify verification plugin is e2e-plugin
				"-c", fmt.Sprintf("%s=e2e-plugin", HeaderVerificationPlugin),
				// specify verification plugin minimum version
				"-c", fmt.Sprintf("%s=0.0.1", HeaderVerificationPluginMinVersion)).
				MatchKeyWords("plugin-key")

			// run signing
			notation.Exec("sign", artifact.ReferenceWithDigest(), "--key", "plugin-key", "-d").
				MatchErrKeyWords(
					"Plugin get-plugin-metadata request",
					"Plugin generate-envelope request",
				).
				MatchKeyWords(SignSuccessfully)

			notation.Exec("verify", artifact.ReferenceWithDigest(), "-d").
				MatchErrKeyWords(
					"Plugin verify-signature request",
					"Plugin verify-signature response",
					`{\"verificationResults\":{\"SIGNATURE_VERIFIER.REVOCATION_CHECK\":{\"success\":true},\"SIGNATURE_VERIFIER.TRUSTED_IDENTITY\":{\"success\":true}},\"processedAttributes\":null}`).
				MatchKeyWords(VerifySuccessfully)
		})
	})
})
