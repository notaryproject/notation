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

package main

import (
	"fmt"
	"runtime"

	"github.com/notaryproject/notation/internal/version"
	"github.com/spf13/cobra"
)

func versionCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "version",
		Short: "Show the notation version information",
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			runVersion()
		},
	}
	return cmd
}

func runVersion() {
	fmt.Printf("Notation - a tool to sign and verify artifacts.\n\n")

	fmt.Printf("Version:     %s\n", version.GetVersion())
	fmt.Printf("Go version:  %s\n", runtime.Version())

	if version.GitCommit != "" {
		fmt.Printf("Git commit:  %s\n", version.GitCommit)
	}
}
