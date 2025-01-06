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

package cmd

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

// GetVerifier returns a Verifier.
// isBlob is set to true when verifying an arbitrary blob.
func GetVerifier(ctx context.Context, isBlob bool) (Verifier, error) {
	// revocation check
	revocationCodeSigningValidator, err := clirev.NewRevocationValidator(ctx, purpose.CodeSigning)
	if err != nil {
		return nil, err
	}
	revocationTimestampingValidator, err := clirev.NewRevocationValidator(ctx, purpose.Timestamping)
	if err != nil {
		return nil, err
	}

	// trust policy and trust store
	x509TrustStore := truststore.NewX509TrustStore(dir.ConfigFS())
	if isBlob {
		blobPolicyDocument, err := trustpolicy.LoadBlobDocument()
		if err != nil {
			return nil, err
		}
		return verifier.NewVerifierWithOptions(nil, blobPolicyDocument, x509TrustStore, plugin.NewCLIManager(dir.PluginFS()), verifier.VerifierOptions{
			RevocationCodeSigningValidator:  revocationCodeSigningValidator,
			RevocationTimestampingValidator: revocationTimestampingValidator,
		})
	}

	policyDocument, err := trustpolicy.LoadOCIDocument()
	if err != nil {
		return nil, err
	}
	return verifier.NewVerifierWithOptions(policyDocument, nil, x509TrustStore, plugin.NewCLIManager(dir.PluginFS()), verifier.VerifierOptions{
		RevocationCodeSigningValidator:  revocationCodeSigningValidator,
		RevocationTimestampingValidator: revocationTimestampingValidator,
	})
}
