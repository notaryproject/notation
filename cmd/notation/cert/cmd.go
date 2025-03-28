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

// Package cert provides implementation of the `notation certificate` command
package cert

import "github.com/spf13/cobra"

func Cmd() *cobra.Command {
	command := &cobra.Command{
		Use:     "certificate",
		Aliases: []string{"cert"},
		Short:   "Manage certificates in trust store",
		Long:    "Manage certificates in trust store for signature verification.",
	}

	command.AddCommand(
		certAddCommand(nil),
		certListCommand(nil),
		certShowCommand(nil),
		certDeleteCommand(nil),
		certGenerateTestCommand(nil),
		certCleanupTestCommand(nil),
	)

	return command
}
