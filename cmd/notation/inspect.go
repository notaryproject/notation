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
	cmderr "github.com/notaryproject/notation/cmd/notation/internal/errors"
	"github.com/notaryproject/notation/cmd/notation/internal/experimental"
	"github.com/notaryproject/notation/cmd/notation/internal/outputs"
	"github.com/notaryproject/notation/internal/cmd"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/spf13/cobra"
	"os"
)

type inspectOpts struct {
	cmd.LoggingFlagOpts
	SecureFlagOpts
	reference         string
	outputFormat      string
	allowReferrersAPI bool
	maxSignatures     int
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
	experimentalExamples := `
Example - [Experimental] Inspect signatures on an OCI artifact identified by a digest using the Referrers API, if not supported (returns 404), fallback to the Referrers tag schema
  notation inspect --allow-referrers-api <registry>/<repository>@<digest>
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
			return experimental.CheckFlagsAndWarn(cmd, "allow-referrers-api")
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
	cmd.SetPflagOutput(command.Flags(), &opts.outputFormat, cmd.PflagOutputUsage)
	command.Flags().IntVar(&opts.maxSignatures, "max-signatures", 100, "maximum number of signatures to evaluate or examine")
	cmd.SetPflagReferrersAPI(command.Flags(), &opts.allowReferrersAPI, fmt.Sprintf(cmd.PflagReferrersUsageFormat, "inspect"))
	experimental.HideFlags(command, experimentalExamples, []string{"allow-referrers-api"})
	return command
}

func runInspect(command *cobra.Command, opts *inspectOpts) error {
	// set log level
	ctx := opts.LoggingFlagOpts.InitializeLogger(command.Context())

	if opts.outputFormat != cmd.OutputJSON && opts.outputFormat != cmd.OutputPlaintext {
		return fmt.Errorf("unrecognized output format %s", opts.outputFormat)
	}

	// initialize
	reference := opts.reference
	sigRepo, err := getRemoteRepository(ctx, &opts.SecureFlagOpts, reference, opts.allowReferrersAPI)
	if err != nil {
		return err
	}
	manifestDesc, resolvedRef, err := resolveReferenceWithWarning(ctx, inputTypeRegistry, reference, sigRepo, "inspect")
	if err != nil {
		return err
	}
	output := outputs.InspectOutput{MediaType: manifestDesc.MediaType, Signatures: []outputs.SignatureOutput{}}
	skippedSignatures := false
	err = listSignatures(ctx, sigRepo, manifestDesc, opts.maxSignatures, func(sigManifestDesc ocispec.Descriptor) error {
		sigBlob, sigDesc, err := sigRepo.FetchSignatureBlob(ctx, sigManifestDesc)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: unable to fetch signature %s due to error: %v\n", sigManifestDesc.Digest.String(), err)
			skippedSignatures = true
			return nil
		}
		digest := sigManifestDesc.Digest.String()
		mediaType := sigDesc.MediaType
		err, output.Signatures = outputs.Signatures(mediaType, digest, output, [][]byte{sigBlob})
		if err != nil {
			return nil
		}

		return nil
	})
	var errorExceedMaxSignatures cmderr.ErrorExceedMaxSignatures
	if err != nil && !errors.As(err, &errorExceedMaxSignatures) {
		return err
	}

	if err := outputs.PrintOutput(opts.outputFormat, resolvedRef, output); err != nil {
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
