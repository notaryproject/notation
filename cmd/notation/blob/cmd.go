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

// Package blob provides blob sign, verify, inspect, and policy commands.
package blob

import (
	"github.com/notaryproject/notation/cmd/notation/blob/policy"
	"github.com/spf13/cobra"
)

// Cmd returns the commands for blob
func Cmd() *cobra.Command {
	command := &cobra.Command{
		Use:   "blob [command]",
		Short: "Sign, inspect, verify signatures, and configure trust policies for blob artifacts",
		Long:  "Sign, inspect, verify signatures, and configure trust policies for blob artifacts.",
	}

	command.AddCommand(
		policy.Cmd(),
	)

	return command
}
