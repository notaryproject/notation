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
	"crypto"
	"crypto/x509"
	"errors"
	"strings"
	"time"

	"github.com/notaryproject/notation-core-go/signature"
	_ "github.com/notaryproject/notation-core-go/signature/cose"
	_ "github.com/notaryproject/notation-core-go/signature/jws"
	"github.com/notaryproject/notation-go/plugin/proto"
	"github.com/notaryproject/notation-go/verifier"
	"github.com/notaryproject/notation/test/e2e/plugin/internal/io"
	"github.com/notaryproject/notation/test/e2e/plugin/mock"
	"github.com/spf13/cobra"
)

const MediaTypePayloadV1 = "application/vnd.cncf.notary.payload.v1+json"

func generateEnvelopeCommand() *cobra.Command {
	return &cobra.Command{
		Use: "generate-envelope",
		RunE: func(cmd *cobra.Command, args []string) error {
			req := &proto.GenerateEnvelopeRequest{}
			if err := io.UnmarshalRequest(req); err != nil {
				return &proto.RequestError{Code: proto.ErrorCodeValidation, Err: err}
			}

			// validate request
			if err := validateGenerateEnvelopeRequest(*req); err != nil {
				return &proto.RequestError{Code: proto.ErrorCodeValidation, Err: err}
			}
			return runGenerateEnvelopeCommand(req)
		},
	}
}

func runGenerateEnvelopeCommand(req *proto.GenerateEnvelopeRequest) error {
	// load cert chain by keyID
	privateKey, certs, err := loadCertChain(req.KeyID)
	if err != nil {
		return &proto.RequestError{Code: ErrorCodeInvalidCertificate, Err: err}
	}

	// prepare to sign
	signer, err := newSigner(privateKey, certs)
	if err != nil {
		return &proto.RequestError{Code: ErrorCodeSigningError, Err: err}
	}

	opts := extractSigningOpts(req)
	sig, err := signer.sign(req.Payload, req.SignatureEnvelopeType, opts)
	if err != nil {
		return &proto.RequestError{Code: ErrorCodeSigningError, Err: err}
	}

	// prepare response
	resp := &proto.GenerateEnvelopeResponse{
		SignatureEnvelope:     sig,
		SignatureEnvelopeType: req.SignatureEnvelopeType,
		Annotations:           map[string]string{"signer": "e2e-plugin"},
	}

	// update response for testing various cases
	if err := updateGenerateEnvelopeResponse(req, resp); err != nil {
		return err
	}

	return io.PrintResponse(resp)
}

// validateGenerateEnvelopeRequest validates required failed existence.
func validateGenerateEnvelopeRequest(req proto.GenerateEnvelopeRequest) error {
	return validateRequiredField(req, fieldSet(
		"ContractVersion",
		"KeyID",
		"PayloadType",
		"SignatureEnvelopeType",
		"Payload",
	))
}

// signer uses notation-core-go to sign the payload.
type signer struct {
	signature.Signer
}

func newSigner(key crypto.PrivateKey, certChain []*x509.Certificate) (*signer, error) {
	localSigner, err := signature.NewLocalSigner(certChain, key)
	if err != nil {
		return nil, err
	}
	return &signer{Signer: localSigner}, nil
}

func (s *signer) sign(payload []byte, signatureMediaType string, opts *signingOpts) ([]byte, error) {
	signReq := &signature.SignRequest{
		Payload: signature.Payload{
			ContentType: MediaTypePayloadV1,
			Content:     payload,
		},
		Signer:                   s.Signer,
		SigningTime:              time.Now(),
		SigningScheme:            signature.SigningSchemeX509,
		SigningAgent:             "e2e-plugin",
		ExtendedSignedAttributes: opts.extendedAttribute,
	}

	// Add expiry only if ExpiryDuration is not zero
	if opts.expiryDuration != 0 {
		signReq.Expiry = signReq.SigningTime.Add(opts.expiryDuration)
	}

	// perform signing
	sigEnv, err := signature.NewEnvelope(signatureMediaType)
	if err != nil {
		return nil, err
	}

	return sigEnv.Sign(signReq)
}

type signingOpts struct {
	extendedAttribute []signature.Attribute
	expiryDuration    time.Duration
}

func extractSigningOpts(req *proto.GenerateEnvelopeRequest) *signingOpts {
	opts := &signingOpts{}

	// update expiry duration
	if req.ExpiryDurationInSeconds != 0 {
		opts.expiryDuration = time.Duration(req.ExpiryDurationInSeconds) * time.Second
	}

	// update extended Attributes
	if v, ok := req.PluginConfig[verifier.HeaderVerificationPlugin]; ok {
		opts.extendedAttribute = append(opts.extendedAttribute, signature.Attribute{
			Key:      verifier.HeaderVerificationPlugin,
			Critical: true,
			Value:    v,
		})
	}
	if v, ok := req.PluginConfig[verifier.HeaderVerificationPluginMinVersion]; ok {
		opts.extendedAttribute = append(opts.extendedAttribute, signature.Attribute{
			Key:      verifier.HeaderVerificationPluginMinVersion,
			Critical: true,
			Value:    v,
		})
	}

	return opts
}

// updateGenerateEnvelopeResponse tampers the response to mock various cases.
func updateGenerateEnvelopeResponse(req *proto.GenerateEnvelopeRequest, resp *proto.GenerateEnvelopeResponse) error {
	if v, ok := req.PluginConfig[mock.TamperSignatureEnvelope]; ok {
		resp.SignatureEnvelope = []byte(v)
	}

	if v, ok := req.PluginConfig[mock.TamperSignatureEnvelopeType]; ok {
		resp.SignatureEnvelopeType = v
	}

	if v, ok := req.PluginConfig[mock.TamperAnnotation]; ok {
		kv := strings.Split(v, "=")
		if len(kv) != 2 {
			return errors.New("invalid annotation")
		}
		resp.Annotations = map[string]string{
			kv[0]: kv[1],
		}
	}
	return nil
}
