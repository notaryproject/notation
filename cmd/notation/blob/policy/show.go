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

	"github.com/notaryproject/notation-go/dir"
	"github.com/notaryproject/notation-go/verifier/trustpolicy"
	"github.com/notaryproject/notation/cmd/notation/internal/option"
	"github.com/spf13/cobra"
)

type showOpts struct {
	option.Common
}

func showCmd() *cobra.Command {
	opts := showOpts{}
	command := &cobra.Command{
		Use:   "show [flags]",
		Short: "Show blob trust policy configuration",
		Long: `Show blob trust policy configuration.

Example - Show current blob trust policy configuration:
  notation blob policy show

Example - Save current blob trust policy configuration to a file:
  notation blob policy show > my_policy.json
`,
		Args: cobra.ExactArgs(0),
		PreRun: func(cmd *cobra.Command, args []string) {
			opts.Common.Parse(cmd)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return runShow(&opts)
		},
	}
	return command
}

func runShow(opts *showOpts) error {
	policyJSON, err := fs.ReadFile(dir.ConfigFS(), dir.PathBlobTrustPolicy)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return fmt.Errorf("failed to show blob trust policy as the configuration does not exist.\nYou can import one using `notation blob policy import <path-to-policy.json>`")
		}
		return fmt.Errorf("failed to show trust policy: %w", err)
	}
	var doc trustpolicy.BlobDocument
	if err = json.Unmarshal(policyJSON, &doc); err == nil {
		err = doc.Validate()
	}
	if err != nil {
		opts.Printer.PrintErrorf("Existing blob trust policy configuration is invalid, you may update or create a new one via `notation blob policy import <path-to-policy.json>`. See https://github.com/notaryproject/specifications/blob/8cf800c60b7315a43f0adbcae463d848a353b412/specs/trust-store-trust-policy.md#trust-policy-for-blobs for a blob trust policy example.\n")
		opts.Printer.Write(policyJSON)
		return err
	}

	// show policy content
	opts.Printer.Write(policyJSON)
	return err
}
