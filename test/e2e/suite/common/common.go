package common

import (
	"fmt"

	. "github.com/notaryproject/notation/test/e2e/internal/notation"
)

const (
	LoginSuccessfully  = "Login Succeeded"
	LogoutSuccessfully = "Logout Succeeded"
	SignSuccessfully   = "Successfully signed"
	VerifySuccessfully = "Successfully verified"
	VerifyFailed       = "signature verification failed"
)

var (
	// HTTPRequest is the base URL for HTTP requests for testing
	// --insecure-registry flag
	HTTPRequest = fmt.Sprintf("http://%s", RegistryWithDomainHost.Host)

	// HTTPSRequest is the base URL for HTTPS requests for testing TLS request.
	HTTPSRequest = fmt.Sprintf("https://%s", RegistryWithDomainHost.Host)
)

const (
	// HeaderVerificationPlugin specifies the name of the verification plugin that should be used to verify the signature.
	HeaderVerificationPlugin = "io.cncf.notary.verificationPlugin"

	// HeaderVerificationPluginMinVersion specifies the minimum version of the verification plugin that should be used to verify the signature.
	HeaderVerificationPluginMinVersion = "io.cncf.notary.verificationPluginMinVersion"
)

// Capability is a feature available in the plugin contract.
type Capability string

const (
	// CapabilitySignatureGenerator is the name of the capability
	// for a plugin to support generating raw signatures.
	CapabilitySignatureGenerator Capability = "SIGNATURE_GENERATOR.RAW"

	// CapabilityEnvelopeGenerator is the name of the capability
	// for a plugin to support generating envelope signatures.
	CapabilityEnvelopeGenerator Capability = "SIGNATURE_GENERATOR.ENVELOPE"

	// CapabilityTrustedIdentityVerifier is the name of the
	// capability for a plugin to support verifying trusted identities.
	CapabilityTrustedIdentityVerifier Capability = "SIGNATURE_VERIFIER.TRUSTED_IDENTITY"

	// CapabilityRevocationCheckVerifier is the name of the
	// capability for a plugin to support verifying revocation checks.
	CapabilityRevocationCheckVerifier Capability = "SIGNATURE_VERIFIER.REVOCATION_CHECK"
)
