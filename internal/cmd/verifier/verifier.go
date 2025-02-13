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

package verifier

import (
	"context"

	"github.com/notaryproject/notation-core-go/revocation/purpose"
	"github.com/notaryproject/notation-go"
	"github.com/notaryproject/notation-go/dir"
	"github.com/notaryproject/notation-go/plugin"
	"github.com/notaryproject/notation-go/verifier"
	"github.com/notaryproject/notation-go/verifier/trustpolicy"
	"github.com/notaryproject/notation-go/verifier/truststore"

	clirev "github.com/notaryproject/notation/internal/revocation"
)

// Verifier is embedded with notation.BlobVerifier and notation.Verifier.
type Verifier interface {
	notation.BlobVerifier
	notation.Verifier
}

// GetVerifier creates a Verifier.
func GetVerifier(ctx context.Context) (Verifier, error) {
	verifierOptions, err := newVerifierOptions(ctx)
	if err != nil {
		return nil, err
	}

	// trust policy and trust store
	x509TrustStore := truststore.NewX509TrustStore(dir.ConfigFS())
	policyDocument, err := trustpolicy.LoadOCIDocument()
	if err != nil {
		return nil, err
	}
	verifierOptions.OCITrustPolicy = policyDocument
	return verifier.NewVerifierWithOptions(x509TrustStore, verifierOptions)
}

// GetBlobVerifier creates a BlobVerifier.
func GetBlobVerifier(ctx context.Context) (Verifier, error) {
	verifierOptions, err := newVerifierOptions(ctx)
	if err != nil {
		return nil, err
	}

	// trust policy and trust store
	x509TrustStore := truststore.NewX509TrustStore(dir.ConfigFS())
	blobPolicyDocument, err := trustpolicy.LoadBlobDocument()
	if err != nil {
		return nil, err
	}
	verifierOptions.BlobTrustPolicy = blobPolicyDocument
	return verifier.NewVerifierWithOptions(x509TrustStore, verifierOptions)
}

// newVerifierOptions creates a verifier.VerifierOptions.
func newVerifierOptions(ctx context.Context) (verifier.VerifierOptions, error) {
	revocationCodeSigningValidator, err := clirev.NewRevocationValidator(ctx, purpose.CodeSigning)
	if err != nil {
		return verifier.VerifierOptions{}, err
	}
	revocationTimestampingValidator, err := clirev.NewRevocationValidator(ctx, purpose.Timestamping)
	if err != nil {
		return verifier.VerifierOptions{}, err
	}
	return verifier.VerifierOptions{
		RevocationCodeSigningValidator:  revocationCodeSigningValidator,
		RevocationTimestampingValidator: revocationTimestampingValidator,
		PluginManager:                   plugin.NewCLIManager(dir.PluginFS()),
	}, nil
}
