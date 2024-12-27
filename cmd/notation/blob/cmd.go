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

// Package blob provides the command for blob trust policy configuration.
package blob

import (
	"github.com/notaryproject/notation/cmd/notation/blob/policy"
	"github.com/spf13/cobra"
)

// Cmd returns the command for blob
func Cmd() *cobra.Command {
	command := &cobra.Command{
		Use:   "blob [commnad]",
		Short: "Sign, verify and inspect signatures associated with blobs",
		Long:  "Sign, inspect, and verify signatures and configure trust policies.",
	}

	command.AddCommand(
		policy.Cmd(),
	)

	return command
}
