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

// Package verify provides utility methods related to verification commands.
package verify

import (
	"context"
	"errors"
	"fmt"
	"io/fs"

	"github.com/notaryproject/notation-core-go/revocation/purpose"
	"github.com/notaryproject/notation-go"
	"github.com/notaryproject/notation-go/dir"
	"github.com/notaryproject/notation-go/plugin"
	"github.com/notaryproject/notation-go/verifier"
	"github.com/notaryproject/notation-go/verifier/trustpolicy"
	"github.com/notaryproject/notation-go/verifier/truststore"

	clirev "github.com/notaryproject/notation/v2/internal/revocation"
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

// ComposeVerificationFailurePrintout composes verification failure print out.
func ComposeVerificationFailurePrintout(outcomes []*notation.VerificationOutcome, reference string, err error) error {
	if verificationErr := parseErrorOnVerificationFailure(err); verificationErr != nil {
		return verificationErr
	}
	if len(outcomes) == 0 {
		return fmt.Errorf("signature verification failed for all the signatures associated with %s", reference)
	}
	return nil
}

// ComposeBlobVerificationFailurePrintout composes blob verification failure
// print out.
func ComposeBlobVerificationFailurePrintout(outcomes []*notation.VerificationOutcome, blobPath string, err error) error {
	if verificationErr := parseErrorOnVerificationFailure(err); verificationErr != nil {
		return verificationErr
	}
	if len(outcomes) == 0 {
		return fmt.Errorf("provided signature verification failed against blob %s", blobPath)
	}
	return nil
}

// parseErrorOnVerificationFailure parses error on verification failure.
func parseErrorOnVerificationFailure(err error) error {
	if err == nil {
		return nil
	}

	var errTrustStore truststore.TrustStoreError
	if errors.As(err, &errTrustStore) {
		if errors.Is(err, fs.ErrNotExist) {
			return fmt.Errorf("%w. Use command 'notation cert add' to create and add trusted certificates to the trust store", errTrustStore)
		} else {
			return fmt.Errorf("%w. %w", errTrustStore, errTrustStore.InnerError)
		}
	}

	var errCertificate truststore.CertificateError
	if errors.As(err, &errCertificate) {
		if errors.Is(err, fs.ErrNotExist) {
			return fmt.Errorf("%w. Use command 'notation cert add' to create and add trusted certificates to the trust store", errCertificate)
		} else {
			return fmt.Errorf("%w. %w", errCertificate, errCertificate.InnerError)
		}
	}

	var errorVerificationFailed notation.VerificationFailedError
	if !errors.As(err, &errorVerificationFailed) {
		return fmt.Errorf("signature verification failed: %w", err)
	}
	return nil
}
