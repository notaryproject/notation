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
	"github.com/notaryproject/notation-core-go/signature/cose"
	. "github.com/notaryproject/notation/test/e2e/internal/notation"
	"github.com/notaryproject/notation/test/e2e/internal/utils"
	. "github.com/notaryproject/notation/test/e2e/suite/common"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

const (
	TamperKeyID                 = "TAMPER_KEY_ID"
	TamperSignature             = "TAMPER_SIGNATURE"
	TamperSignatureAlgorithm    = "TAMPER_SIGNATURE_ALGORITHM"
	TamperCertificateChain      = "TAMPER_CERTIFICATE_CHAIN"
	TamperSignatureEnvelope     = "TAMPER_SIGNATURE_ENVELOPE"
	TamperSignatureEnvelopeType = "TAMPER_SIGNATURE_ENVELOPE_TYPE"
	TamperAnnotation            = "TAMPER_ANNOTATION"
)

var _ = Describe("notation plugin sign", func() {
	It("with JWS format and capability SIGNATURE_GENERATOR.RAW", func() {
		Host(BaseOptions(), func(notation *utils.ExecOpts, artifact *Artifact, vhost *utils.VirtualHost) {
			// setup plugin and plugin-key
			vhost.SetOption(AddPlugin(NotationE2EPluginPath))
			notation.Exec("key", "add", "plugin-key", "--id", "key1", "--plugin", "e2e-plugin",
				"--plugin-config", string(CapabilitySignatureGenerator)+"=true").
				MatchKeyWords("plugin-key")

			// run signing
			notation.Exec("sign", artifact.ReferenceWithDigest(), "--key", "plugin-key", "-d").
				MatchErrKeyWords(
					"Plugin get-plugin-metadata request",
					"Plugin describe-key request",
					"Plugin generate-signature request",
				).
				MatchKeyWords(SignSuccessfully)

			OldNotation().Exec("verify", artifact.ReferenceWithDigest()).
				MatchKeyWords(VerifySuccessfully)
		})
	})

	It("with COSE format and capability SIGNATURE_GENERATOR.RAW", func() {
		Host(BaseOptions(), func(notation *utils.ExecOpts, artifact *Artifact, vhost *utils.VirtualHost) {
			// setup plugin and plugin-key
			vhost.SetOption(AddPlugin(NotationE2EPluginPath))
			notation.Exec("key", "add", "plugin-key", "--id", "key1", "--plugin", "e2e-plugin",
				"--plugin-config", string(CapabilitySignatureGenerator)+"=true").
				MatchKeyWords("plugin-key")

			// run signing
			notation.Exec("sign", artifact.ReferenceWithDigest(), "--key", "plugin-key", "--signature-format", "cose", "-d").
				MatchErrKeyWords(
					"Plugin get-plugin-metadata request",
					"Plugin describe-key request",
					"Plugin generate-signature request",
				).
				MatchKeyWords(SignSuccessfully)

			OldNotation().Exec("verify", artifact.ReferenceWithDigest()).
				MatchKeyWords(VerifySuccessfully)
		})
	})

	It("with JWS format and capability SIGNATURE_GENERATOR.ENVELOPE", func() {
		Host(BaseOptions(), func(notation *utils.ExecOpts, artifact *Artifact, vhost *utils.VirtualHost) {
			// setup plugin and plugin-key
			vhost.SetOption(AddPlugin(NotationE2EPluginPath))
			notation.Exec("key", "add", "plugin-key", "--id", "key1", "--plugin", "e2e-plugin",
				"--plugin-config", string(CapabilityEnvelopeGenerator)+"=true").
				MatchKeyWords("plugin-key")

			// run signing
			notation.Exec("sign", artifact.ReferenceWithDigest(), "--key", "plugin-key", "-d").
				MatchErrKeyWords(
					"Plugin get-plugin-metadata request",
					"Plugin generate-envelope request",
				).
				MatchKeyWords(SignSuccessfully)

			OldNotation().Exec("verify", artifact.ReferenceWithDigest()).
				MatchKeyWords(VerifySuccessfully)
		})
	})

	It("with COSE format and capability SIGNATURE_GENERATOR.ENVELOPE", func() {
		Host(BaseOptions(), func(notation *utils.ExecOpts, artifact *Artifact, vhost *utils.VirtualHost) {
			// setup plugin and plugin-key
			vhost.SetOption(AddPlugin(NotationE2EPluginPath))
			notation.Exec("key", "add", "plugin-key", "--id", "key1", "--plugin", "e2e-plugin",
				"--plugin-config", string(CapabilityEnvelopeGenerator)+"=true").
				MatchKeyWords("plugin-key")

			// run signing
			notation.Exec("sign", artifact.ReferenceWithDigest(), "--key", "plugin-key", "--signature-format", "cose", "-d").
				MatchErrKeyWords(
					"Plugin get-plugin-metadata request",
					"Plugin generate-envelope request",
				).
				MatchKeyWords(SignSuccessfully)

			OldNotation().Exec("verify", artifact.ReferenceWithDigest()).
				MatchKeyWords(VerifySuccessfully)
		})
	})

	It("with capability SIGNATURE_GENERATOR.RAW and tampered KeyID", func() {
		Host(BaseOptions(), func(notation *utils.ExecOpts, artifact *Artifact, vhost *utils.VirtualHost) {
			// setup plugin and plugin-key
			vhost.SetOption(AddPlugin(NotationE2EPluginPath))
			notation.Exec("key", "add", "plugin-key", "--id", "key1", "--plugin", "e2e-plugin",
				"--plugin-config", string(CapabilitySignatureGenerator)+"=true",
				"--plugin-config", TamperKeyID+"=key10").
				MatchKeyWords("plugin-key")

			// run signing
			notation.ExpectFailure().Exec("sign", artifact.ReferenceWithDigest(), "--key", "plugin-key", "-d").
				MatchErrKeyWords(
					"Plugin get-plugin-metadata request",
					"Plugin describe-key request",
					"Plugin generate-signature request",
					`keyID in generateSignature response "key10" does not match request "key1"`,
				)
		})
	})

	It("with capability SIGNATURE_GENERATOR.RAW and tampered signature", func() {
		Host(BaseOptions(), func(notation *utils.ExecOpts, artifact *Artifact, vhost *utils.VirtualHost) {
			// setup plugin and plugin-key
			vhost.SetOption(AddPlugin(NotationE2EPluginPath))
			notation.Exec("key", "add", "plugin-key", "--id", "key1", "--plugin", "e2e-plugin",
				"--plugin-config", string(CapabilitySignatureGenerator)+"=true",
				"--plugin-config", TamperSignature+"=invalid_sig").
				MatchKeyWords("plugin-key")

			// run signing
			notation.ExpectFailure().Exec("sign", artifact.ReferenceWithDigest(), "--key", "plugin-key", "-d").
				MatchErrKeyWords(
					"Plugin get-plugin-metadata request",
					"Plugin describe-key request",
					"Plugin generate-signature request",
					"generated signature failed verification: signature is invalid. Error: crypto/rsa: verification error",
				)
		})
	})

	It("with capability SIGNATURE_GENERATOR.RAW and tampered signatureAlgorithm", func() {
		Skip("signatureAlgorithm returned by plugin is not verified.")
		Host(BaseOptions(), func(notation *utils.ExecOpts, artifact *Artifact, vhost *utils.VirtualHost) {
			// setup plugin and plugin-key
			vhost.SetOption(AddPlugin(NotationE2EPluginPath))
			notation.Exec("key", "add", "plugin-key", "--id", "key1", "--plugin", "e2e-plugin",
				"--plugin-config", string(CapabilitySignatureGenerator)+"=true",
				"--plugin-config", TamperSignatureAlgorithm+"=invalid_alg").
				MatchKeyWords("plugin-key")

			// run signing
			notation.ExpectFailure().Exec("sign", artifact.ReferenceWithDigest(), "--key", "plugin-key", "-d").
				MatchErrKeyWords(
					"Plugin get-plugin-metadata request",
					"Plugin describe-key request",
					"Plugin generate-signature request",
				)
		})
	})

	It("with capability SIGNATURE_GENERATOR.RAW and tampered certificate chain", func() {
		Host(BaseOptions(), func(notation *utils.ExecOpts, artifact *Artifact, vhost *utils.VirtualHost) {
			// setup plugin and plugin-key
			vhost.SetOption(AddPlugin(NotationE2EPluginPath))
			notation.Exec("key", "add", "plugin-key", "--id", "key1", "--plugin", "e2e-plugin",
				"--plugin-config", string(CapabilitySignatureGenerator)+"=true",
				"--plugin-config", TamperCertificateChain+"=invalid_cert_chain").
				MatchKeyWords("plugin-key")

			// run signing
			notation.ExpectFailure().Exec("sign", artifact.ReferenceWithDigest(), "--key", "plugin-key", "-d").
				MatchErrKeyWords(
					"Plugin get-plugin-metadata request",
					"Plugin describe-key request",
					"Plugin generate-signature request",
					"x509: malformed certificate",
				)
		})
	})

	It("with capability SIGNATURE_GENERATOR.ENVELOPE and tampered signature envelope", func() {
		Host(BaseOptions(), func(notation *utils.ExecOpts, artifact *Artifact, vhost *utils.VirtualHost) {
			// setup plugin and plugin-key
			vhost.SetOption(AddPlugin(NotationE2EPluginPath))
			notation.Exec("key", "add", "plugin-key", "--id", "key1", "--plugin", "e2e-plugin",
				"--plugin-config", string(CapabilityEnvelopeGenerator)+"=true",
				"--plugin-config", TamperSignatureEnvelope+"={}").
				MatchKeyWords("plugin-key")

			// run signing
			notation.ExpectFailure().Exec("sign", artifact.ReferenceWithDigest(), "--key", "plugin-key", "-d").
				MatchErrKeyWords(
					"Plugin get-plugin-metadata request",
					"Plugin generate-envelope request",
					"Verifying signature envelope generated by the plugin",
					"generated signature failed verification: certificate chain is not present",
				)
		})
	})

	It("with capability SIGNATURE_GENERATOR.ENVELOPE and tampered signature envelope type", func() {
		Host(BaseOptions(), func(notation *utils.ExecOpts, artifact *Artifact, vhost *utils.VirtualHost) {
			// setup plugin and plugin-key
			vhost.SetOption(AddPlugin(NotationE2EPluginPath))
			notation.Exec("key", "add", "plugin-key", "--id", "key1", "--plugin", "e2e-plugin",
				"--plugin-config", string(CapabilityEnvelopeGenerator)+"=true",
				"--plugin-config", TamperSignatureEnvelopeType+"="+cose.MediaTypeEnvelope).
				MatchKeyWords("plugin-key")

			// run signing
			notation.ExpectFailure().Exec("sign", artifact.ReferenceWithDigest(), "--key", "plugin-key", "-d").
				MatchErrKeyWords(
					"Plugin get-plugin-metadata request",
					"Plugin generate-envelope request",
					`signatureEnvelopeType in generateEnvelope response "application/cose" does not match request "application/jose+json"`,
				)
		})
	})

	It("with capability SIGNATURE_GENERATOR.ENVELOPE and tampered annotation", func() {
		Host(BaseOptions(), func(notation *utils.ExecOpts, artifact *Artifact, vhost *utils.VirtualHost) {
			Skip("annotation returned by plugin is not processed")
			// setup plugin and plugin-key
			vhost.SetOption(AddPlugin(NotationE2EPluginPath))
			notation.Exec("key", "add", "plugin-key", "--id", "key1", "--plugin", "e2e-plugin",
				"--plugin-config", string(CapabilityEnvelopeGenerator)+"=true",
				"--plugin-config", TamperAnnotation+"=k1=v1").
				MatchKeyWords("plugin-key")

			// run signing
			notation.Exec("sign", artifact.ReferenceWithDigest(), "--key", "plugin-key", "-d").
				MatchErrKeyWords(
					"Plugin get-plugin-metadata request",
					"Plugin generate-envelope request",
				)

			// check signature annotation
			descriptors, err := artifact.SignatureDescriptors()
			Expect(err).ShouldNot(HaveOccurred())

			// should have 1 signature
			Expect(len(descriptors)).Should(Equal(1))
			// should have the annotation
			Expect(descriptors[0].Annotations).Should(HaveKeyWithValue("k1", "v1"))
		})
	})

	It("incorrect NOTATION_LIBEXEC path", func() {
		Host(BaseOptions(), func(notation *utils.ExecOpts, artifact *Artifact, vhost *utils.VirtualHost) {
			// setup incorrect NOTATION_LIBEXEC path
			vhost.SetOption(AddPlugin(NotationE2EPluginPath))
			notation.Exec("key", "add", "plugin-key", "--id", "key1", "--plugin", "e2e-plugin",
				"--plugin-config", string(CapabilityEnvelopeGenerator)+"=true",
				"--plugin-config", TamperAnnotation+"=k1=v1").
				MatchKeyWords("plugin-key")

			vhost.UpdateEnv(map[string]string{"NOTATION_LIBEXEC": "/not/exist"})

			// run signing
			notation.ExpectFailure().Exec("sign", artifact.ReferenceWithDigest(), "--key", "plugin-key", "-d").
				MatchErrKeyWords("no such file or directory")
		})
	})

	It("correct NOTATION_LIBEXEC path", func() {
		Host(BaseOptions(), func(notation *utils.ExecOpts, artifact *Artifact, vhost *utils.VirtualHost) {
			// setup incorrect NOTATION_LIBEXEC path
			vhost.SetOption(AddPlugin(NotationE2EPluginPath))
			notation.Exec("key", "add", "plugin-key", "--id", "key1", "--plugin", "e2e-plugin",
				"--plugin-config", string(CapabilityEnvelopeGenerator)+"=true",
				"--plugin-config", TamperAnnotation+"=k1=v1").
				MatchKeyWords("plugin-key")

			vhost.UpdateEnv(map[string]string{"NOTATION_LIBEXEC": vhost.AbsolutePath(NotationDirName)})

			// run signing
			notation.Exec("sign", artifact.ReferenceWithDigest(), "--key", "plugin-key", "-d").
				MatchKeyWords("Successfully signed")
		})
	})
})
