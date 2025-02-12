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
	"path/filepath"
	"strings"

	"github.com/notaryproject/notation-core-go/signature"
	"github.com/notaryproject/notation/cmd/notation/internal/display"
	"github.com/notaryproject/notation/cmd/notation/internal/option"
	"github.com/notaryproject/notation/internal/envelope"
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
	displayHandler, err := display.NewBlobInspectHandler(opts.Printer, opts.Format)
	if err != nil {
		return err
	}
	envelopeMediaType, err := parseEnvelopeMediaType(filepath.Base(opts.sigPath))
	if err != nil {
		return err
	}

	envelopeBytes, err := os.ReadFile(opts.sigPath)
	if err != nil {
		return fmt.Errorf("failed to read signature file: %w", err)
	}
	sigEnvelope, err := signature.ParseEnvelope(envelopeMediaType, envelopeBytes)
	if err != nil {
		return fmt.Errorf("failed to parse signature: %w", err)
	}
	if err := displayHandler.OnEnvelopeParsed(opts.sigPath, envelopeMediaType, sigEnvelope); err != nil {
		return err
	}

	return displayHandler.Render()
}

// parseEnvelopeMediaType returns the envelope media type based on the filename.
func parseEnvelopeMediaType(filename string) (string, error) {
	parts := strings.Split(filename, ".")
	if len(parts) < 3 || parts[len(parts)-1] != "sig" {
		return "", fmt.Errorf("invalid signature filename: %s", filename)
	}
	return envelope.GetEnvelopeMediaType(strings.ToLower(parts[len(parts)-2]))
}
