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
	HTTPRequest = fmt.Sprintf("http://%s", TestRegistry.DomainHost)

	// HTTPSRequest is the base URL for HTTPS requests for testing TLS request.
	HTTPSRequest = fmt.Sprintf("https://%s", TestRegistry.DomainHost)
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
