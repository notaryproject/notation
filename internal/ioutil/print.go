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

package ioutil

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"reflect"
	"text/tabwriter"

	"github.com/notaryproject/notation-go"
	"github.com/notaryproject/notation-go/config"
	"github.com/notaryproject/notation-go/verifier/trustpolicy"
	"github.com/notaryproject/notation-go/verifier/truststore"
)

func newTabWriter(w io.Writer) *tabwriter.Writer {
	return tabwriter.NewWriter(w, 0, 0, 3, ' ', 0)
}

// PrintKeyMap prints out key information given array of KeySuite
func PrintKeyMap(w io.Writer, target *string, v []config.KeySuite) error {
	tw := newTabWriter(w)
	fmt.Fprintln(tw, "NAME\tKEY PATH\tCERTIFICATE PATH\tID\tPLUGIN NAME\t")
	for _, key := range v {
		name := key.Name
		if target != nil && key.Name == *target {
			name = "* " + name
		}
		kp := key.X509KeyPair
		if kp == nil {
			kp = &config.X509KeyPair{}
		}
		ext := key.ExternalKey
		if ext == nil {
			ext = &config.ExternalKey{}
		}
		fmt.Fprintf(tw, "%s\t%s\t%s\t%s\t%s\t\n", name, kp.KeyPath, kp.CertificatePath, ext.ID, ext.PluginName)
	}
	return tw.Flush()
}

// PrintCertMap lists certificate files in the trust store given array of cert
// paths
func PrintCertMap(w io.Writer, certPaths []string) error {
	if len(certPaths) == 0 {
		return nil
	}
	tw := newTabWriter(w)
	fmt.Fprintln(tw, "STORE TYPE\tSTORE NAME\tCERTIFICATE\t")
	for _, cert := range certPaths {
		fileName := filepath.Base(cert)
		dir := filepath.Dir(cert)
		namedStore := filepath.Base(dir)
		dir = filepath.Dir(dir)
		storeType := filepath.Base(dir)
		fmt.Fprintf(tw, "%s\t%s\t%s\t\n", storeType, namedStore, fileName)
	}
	return tw.Flush()
}

// PrintObjectAsJSON takes an interface and prints it as an indented JSON string
func PrintObjectAsJSON(i interface{}) error {
	jsonBytes, err := json.MarshalIndent(i, "", "    ")
	if err != nil {
		return err
	}

	fmt.Println(string(jsonBytes))

	return nil
}

// PrintVerificationFailure prints out messages when verification fails
func PrintVerificationFailure(outcomes []*notation.VerificationOutcome, printOut string, err error, isBlob bool) error {
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
		if isBlob {
			return fmt.Errorf("provided signature verification failed against blob %s", printOut)
		}
		return fmt.Errorf("signature verification failed for all the signatures associated with %s", printOut)
	}
	return nil
}

// PrintVerificationSuccess prints out messages when verification succeeds
func PrintVerificationSuccess(outcomes []*notation.VerificationOutcome, printout string) {
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
		PrintMetadataIfPresent(outcome)
	}
}

// PrintMetadataIfPresent prints out user metadata if present
func PrintMetadataIfPresent(outcome *notation.VerificationOutcome) {
	// the signature envelope is parsed as part of verification.
	// since user metadata is only printed on successful verification,
	// this error can be ignored
	metadata, _ := outcome.UserMetadata()

	if len(metadata) > 0 {
		fmt.Println("\nThe artifact was signed with the following user metadata.")
		PrintMetadataMap(os.Stdout, metadata)
	}
}
