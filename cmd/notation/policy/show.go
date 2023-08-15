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
	"github.com/spf13/cobra"
)

type showOpts struct {
}

func showCmd() *cobra.Command {
	var opts showOpts
	command := &cobra.Command{
		Use:   "show [flags]",
		Short: "Show trust policy configuration",
		Long: `Show trust policy configuration.

** This command is in preview and under development. **

Example - Show current trust policy configuration:
  notation policy show

Example - Save current trust policy configuration to a file:
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
	// get policy file path
	policyPath, err := dir.ConfigFS().SysPath(dir.PathTrustPolicy)
	if err != nil {
		return fmt.Errorf("failed to obtain path of trust policy configuration file: %w", err)
	}

	// core process
	policyJSON, err := os.ReadFile(policyPath)
	if err != nil {
		return fmt.Errorf("failed to load trust policy configuration, you may import one via `notation policy import <path-to-policy.json>`: %w", err)
	}
	var doc trustpolicy.Document
	if err = json.Unmarshal(policyJSON, &doc); err == nil {
		err = doc.Validate()
	}
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", err.Error())
		fmt.Fprintf(os.Stderr, "Existing trust policy configuration is invalid, you may update or create a new one via `notation policy import <path-to-policy.json>`\n")
		// not returning to show the invalid policy configuration
	}

	// show policy content
	_, err = os.Stdout.Write(policyJSON)
	return err
}
