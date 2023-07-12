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

package main

import (
	"github.com/notaryproject/notation-go/plugin/proto"
	"github.com/notaryproject/notation/test/e2e/plugin/internal/io"
	"github.com/spf13/cobra"
)

func verifySignatureCommand() *cobra.Command {
	return &cobra.Command{
		Use: "verify-signature",
		RunE: func(cmd *cobra.Command, args []string) error {
			req := &proto.VerifySignatureRequest{}
			if err := io.UnmarshalRequest(req); err != nil {
				return &proto.RequestError{Code: proto.ErrorCodeValidation, Err: err}
			}

			if err := validateVerifySignatureRequest(*req); err != nil {
				return &proto.RequestError{Code: proto.ErrorCodeValidation, Err: err}
			}
			return runVerifySignature(req)
		},
	}
}

func runVerifySignature(req *proto.VerifySignatureRequest) error {
	return io.PrintResponse(extractVerificationResult(req))
}

// validateVerifySignatureRequest validates required field existence.
func validateVerifySignatureRequest(req proto.VerifySignatureRequest) error {
	// check req.Signature.CriticalAttributes
	if err := validateRequiredField(req.Signature.CriticalAttributes, fieldSet(
		"ContentType",
		"SigningScheme",
		"Expiry")); err != nil {
		return err
	}

	// check req.Signature
	if err := validateRequiredField(req.Signature, fieldSet("CertificateChain")); err != nil {
		return err
	}

	// check req.TrustPolicy
	if err := validateRequiredField(req.TrustPolicy, fieldSet(
		"TrustIdentities",
		"SignatureVerification")); err != nil {
		return err
	}

	return validateRequiredField(req, fieldSet("ContractVersion"))
}

func extractVerificationResult(req *proto.VerifySignatureRequest) *proto.VerifySignatureResponse {
	resp := &proto.VerifySignatureResponse{
		VerificationResults: make(map[proto.Capability]*proto.VerificationResult),
	}

	// set verification result based on req.PluginConfig
	if v, ok := req.PluginConfig[string(proto.CapabilityRevocationCheckVerifier)]; !ok || v == "success" {
		resp.VerificationResults[proto.CapabilityRevocationCheckVerifier] = &proto.VerificationResult{
			Success: true,
		}
	} else {
		resp.VerificationResults[proto.CapabilityRevocationCheckVerifier] = &proto.VerificationResult{
			Success: false,
			Reason:  "revocation check failed",
		}
	}
	if v, ok := req.PluginConfig[string(proto.CapabilityTrustedIdentityVerifier)]; !ok || v == "success" {
		resp.VerificationResults[proto.CapabilityTrustedIdentityVerifier] = &proto.VerificationResult{
			Success: true,
		}
	} else {
		resp.VerificationResults[proto.CapabilityTrustedIdentityVerifier] = &proto.VerificationResult{
			Success: false,
			Reason:  "trusted identity check failed",
		}
	}

	return resp
}
