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
	"fmt"

	"github.com/notaryproject/notation/cmd/notation/internal/policy"
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
		Short: "Import trust policy configuration from a JSON file",
		Long: `Import trust policy configuration from a JSON file.

** This command is in preview and under development. **

Example - Import trust policy configuration from a file:
  notation policy import my_policy.json
`,
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) != 1 {
				return fmt.Errorf("requires 1 argument but received %d.\nUsage: notation policy import <path-to-policy.json>\nPlease specify a trust policy file location as the argument", len(args))
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.filePath = args[0]
			return policy.Import(opts.filePath, opts.force, true)
		},
	}
	command.Flags().BoolVar(&opts.force, "force", false, "override the existing trust policy configuration, never prompt")
	return command
}
