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
	"errors"
	"fmt"
	"os"

	"github.com/notaryproject/notation-core-go/signature"
	"github.com/notaryproject/notation/cmd/notation/internal/display"
	"github.com/notaryproject/notation/cmd/notation/internal/display/output"
	cmderr "github.com/notaryproject/notation/cmd/notation/internal/errors"
	"github.com/notaryproject/notation/cmd/notation/internal/flag"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/spf13/cobra"
)

type inspectOpts struct {
	flag.LoggingFlagOpts
	flag.SecureFlagOpts
	outputFormat  flag.OutputFormatFlagOpts
	outputPrinter *output.Printer
	reference     string
	maxSignatures int
}

func inspectCommand(opts *inspectOpts) *cobra.Command {
	if opts == nil {
		opts = &inspectOpts{}
	}
	longMessage := `Inspect all signatures associated with the signed artifact.

Example - Inspect signatures on an OCI artifact identified by a digest:
  notation inspect <registry>/<repository>@<digest>

Example - Inspect signatures on an OCI artifact identified by a tag  (Notation will resolve tag to digest):
  notation inspect <registry>/<repository>:<tag>

Example - Inspect signatures on an OCI artifact identified by a digest and output as json:
  notation inspect --output json <registry>/<repository>@<digest>
`
	command := &cobra.Command{
		Use:   "inspect [reference]",
		Short: "Inspect all signatures associated with the signed artifact",
		Long:  longMessage,
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return errors.New("missing reference to the artifact: use `notation inspect --help` to see what parameters are required")
			}
			opts.reference = args[0]
			return nil
		},
		PreRunE: func(cmd *cobra.Command, args []string) error {
			if err := opts.outputFormat.Validate(cmd); err != nil {
				return err
			}
			opts.outputPrinter = output.NewPrinter(cmd.OutOrStdout(), cmd.OutOrStderr())
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			if opts.maxSignatures <= 0 {
				return fmt.Errorf("max-signatures value %d must be a positive number", opts.maxSignatures)
			}
			return runInspect(cmd, opts)
		},
	}
	opts.LoggingFlagOpts.ApplyFlags(command.Flags())
	opts.SecureFlagOpts.ApplyFlags(command.Flags())
	command.Flags().IntVar(&opts.maxSignatures, "max-signatures", 100, "maximum number of signatures to evaluate or examine")

	// set output format
	opts.outputFormat.ApplyFlags(command.Flags(), output.FormatTree, output.FormatJSON)
	return command
}

func runInspect(command *cobra.Command, opts *inspectOpts) error {
	// set log level
	ctx := opts.LoggingFlagOpts.InitializeLogger(command.Context())

	displayHandler, err := display.NewInspectHandler(opts.outputPrinter, output.Format(opts.outputFormat.CurrentFormat))
	if err != nil {
		return err
	}

	// initialize
	reference := opts.reference
	// always use the Referrers API, if not supported, automatically fallback to
	// the referrers tag schema
	sigRepo, err := getRemoteRepository(ctx, &opts.SecureFlagOpts, reference, false)
	if err != nil {
		return err
	}
	manifestDesc, resolvedRef, err := resolveReferenceWithWarning(ctx, inputTypeRegistry, reference, sigRepo, "inspect")
	if err != nil {
		return err
	}
	displayHandler.OnReferenceResolved(resolvedRef, manifestDesc.MediaType)

	skippedSignatures := false
	err = listSignatures(ctx, sigRepo, manifestDesc, opts.maxSignatures, func(sigManifestDesc ocispec.Descriptor) error {
		sigBlob, sigDesc, err := sigRepo.FetchSignatureBlob(ctx, sigManifestDesc)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: unable to fetch signature %s due to error: %v\n", sigManifestDesc.Digest.String(), err)
			skippedSignatures = true
			return nil
		}

		envelope, err := signature.ParseEnvelope(sigDesc.MediaType, sigBlob)
		if err != nil {
			logSkippedSignature(sigManifestDesc, err)
			skippedSignatures = true
			return nil
		}

		if err := displayHandler.InspectSignature(sigManifestDesc, sigDesc, envelope); err != nil {
			logSkippedSignature(sigManifestDesc, err)
			skippedSignatures = true
			return nil
		}
		return nil
	})
	var errorExceedMaxSignatures cmderr.ErrorExceedMaxSignatures
	if err != nil && !errors.As(err, &errorExceedMaxSignatures) {
		return err
	}

	if err := displayHandler.Render(); err != nil {
		return err
	}

	if errorExceedMaxSignatures.MaxSignatures > 0 {
		fmt.Println("Warning:", errorExceedMaxSignatures)
	}

	if skippedSignatures {
		return errors.New("at least one signature was skipped and not displayed")
	}

	return nil
}

func logSkippedSignature(sigDesc ocispec.Descriptor, err error) {
	fmt.Fprintf(os.Stderr, "Warning: Skipping signature %s because of error: %v\n", sigDesc.Digest.String(), err)
}
