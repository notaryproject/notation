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
	"github.com/notaryproject/notation/cmd/notation/internal/policy"
	"github.com/spf13/cobra"
)

func showCmd() *cobra.Command {
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
			return policy.Show(true)
		},
	}
	return command
}
