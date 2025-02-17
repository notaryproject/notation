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

package blob

import (
	"errors"
	"fmt"
	"os"

	"github.com/notaryproject/notation-core-go/signature"
	"github.com/notaryproject/notation/cmd/notation/internal/display"
	"github.com/notaryproject/notation/cmd/notation/internal/option"
	"github.com/spf13/cobra"
)

type inspectOpts struct {
	option.Common
	option.Format
	sigPath string
}

func inspectCommand() *cobra.Command {
	opts := &inspectOpts{}
	command := &cobra.Command{
		Use:   "inspect [flags] <signature_path>",
		Short: "Inspect a signature associated with a blob",
		Long: `Inspect a signature associated with a blob.

Example - Inspect a signature:
  notation blob inspect blob.cose.sig

Example - Inspect a signature and output as JSON:
  notation blob inspect -o json blob.cose.sig
`,
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return errors.New("missing signature path: use `notation blob inspect --help` to see what parameters are required")
			}
			opts.sigPath = args[0]
			return nil
		},
		PreRun: func(cmd *cobra.Command, args []string) {
			opts.Common.Parse(cmd)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return runInspect(opts)
		},
	}

	opts.Format.ApplyFlags(command.Flags(), option.FormatTypeText, option.FormatTypeJSON)
	return command
}

func runInspect(opts *inspectOpts) error {
	// initialize display handler
	displayHandler, err := display.NewBlobInspectHandler(opts.Printer, opts.Format)
	if err != nil {
		return err
	}

	// parse signature file
	signatureMediaType, err := parseSignatureMediaType(opts.sigPath)
	if err != nil {
		return err
	}
	envelopeBytes, err := os.ReadFile(opts.sigPath)
	if err != nil {
		return fmt.Errorf("failed to read signature file: %w", err)
	}
	envelope, err := signature.ParseEnvelope(signatureMediaType, envelopeBytes)
	if err != nil {
		return fmt.Errorf("failed to parse signature: %w", err)
	}
	if err := displayHandler.OnEnvelopeParsed(opts.sigPath, signatureMediaType, envelope); err != nil {
		return err
	}

	return displayHandler.Render()
}
