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

// Package policy provides the import and show commands for blob trust policy.
package policy

import (
	"github.com/spf13/cobra"
)

// Cmd returns the commands for policy including import and show.
func Cmd() *cobra.Command {
	command := &cobra.Command{
		Use:   "policy [command]",
		Short: "Manage trust policy configuration for signed blobs",
		Long:  "Manage trust policy configuration for arbitrary blob signature verification.",
	}

	command.AddCommand(
		importCmd(),
		showCmd(),
		initCmd(),
	)

	return command
}
