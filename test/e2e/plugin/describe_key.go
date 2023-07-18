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
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"

	"github.com/notaryproject/notation-core-go/signature"
	"github.com/notaryproject/notation-go/config"
	"github.com/notaryproject/notation-go/dir"
	"github.com/notaryproject/notation-go/plugin/proto"
	"github.com/notaryproject/notation/test/e2e/plugin/internal/io"
	"github.com/spf13/cobra"
)

func describeKeyCommand() *cobra.Command {
	return &cobra.Command{
		Use: "describe-key",
		RunE: func(cmd *cobra.Command, args []string) error {
			req := &proto.DescribeKeyRequest{}
			if err := io.UnmarshalRequest(req); err != nil {
				return &proto.RequestError{Code: proto.ErrorCodeValidation, Err: err}
			}

			if err := validateDescribeKeyRequest(*req); err != nil {
				return &proto.RequestError{Code: proto.ErrorCodeValidation, Err: err}
			}
			return runDescribeKey(req)
		},
	}
}

func runDescribeKey(req *proto.DescribeKeyRequest) error {
	// load cert chain by keyID
	_, certs, err := loadCertChain(req.KeyID)
	if err != nil {
		return &proto.RequestError{Code: ErrorCodeInvalidCertificate, Err: err}
	}

	// prepare response
	ks, err := signature.ExtractKeySpec(certs[0])
	if err != nil {
		return &proto.RequestError{Code: ErrorCodeInvalidCertificate, Err: err}
	}

	keySpec, err := proto.EncodeKeySpec(ks)
	if err != nil {
		return &proto.RequestError{Code: proto.ErrorCodeGeneric, Err: err}
	}

	return io.PrintResponse(&proto.DescribeKeyResponse{
		KeyID:   req.KeyID,
		KeySpec: keySpec,
	})
}

func loadCertChain(keyID string) (crypto.PrivateKey, []*x509.Certificate, error) {
	keyPath, certPath, err := getKeyPairPath(keyID)
	if err != nil {
		return nil, nil, err
	}
	certs, err := tls.LoadX509KeyPair(certPath, keyPath)
	if err != nil {
		return nil, nil, err
	}
	if len(certs.Certificate) == 0 {
		return nil, nil, fmt.Errorf("%q does not contain certificate", certPath)
	}

	var certChain []*x509.Certificate
	for _, cert := range certs.Certificate {
		c, err := x509.ParseCertificate(cert)
		if err != nil {
			return nil, nil, err
		}
		certChain = append(certChain, c)
	}
	return certs.PrivateKey, certChain, nil
}

func getKeyPairPath(keyId string) (string, string, error) {
	cfg := &config.SigningKeys{}
	file, err := dir.ConfigFS().Open("pluginkeys.json")
	if err != nil {
		return "", "", err
	}
	defer file.Close()

	if err := json.NewDecoder(file).Decode(cfg); err != nil {
		return "", "", err
	}

	var keyPath, certPath string
	for _, key := range cfg.Keys {
		if key.ID == keyId {
			keyPath = key.KeyPath
			certPath = key.CertificatePath
			break
		}
	}
	if keyPath == "" || certPath == "" {
		return "", "", fmt.Errorf("keyPath or certPath did't find for keyId: %s", keyId)
	}
	return keyPath, certPath, nil
}

func validateDescribeKeyRequest(req proto.DescribeKeyRequest) error {
	return validateRequiredField(req, fieldSet("ContractVesion", "KeyId"))
}
