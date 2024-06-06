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
	"encoding/base64"
	"fmt"

	"github.com/golang-jwt/jwt"
	"github.com/notaryproject/notation-core-go/signature"
	"github.com/notaryproject/notation-go/plugin/proto"
	"github.com/notaryproject/notation/test/e2e/plugin/internal/io"
	"github.com/notaryproject/notation/test/e2e/plugin/mock"
	"github.com/spf13/cobra"
)

var (
	ps256 = jwt.SigningMethodPS256.Name
	ps384 = jwt.SigningMethodPS384.Name
	ps512 = jwt.SigningMethodPS512.Name
	es256 = jwt.SigningMethodES256.Name
	es384 = jwt.SigningMethodES384.Name
	es512 = jwt.SigningMethodES512.Name
)

var validMethods = []string{ps256, ps384, ps512, es256, es384, es512}

var signatureAlgJWSAlgMap = map[signature.Algorithm]string{
	signature.AlgorithmPS256: ps256,
	signature.AlgorithmPS384: ps384,
	signature.AlgorithmPS512: ps512,
	signature.AlgorithmES256: es256,
	signature.AlgorithmES384: es384,
	signature.AlgorithmES512: es512,
}

func generateSignatureCommand() *cobra.Command {
	return &cobra.Command{
		Use: "generate-signature",
		RunE: func(cmd *cobra.Command, args []string) error {
			req := &proto.GenerateSignatureRequest{}
			if err := io.UnmarshalRequest(req); err != nil {
				return &proto.RequestError{Code: proto.ErrorCodeValidation, Err: err}
			}

			// validate request
			if err := validateGenerateSignatureRequest(*req); err != nil {
				return &proto.RequestError{Code: proto.ErrorCodeValidation, Err: err}
			}
			return runGenerateSignature(req)
		},
	}
}

func runGenerateSignature(req *proto.GenerateSignatureRequest) error {
	// load cert chain by keyID
	privateKey, certs, err := loadCertChain(req.KeyID)
	if err != nil {
		return &proto.RequestError{Code: ErrorCodeInvalidCertificate, Err: err}
	}

	// prepare to sign
	algorithm, err := extractAlgorithm(req)
	if err != nil {
		return &proto.RequestError{Code: proto.ErrorCodeValidation, Err: err}
	}

	rawSig, err := sign(string(req.Payload), privateKey, algorithm)
	if err != nil {
		return &proto.RequestError{Code: proto.ErrorCodeGeneric, Err: err}
	}

	// prepare response
	signingAlg, err := proto.EncodeSigningAlgorithm(algorithm)
	if err != nil {
		return &proto.RequestError{Code: proto.ErrorCodeGeneric, Err: err}
	}
	resp := &proto.GenerateSignatureResponse{
		KeyID:            req.KeyID,
		Signature:        rawSig,
		SigningAlgorithm: string(signingAlg),
		CertificateChain: toRawCerts(certs),
	}

	// update response for testing various cases
	updateGenerateSignatureResponse(req, resp)
	return io.PrintResponse(resp)
}

func sign(payload string, privateKey crypto.PrivateKey, algorithm signature.Algorithm) ([]byte, error) {
	jwtAlg, err := toJWTAlgorithm(algorithm)
	if err != nil {
		return nil, err
	}

	// use JWT package to sign raw signature.
	method := jwt.GetSigningMethod(jwtAlg)
	sig, err := method.Sign(payload, privateKey)
	if err != nil {
		return nil, err
	}

	return base64.RawURLEncoding.DecodeString(sig)
}

func toRawCerts(certs []*x509.Certificate) [][]byte {
	var rawCerts [][]byte
	for _, cert := range certs {
		rawCerts = append(rawCerts, cert.Raw)
	}
	return rawCerts
}

func extractAlgorithm(req *proto.GenerateSignatureRequest) (signature.Algorithm, error) {
	// extract algorithm from signer
	keySpec, err := proto.DecodeKeySpec(req.KeySpec)
	if err != nil {
		return -1, err
	}
	return keySpec.SignatureAlgorithm(), nil
}

func toJWTAlgorithm(alg signature.Algorithm) (string, error) {
	// converts the signature.Algorithm to be jwt package defined
	// algorithm name.
	jwsAlg, ok := signatureAlgJWSAlgMap[alg]
	if !ok {
		return "", &signature.UnsupportedSignatureAlgoError{
			Alg: fmt.Sprintf("#%d", alg)}
	}
	return jwsAlg, nil
}

// validateGenerateSignatureRequest validates required field existence.
func validateGenerateSignatureRequest(req proto.GenerateSignatureRequest) error {
	return validateRequiredField(req, fieldSet(
		"ContractVersion",
		"KeyID",
		"KeySpec",
		"Hash",
		"Payload"))
}

// updateGenerateSignatureResponse tampers the response to test various cases.
func updateGenerateSignatureResponse(req *proto.GenerateSignatureRequest, resp *proto.GenerateSignatureResponse) {
	if v, ok := req.PluginConfig[mock.TamperKeyID]; ok {
		resp.KeyID = v
	}

	if v, ok := req.PluginConfig[mock.TamperSignature]; ok {
		resp.Signature = []byte(v)
	}

	if v, ok := req.PluginConfig[mock.TamperSignatureAlgorithm]; ok {
		resp.SigningAlgorithm = v
	}

	if v, ok := req.PluginConfig[mock.TamperCertificateChain]; ok {
		resp.CertificateChain = [][]byte{[]byte(v)}
	}
}
