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
	"encoding/json"
	"os"

	"github.com/spf13/cobra"
)

func main() {
	cmd := &cobra.Command{
		Use:           "plugin for Notation E2E test",
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	cmd.AddCommand(
		getPluginMetadataCommand(),
		describeKeyCommand(),
		generateSignatureCommand(),
		generateEnvelopeCommand(),
		verifySignatureCommand(),
	)

	if err := cmd.Execute(); err != nil {
		if newErr := json.NewEncoder(os.Stderr).Encode(err); newErr != nil {
			panic(newErr)
		}
		os.Exit(1)
	}
}
