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

package sharedutils

import (
	"errors"
	"fmt"
	"github.com/notaryproject/notation-go"
	"github.com/notaryproject/notation-go/verifier/trustpolicy"
	"github.com/notaryproject/notation-go/verifier/truststore"
	"github.com/notaryproject/notation/internal/ioutil"
	"io/fs"
	"os"
	"reflect"
)

func CheckVerificationFailure(outcomes []*notation.VerificationOutcome, printOut string, err error) error {
	// write out on failure
	if err != nil || len(outcomes) == 0 {
		if err != nil {
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

			var errorVerificationFailed notation.ErrorVerificationFailed
			if !errors.As(err, &errorVerificationFailed) {
				return fmt.Errorf("signature verification failed: %w", err)
			}
		}
		return fmt.Errorf("signature verification failed for all the signatures associated with %s", printOut)
	}
	return nil
}

func ReportVerificationSuccess(outcomes []*notation.VerificationOutcome, printout string) {
	// write out on success
	outcome := outcomes[0]
	// print out warning for any failed result with logged verification action
	for _, result := range outcome.VerificationResults {
		if result.Error != nil {
			// at this point, the verification action has to be logged and
			// it's failed
			fmt.Fprintf(os.Stderr, "Warning: %v was set to %q and failed with error: %v\n", result.Type, result.Action, result.Error)
		}
	}
	if reflect.DeepEqual(outcome.VerificationLevel, trustpolicy.LevelSkip) {
		fmt.Println("Trust policy is configured to skip signature verification for", printout)
	} else {
		fmt.Println("Successfully verified signature for", printout)
		printMetadataIfPresent(outcome)
	}
}

func printMetadataIfPresent(outcome *notation.VerificationOutcome) {
	// the signature envelope is parsed as part of verification.
	// since user metadata is only printed on successful verification,
	// this error can be ignored
	metadata, _ := outcome.UserMetadata()

	if len(metadata) > 0 {
		fmt.Println("\nThe artifact was signed with the following user metadata.")
		ioutil.PrintMetadataMap(os.Stdout, metadata)
	}
}
