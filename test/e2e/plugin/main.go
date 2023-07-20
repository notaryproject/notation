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
	"errors"
	"os"

	"github.com/notaryproject/notation-go/plugin/proto"
	"github.com/spf13/cobra"
)

const NOTATION_USERNAME = "NOTATION_USERNAME"
const NOTATION_PASSWORD = "NOTATION_PASSWORD"

func main() {
	cmd := &cobra.Command{
		Use:           "plugin for Notation E2E test",
		SilenceUsage:  true,
		SilenceErrors: true,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			// check registry credentials are eliminated
			if os.Getenv(NOTATION_USERNAME) != "" || os.Getenv(NOTATION_PASSWORD) != "" {
				return &proto.RequestError{
					Code: proto.ErrorCodeValidation,
					Err:  errors.New("registry credentials are not eliminated"),
				}
			}
			return nil
		},
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
