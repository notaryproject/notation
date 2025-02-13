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
	"errors"
	"fmt"
	"io"
	"io/fs"
	"path/filepath"
	"text/tabwriter"

	"github.com/notaryproject/notation-go"
	"github.com/notaryproject/notation-go/config"
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
	return nil
}
