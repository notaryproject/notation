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
	"io"
	"io/fs"
	"os"

	"github.com/notaryproject/notation-go/dir"
	"github.com/notaryproject/notation-go/verifier/trustpolicy"
	"github.com/spf13/cobra"
)

func showCmd() *cobra.Command {
	command := &cobra.Command{
		Use:   "show [flags]",
		Short: "Show trust policy configuration",
		Long: `Show blob trust policy configuration.

Example - Show current blob trust policy configuration:
  notation blob policy show

Example - Save current blob trust policy configuration to a file:
  notation blob policy show > my_policy.json
`,
		Args: cobra.ExactArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runShow()
		},
	}
	return command
}

func runShow() error {
	policyJSON, err := loadBlobTrustPolicy()
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return fmt.Errorf("failed to show blob trust policy as the trust policy file does not exist.\nYou can import one using `notation blob policy import <path-to-policy.json>`")
		}
		return fmt.Errorf("failed to show trust policy: %w", err)
	}
	var doc trustpolicy.BlobDocument
	if err = json.Unmarshal(policyJSON, &doc); err == nil {
		err = doc.Validate()
	}
	if err != nil {
		fmt.Fprintf(os.Stderr, "Existing blob trust policy file is invalid, you may update or create a new one via `notation blob policy import <path-to-policy.json>`\n")
		os.Stdout.Write(policyJSON)
		return err
	}

	// show policy content
	_, err = os.Stdout.Write(policyJSON)
	return err
}

// loadBlobTrustPolicy loads the blob trust policy from notation configuration
// directory.
func loadBlobTrustPolicy() ([]byte, error) {
	f, err := dir.ConfigFS().Open(dir.PathBlobTrustPolicy)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return io.ReadAll(f)
}
