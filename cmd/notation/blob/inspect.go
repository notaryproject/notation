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

	"github.com/notaryproject/notation/internal/cmd"
	"github.com/notaryproject/notation/internal/envelope"
	"github.com/notaryproject/notation/internal/ioutil"
	"github.com/spf13/cobra"
)

type inspectOpts struct {
	sigPath      string
	outputFormat string
}

func inspectCommand() *cobra.Command {
	opts := &inspectOpts{}
	command := &cobra.Command{
		Use:   "inspect [flags] <signature_path>",
		Short: "Inspect a signature associated with a blob",
		Long: `Inspect a signature associated with a blob.

Example - Inspect a signature:
  notation inspect blob.cose.sig
`,
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return errors.New("missing signature path: use `notation blob inspect --help` to see what parameters are required")
			}
			opts.sigPath = args[0]
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return runInspect(opts)
		},
	}

	cmd.SetPflagOutput(command.Flags(), &opts.outputFormat, cmd.PflagOutputUsage)
	return command
}

func runInspect(opts *inspectOpts) error {
	if opts.outputFormat != cmd.OutputJSON && opts.outputFormat != cmd.OutputPlaintext {
		return fmt.Errorf("unrecognized output format %s", opts.outputFormat)
	}

	envelopeMediaType, err := parseEnvelopeMediaType(filepath.Base(opts.sigPath))
	if err != nil {
		return err
	}

	envelopeBytes, err := os.ReadFile(opts.sigPath)
	if err != nil {
		return err
	}

	sig, err := envelope.Parse(envelopeMediaType, envelopeBytes)
	if err != nil {
		return err
	}

	// displayed as UserDefinedAttributes
	sig.SignedArtifact.Annotations = nil

	switch opts.outputFormat {
	case cmd.OutputJSON:
		return ioutil.PrintObjectAsJSON(sig)
	case cmd.OutputPlaintext:
		sig.ToNode(opts.sigPath).Print()
	}
	return nil
}

func parseEnvelopeMediaType(filename string) (string, error) {
	parts := strings.Split(filename, ".")
	if len(parts) < 3 {
		return "", fmt.Errorf("invalid signature filename: %s", filename)
	}
	return envelope.GetEnvelopeMediaType(strings.ToLower(parts[len(parts)-2]))
}
