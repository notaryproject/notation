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
				"--plugin-config", fmt.Sprintf("%s=true", CapabilityEnvelopeGenerator),
				// specify verification plugin is e2e-plugin
				"--plugin-config", fmt.Sprintf("%s=e2e-plugin", HeaderVerificationPlugin)).
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

	It("with plugin revocation check failed", func() {
		Host(BaseOptions(), func(notation *utils.ExecOpts, artifact *Artifact, vhost *utils.VirtualHost) {
			// setup plugin and plugin-key
			vhost.SetOption(AddPlugin(NotationE2EPluginPath))
			notation.Exec("key", "add", "plugin-key", "--id", "key1", "--plugin", "e2e-plugin",
				// add pluginConfig to enable generating envelope capability and update extended attribute
				"--plugin-config", fmt.Sprintf("%s=true", CapabilityEnvelopeGenerator),
				// specify verification plugin is e2e-plugin
				"--plugin-config", fmt.Sprintf("%s=e2e-plugin", HeaderVerificationPlugin)).
				MatchKeyWords("plugin-key")

			// run signing
			notation.Exec("sign", artifact.ReferenceWithDigest(), "--key", "plugin-key", "-d").
				MatchErrKeyWords(
					"Plugin get-plugin-metadata request",
					"Plugin generate-envelope request",
				).
				MatchKeyWords(SignSuccessfully)

			notation.ExpectFailure().Exec("verify", artifact.ReferenceWithDigest(), "-d",
				// set revocation check failed for plugin
				"--plugin-config", fmt.Sprintf("%s=failed", CapabilityRevocationCheckVerifier),
			).
				MatchErrKeyWords(
					"Plugin verify-signature request",
					"Plugin verify-signature response",
					`revocation check by verification plugin \"e2e-plugin\" failed with reason \"revocation check failed\"`,
					VerifyFailed)
		})
	})

	It("with plugin trusted identity check failed", func() {
		Host(BaseOptions(), func(notation *utils.ExecOpts, artifact *Artifact, vhost *utils.VirtualHost) {
			// setup plugin and plugin-key
			vhost.SetOption(AddPlugin(NotationE2EPluginPath))
			notation.Exec("key", "add", "plugin-key", "--id", "key1", "--plugin", "e2e-plugin",
				// add pluginConfig to enable generating envelope capability and update extended attribute
				"--plugin-config", fmt.Sprintf("%s=true", CapabilityEnvelopeGenerator),
				// specify verification plugin is e2e-plugin
				"--plugin-config", fmt.Sprintf("%s=e2e-plugin", HeaderVerificationPlugin)).
				MatchKeyWords("plugin-key")

			// run signing
			notation.Exec("sign", artifact.ReferenceWithDigest(), "--key", "plugin-key", "-d").
				MatchErrKeyWords(
					"Plugin get-plugin-metadata request",
					"Plugin generate-envelope request",
				).
				MatchKeyWords(SignSuccessfully)

			notation.ExpectFailure().Exec("verify", artifact.ReferenceWithDigest(), "-d",
				// set trusted identity check failed for plugin
				"--plugin-config", fmt.Sprintf("%s=failed", CapabilityTrustedIdentityVerifier),
			).
				MatchErrKeyWords(
					"Plugin verify-signature request",
					"Plugin verify-signature response",
					`trusted identify verification by plugin \"e2e-plugin\" failed with reason \"trusted identity check failed\"`,
					VerifyFailed)
		})
	})

	It("with plugin minimum version 1.0.0", func() {
		Host(BaseOptions(), func(notation *utils.ExecOpts, artifact *Artifact, vhost *utils.VirtualHost) {
			// setup plugin and plugin-key
			vhost.SetOption(AddPlugin(NotationE2EPluginPath))
			notation.Exec("key", "add", "plugin-key", "--id", "key1", "--plugin", "e2e-plugin",
				// add pluginConfig to enable generating envelope capability and update extended attribute
				"--plugin-config", fmt.Sprintf("%s=true", CapabilityEnvelopeGenerator),
				// specify verification plugin is e2e-plugin
				"--plugin-config", fmt.Sprintf("%s=e2e-plugin", HeaderVerificationPlugin),
				// specify verification plugin minimum version
				"--plugin-config", fmt.Sprintf("%s=1.0.0", HeaderVerificationPluginMinVersion)).
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
				"--plugin-config", fmt.Sprintf("%s=true", CapabilityEnvelopeGenerator),
				// specify verification plugin is e2e-plugin
				"--plugin-config", fmt.Sprintf("%s=e2e-plugin", HeaderVerificationPlugin),
				// specify verification plugin minimum version
				"--plugin-config", fmt.Sprintf("%s=1.0.11", HeaderVerificationPluginMinVersion)).
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
				"--plugin-config", fmt.Sprintf("%s=true", CapabilityEnvelopeGenerator),
				// specify verification plugin is e2e-plugin
				"--plugin-config", fmt.Sprintf("%s=e2e-plugin", HeaderVerificationPlugin),
				// specify verification plugin minimum version
				"--plugin-config", fmt.Sprintf("%s=0.0.1", HeaderVerificationPluginMinVersion)).
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
