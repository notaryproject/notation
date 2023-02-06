package trustpolicy

import (
	"path/filepath"

	. "github.com/notaryproject/notation/test/e2e/internal/notation"
	"github.com/notaryproject/notation/test/e2e/internal/utils"
	. "github.com/notaryproject/notation/test/e2e/suite/common"
	. "github.com/onsi/ginkgo/v2"
)

var _ = Describe("notation trust policy verification level test", func() {
	It("strict level with expired signature", func() {
		Host(BaseOptions(), func(notation *utils.ExecOpts, _ *Artifact, vhost *utils.VirtualHost) {
			artifact := GenerateArtifact("e2e-expired-signature", "")

			notation.ExpectFailure().Exec("verify", artifact.ReferenceWithDigest(), "-v").
				MatchErrKeyWords("expiry validation failed.",
					VerifyFailed)
		})
	})

	It("strict level with expired authentic timestamp", func() {
		Host(nil, func(notation *utils.ExecOpts, _ *Artifact, vhost *utils.VirtualHost) {
			artifact := GenerateArtifact("e2e-with-expired-cert", "")

			vhost.SetOption(AuthOption("", ""),
				AddTrustPolicyOption("trustpolicy.json"),
				AddTrustStoreOption("e2e", filepath.Join(NotationE2EConfigPath, "localkeys", "expired_e2e.crt")))

			notation.ExpectFailure().Exec("verify", artifact.ReferenceWithDigest(), "-v").
				MatchErrKeyWords("authenticTimestamp validation failed",
					VerifyFailed)
		})
	})

	It("strict level with invalid authenticity", func() {
		Host(nil, func(notation *utils.ExecOpts, _ *Artifact, vhost *utils.VirtualHost) {
			vhost.SetOption(AuthOption("", ""),
				AddTrustPolicyOption("trustpolicy.json"),
				AddTrustStoreOption("e2e", filepath.Join(NotationE2ELocalKeysDir, "new_e2e.crt")))

			// the artifact signed with a different cert from the cert in
			// trust store.
			artifact := GenerateArtifact("e2e-valid-signature", "")

			notation.ExpectFailure().Exec("verify", artifact.ReferenceWithDigest(), "-v").
				MatchErrKeyWords("authenticity validation failed",
					VerifyFailed)
		})
	})

	It("strict level with invalid integrity", func() {
		Host(BaseOptions(), func(notation *utils.ExecOpts, _ *Artifact, vhost *utils.VirtualHost) {
			artifact := GenerateArtifact("e2e-invalid-signature", "")

			notation.ExpectFailure().Exec("verify", artifact.ReferenceWithDigest(), "-v").
				MatchErrKeyWords("integrity validation failed",
					VerifyFailed)
		})
	})

	It("permissive level with expired signature", func() {
		Host(BaseOptions(), func(notation *utils.ExecOpts, _ *Artifact, vhost *utils.VirtualHost) {
			vhost.SetOption(AddTrustPolicyOption("permissive_trustpolicy.json"))

			artifact := GenerateArtifact("e2e-expired-signature", "")

			notation.Exec("verify", artifact.ReferenceWithDigest(), "-v").
				MatchErrKeyWords("expiry was set to \"log\" and failed with error: digital signature has expired").
				MatchKeyWords(VerifySuccessfully)
		})
	})

	It("permissive level with expired authentic timestamp", func() {
		Host(nil, func(notation *utils.ExecOpts, _ *Artifact, vhost *utils.VirtualHost) {
			artifact := GenerateArtifact("e2e-with-expired-cert", "")

			vhost.SetOption(AuthOption("", ""),
				AddTrustPolicyOption("permissive_trustpolicy.json"),
				AddTrustStoreOption("e2e", filepath.Join(NotationE2EConfigPath, "localkeys", "expired_e2e.crt")))

			notation.Exec("verify", artifact.ReferenceWithDigest(), "-v").
				MatchErrKeyWords("Warning: authenticTimestamp was set to \"log\"",
					"error: certificate \"O=Internet Widgits Pty Ltd,ST=Some-State,C=AU\" is not valid anymore, it was expired").
				MatchKeyWords(VerifySuccessfully)
		})
	})

	It("permissive level with invalid authenticity", func() {
		Host(nil, func(notation *utils.ExecOpts, _ *Artifact, vhost *utils.VirtualHost) {
			vhost.SetOption(AuthOption("", ""),
				AddTrustPolicyOption("permissive_trustpolicy.json"),
				AddTrustStoreOption("e2e", filepath.Join(NotationE2ELocalKeysDir, "new_e2e.crt")))

			// the artifact signed with a different cert from the cert in
			// trust store.
			artifact := GenerateArtifact("e2e-valid-signature", "")

			notation.ExpectFailure().Exec("verify", artifact.ReferenceWithDigest(), "-v").
				MatchErrKeyWords("authenticity validation failed",
					VerifyFailed)
		})
	})

	It("permissive level with invalid integrity", func() {
		Host(BaseOptions(), func(notation *utils.ExecOpts, _ *Artifact, vhost *utils.VirtualHost) {
			vhost.SetOption(AddTrustPolicyOption("permissive_trustpolicy.json"))

			artifact := GenerateArtifact("e2e-invalid-signature", "")

			notation.ExpectFailure().Exec("verify", artifact.ReferenceWithDigest(), "-v").
				MatchErrKeyWords("integrity validation failed",
					VerifyFailed)
		})
	})

	It("audit level with expired signature", func() {
		Host(BaseOptions(), func(notation *utils.ExecOpts, _ *Artifact, vhost *utils.VirtualHost) {
			vhost.SetOption(AddTrustPolicyOption("audit_trustpolicy.json"))

			artifact := GenerateArtifact("e2e-expired-signature", "")

			notation.Exec("verify", artifact.ReferenceWithDigest(), "-v").
				MatchErrKeyWords("digital signature has expired",
					"expiry was set to \"log\"").
				MatchKeyWords(VerifySuccessfully)
		})
	})

	It("audit level with expired authentic timestamp", func() {
		Host(nil, func(notation *utils.ExecOpts, _ *Artifact, vhost *utils.VirtualHost) {
			artifact := GenerateArtifact("e2e-with-expired-cert", "")

			vhost.SetOption(AuthOption("", ""),
				AddTrustPolicyOption("audit_trustpolicy.json"),
				AddTrustStoreOption("e2e", filepath.Join(NotationE2EConfigPath, "localkeys", "expired_e2e.crt")))

			notation.Exec("verify", artifact.ReferenceWithDigest(), "-v").
				MatchErrKeyWords("Warning: authenticTimestamp was set to \"log\"",
					"error: certificate \"O=Internet Widgits Pty Ltd,ST=Some-State,C=AU\" is not valid anymore, it was expired").
				MatchKeyWords(VerifySuccessfully)
		})
	})

	It("audit level with invalid authenticity", func() {
		Host(nil, func(notation *utils.ExecOpts, _ *Artifact, vhost *utils.VirtualHost) {
			vhost.SetOption(AuthOption("", ""),
				AddTrustPolicyOption("audit_trustpolicy.json"),
				AddTrustStoreOption("e2e", filepath.Join(NotationE2ELocalKeysDir, "new_e2e.crt")))

			// the artifact signed with a different cert from the cert in
			// trust store.
			artifact := GenerateArtifact("e2e-valid-signature", "")

			notation.Exec("verify", artifact.ReferenceWithDigest(), "-v").
				MatchErrKeyWords("Warning: authenticity was set to \"log\"",
					"signature is not produced by a trusted signer").
				MatchKeyWords(VerifySuccessfully)
		})
	})

	It("audit level with invalid integrity", func() {
		Host(BaseOptions(), func(notation *utils.ExecOpts, _ *Artifact, vhost *utils.VirtualHost) {
			vhost.SetOption(AddTrustPolicyOption("audit_trustpolicy.json"))

			artifact := GenerateArtifact("e2e-invalid-signature", "")

			notation.ExpectFailure().Exec("verify", artifact.ReferenceWithDigest(), "-v").
				MatchErrKeyWords("integrity validation failed",
					VerifyFailed)
		})
	})

	It("skip level with invalid integrity", func() {
		Host(BaseOptions(), func(notation *utils.ExecOpts, _ *Artifact, vhost *utils.VirtualHost) {
			vhost.SetOption(AddTrustPolicyOption("skip_trustpolicy.json"))

			artifact := GenerateArtifact("e2e-invalid-signature", "")

			notation.Exec("verify", artifact.ReferenceWithDigest(), "-v").
				MatchKeyWords("Trust policy is configured to skip signature verification")
		})
	})

	It("strict level with Expiry overridden as log level", func() {
		Host(BaseOptions(), func(notation *utils.ExecOpts, _ *Artifact, vhost *utils.VirtualHost) {
			vhost.SetOption(AddTrustPolicyOption("override_strict_trustpolicy.json"))

			artifact := GenerateArtifact("e2e-expired-signature", "")

			notation.Exec("verify", artifact.ReferenceWithDigest(), "-v").
				MatchErrKeyWords("digital signature has expired",
					"expiry was set to \"log\"").
				MatchKeyWords(VerifySuccessfully)
		})
	})

	It("strict level with Authentic timestamp overridden as log level", func() {
		Host(nil, func(notation *utils.ExecOpts, _ *Artifact, vhost *utils.VirtualHost) {
			artifact := GenerateArtifact("e2e-with-expired-cert", "")

			vhost.SetOption(AuthOption("", ""),
				AddTrustPolicyOption("override_strict_trustpolicy.json"),
				AddTrustStoreOption("e2e", filepath.Join(NotationE2EConfigPath, "localkeys", "expired_e2e.crt")))

			notation.Exec("verify", artifact.ReferenceWithDigest(), "-v").
				MatchErrKeyWords("Warning: authenticTimestamp was set to \"log\"",
					"error: certificate \"O=Internet Widgits Pty Ltd,ST=Some-State,C=AU\" is not valid anymore, it was expired").
				MatchKeyWords(VerifySuccessfully)
		})
	})

	It("strict level with Authenticity overridden as log level", func() {
		Host(nil, func(notation *utils.ExecOpts, _ *Artifact, vhost *utils.VirtualHost) {
			vhost.SetOption(AuthOption("", ""),
				AddTrustPolicyOption("override_strict_trustpolicy.json"),
				AddTrustStoreOption("e2e", filepath.Join(NotationE2ELocalKeysDir, "new_e2e.crt")))
			// the artifact signed with a different cert from the cert in
			// trust store.
			artifact := GenerateArtifact("e2e-valid-signature", "")

			notation.Exec("verify", artifact.ReferenceWithDigest(), "-v").
				MatchErrKeyWords("Warning: authenticity was set to \"log\"",
					"signature is not produced by a trusted signer").
				MatchKeyWords(VerifySuccessfully)
		})
	})

	It("permissive level with Expiry overridden as enforce level", func() {
		Host(BaseOptions(), func(notation *utils.ExecOpts, _ *Artifact, vhost *utils.VirtualHost) {
			vhost.SetOption(AddTrustPolicyOption("override_permissive_trustpolicy.json"))

			artifact := GenerateArtifact("e2e-expired-signature", "")

			notation.ExpectFailure().Exec("verify", artifact.ReferenceWithDigest(), "-v").
				MatchErrKeyWords("expiry validation failed.",
					VerifyFailed)
		})
	})

	It("permissive level with Authentic timestamp overridden as enforce level", func() {
		Host(nil, func(notation *utils.ExecOpts, _ *Artifact, vhost *utils.VirtualHost) {
			vhost.SetOption(AddTrustPolicyOption("override_permissive_trustpolicy.json"))

			artifact := GenerateArtifact("e2e-with-expired-cert", "")

			vhost.SetOption(AuthOption("", ""),
				AddTrustPolicyOption("trustpolicy.json"),
				AddTrustStoreOption("e2e", filepath.Join(NotationE2EConfigPath, "localkeys", "expired_e2e.crt")))

			notation.ExpectFailure().Exec("verify", artifact.ReferenceWithDigest(), "-v").
				MatchErrKeyWords("authenticTimestamp validation failed",
					VerifyFailed)
		})
	})

	It("permissive level with Authenticity overridden as log level", func() {
		Host(nil, func(notation *utils.ExecOpts, _ *Artifact, vhost *utils.VirtualHost) {
			vhost.SetOption(AuthOption("", ""),
				AddTrustPolicyOption("override_permissive_trustpolicy.json"),
				AddTrustStoreOption("e2e", filepath.Join(NotationE2ELocalKeysDir, "new_e2e.crt")))

			artifact := GenerateArtifact("e2e-valid-signature", "")

			notation.Exec("verify", artifact.ReferenceWithDigest(), "-v").
				MatchErrKeyWords("Warning: authenticity was set to \"log\"",
					"signature is not produced by a trusted signer").
				MatchKeyWords(VerifySuccessfully)
		})
	})

	It("permissive level with Integrity overridden as log level", func() {
		Host(nil, func(notation *utils.ExecOpts, _ *Artifact, vhost *utils.VirtualHost) {
			vhost.SetOption(AuthOption("", ""),
				AddTrustPolicyOption("override_integrity_for_permissive_trustpolicy.json"),
				AddTrustStoreOption("e2e", filepath.Join(NotationE2ELocalKeysDir, "new_e2e.crt")))

			artifact := GenerateArtifact("e2e-valid-signature", "")

			notation.ExpectFailure().Exec("verify", artifact.ReferenceWithDigest(), "-v").
				MatchErrKeyWords(`"integrity" verification can not be overridden in custom signature verification`)
		})
	})

	It("audit level with Expiry overridden as enforce level", func() {
		Host(BaseOptions(), func(notation *utils.ExecOpts, _ *Artifact, vhost *utils.VirtualHost) {
			vhost.SetOption(AddTrustPolicyOption("override_audit_trustpolicy.json"))

			artifact := GenerateArtifact("e2e-expired-signature", "")

			notation.ExpectFailure().Exec("verify", artifact.ReferenceWithDigest(), "-v").
				MatchErrKeyWords("expiry validation failed.",
					VerifyFailed)
		})
	})

	It("audit level with Authentic timestamp overridden as enforce level", func() {
		Host(nil, func(notation *utils.ExecOpts, _ *Artifact, vhost *utils.VirtualHost) {
			vhost.SetOption(AddTrustPolicyOption("override_audit_trustpolicy.json"))

			artifact := GenerateArtifact("e2e-with-expired-cert", "")

			vhost.SetOption(AuthOption("", ""),
				AddTrustPolicyOption("trustpolicy.json"),
				AddTrustStoreOption("e2e", filepath.Join(NotationE2EConfigPath, "localkeys", "expired_e2e.crt")))

			notation.ExpectFailure().Exec("verify", artifact.ReferenceWithDigest(), "-v").
				MatchErrKeyWords("authenticTimestamp validation failed",
					VerifyFailed)
		})
	})

	It("audit level with Authenticity overridden as enforce level", func() {
		Host(nil, func(notation *utils.ExecOpts, _ *Artifact, vhost *utils.VirtualHost) {
			vhost.SetOption(AuthOption("", ""),
				AddTrustPolicyOption("override_audit_trustpolicy.json"),
				AddTrustStoreOption("e2e", filepath.Join(NotationE2ELocalKeysDir, "new_e2e.crt")))

			// the artifact signed with a different cert from the cert in
			// trust store.
			artifact := GenerateArtifact("e2e-valid-signature", "")

			notation.ExpectFailure().Exec("verify", artifact.ReferenceWithDigest(), "-v").
				MatchErrKeyWords("authenticity validation failed",
					VerifyFailed)
		})
	})
})
