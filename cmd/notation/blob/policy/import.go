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
	"fmt"
	"os"

	"github.com/notaryproject/notation-go/dir"
	"github.com/notaryproject/notation-go/verifier/trustpolicy"
	"github.com/notaryproject/notation/cmd/notation/internal/cmdutil"
	"github.com/notaryproject/notation/internal/osutil"
	"github.com/spf13/cobra"
)

type importOpts struct {
	filePath string
	force    bool
}

func importCmd() *cobra.Command {
	var opts importOpts
	command := &cobra.Command{
		Use:   "import [flags] <file_path>",
		Short: "Import blob trust policy configuration from a JSON file",
		Long: `Import blob trust policy configuration from a JSON file.

Example - Import blob trust policy configuration from a JSON file and store as "trustpolicy.blob.json":
  notation blob policy import my_policy.json

Example - Import blob trust policy and override existing configuration without prompt:
  notation blob policy import --force my_policy.json
`,
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) != 1 {
				return fmt.Errorf("requires 1 argument but received %d.\nUsage: notation blob policy import <path-to-policy.json>\nPlease specify a trust policy configuration location as the argument", len(args))
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.filePath = args[0]
			return runImport(opts)
		},
	}
	command.Flags().BoolVar(&opts.force, "force", false, "override the existing blob trust policy configuration without prompt")
	return command
}

func runImport(opts importOpts) error {
	// read configuration
	policyJSON, err := os.ReadFile(opts.filePath)
	if err != nil {
		return fmt.Errorf("failed to read blob trust policy file: %w", err)
	}

	var doc trustpolicy.BlobDocument
	if err = json.Unmarshal(policyJSON, &doc); err != nil {
		return fmt.Errorf("failed to parse blob trust policy configuration: %w", err)
	}
	if err = doc.Validate(); err != nil {
		return fmt.Errorf("failed to validate blob trust policy configuration: %w", err)
	}

	// optional confirmation
	if !opts.force {
		if _, err = trustpolicy.LoadBlobDocument(); err == nil {
			confirmed, err := cmdutil.AskForConfirmation(os.Stdin, "The blob trust policy configuration already exists, do you want to overwrite it?", opts.force)
			if err != nil {
				return err
			}
			if !confirmed {
				return nil
			}
		}
	} else {
		fmt.Fprintln(os.Stderr, "Warning: existing blob trust policy configuration will be overwritten")
	}

	// write
	policyPath, err := dir.ConfigFS().SysPath(dir.PathBlobTrustPolicy)
	if err != nil {
		return fmt.Errorf("failed to obtain path of blob trust policy configuration: %w", err)
	}
	if err = osutil.WriteFile(policyPath, policyJSON); err != nil {
		return fmt.Errorf("failed to write blob trust policy configuration: %w", err)
	}

	_, err = fmt.Fprintf(os.Stdout, "Successfully imported blob trust policy configuration to %s.\n", policyPath)
	return err
}
