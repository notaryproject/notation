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

package policy

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"

	"github.com/notaryproject/notation-go/dir"
	"github.com/notaryproject/notation-go/verifier/trustpolicy"
	"github.com/spf13/cobra"
)

type showOpts struct {
}

func showCmd() *cobra.Command {
	var opts showOpts
	command := &cobra.Command{
		Use:   "show [flags]",
		Short: "Show OCI trust policy file",
		Long: `Show OCI trust policy file.

Example - Show current OCI trust policy file:
  notation policy show

Example - Save current OCI trust policy file to a file:
  notation policy show > my_policy.json
`,
		Args: cobra.ExactArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runShow(cmd, opts)
		},
	}
	return command
}

func runShow(command *cobra.Command, opts showOpts) error {
	// core process
	policyJSON, err := loadOCITrustPolicy()
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return fmt.Errorf("failed to show OCI trust policy as the trust policy file does not exist.\nYou can import one using `notation policy import <path-to-policy.json>`")
		}
		return fmt.Errorf("failed to show OCI trust policy: %w", err)
	}
	var doc trustpolicy.OCIDocument
	if err = json.Unmarshal(policyJSON, &doc); err == nil {
		err = doc.Validate()
	}
	if err != nil {
		fmt.Fprintf(os.Stderr, "Existing OCI trust policy file is invalid, you may update or create a new one via `notation policy import <path-to-policy.json>`. See https://github.com/notaryproject/specifications/blob/8cf800c60b7315a43f0adbcae463d848a353b412/specs/trust-store-trust-policy.md#trust-policy-for-oci-artifacts for a trust policy example.\n")
		os.Stdout.Write(policyJSON)
		return err
	}

	// show policy content
	_, err = os.Stdout.Write(policyJSON)
	return err
}

func loadOCITrustPolicy() ([]byte, error) {
	if _, err := fs.Stat(dir.ConfigFS(), dir.PathTrustPolicy); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: old trust policy `trustpolicy.json` was be deprecated; please update the trust policy file by using `notation policy import`.\n")
	}
	data, err := fs.ReadFile(dir.ConfigFS(), dir.PathOCITrustPolicy)
	if err != nil && errors.Is(err, fs.ErrNotExist) {
		return fs.ReadFile(dir.ConfigFS(), dir.PathTrustPolicy)
	}
	return data, err
}
