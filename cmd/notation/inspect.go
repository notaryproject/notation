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

	"github.com/notaryproject/notation-go/registry"
	cmderr "github.com/notaryproject/notation/cmd/notation/internal/errors"
	"github.com/notaryproject/notation/cmd/notation/internal/experimental"
	"github.com/notaryproject/notation/internal/cmd"
	"github.com/notaryproject/notation/internal/envelope"
	"github.com/notaryproject/notation/internal/ioutil"
	"github.com/notaryproject/notation/internal/tree"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/spf13/cobra"
)

type inspectOpts struct {
	cmd.LoggingFlagOpts
	SecureFlagOpts
	reference         string
	outputFormat      string
	allowReferrersAPI bool
	maxSignatures     int
}

type inspectOutput struct {
	MediaType  string                `json:"mediaType"`
	Signatures []*envelope.Signature `json:"signatures"`
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
			return experimental.CheckFlagsAndWarn(cmd, "allow-referrers-api")
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			if opts.maxSignatures <= 0 {
				return fmt.Errorf("max-signatures value %d must be a positive number", opts.maxSignatures)
			}
			if cmd.Flags().Changed("allow-referrers-api") {
				fmt.Fprintln(os.Stderr, "Warning: flag '--allow-referrers-api' is deprecated and will be removed in future versions.")
			}
			return runInspect(cmd, opts)
		},
	}

	opts.LoggingFlagOpts.ApplyFlags(command.Flags())
	opts.SecureFlagOpts.ApplyFlags(command.Flags())
	cmd.SetPflagOutput(command.Flags(), &opts.outputFormat, cmd.PflagOutputUsage)
	command.Flags().IntVar(&opts.maxSignatures, "max-signatures", 100, "maximum number of signatures to evaluate or examine")
	cmd.SetPflagReferrersAPI(command.Flags(), &opts.allowReferrersAPI, fmt.Sprintf(cmd.PflagReferrersUsageFormat, "inspect"))
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
	output := inspectOutput{MediaType: manifestDesc.MediaType, Signatures: []*envelope.Signature{}}
	skippedSignatures := false
	err = listSignatures(ctx, sigRepo, manifestDesc, opts.maxSignatures, func(sigManifestDesc ocispec.Descriptor) error {
		sigBlob, sigDesc, err := sigRepo.FetchSignatureBlob(ctx, sigManifestDesc)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: unable to fetch signature %s due to error: %v\n", sigManifestDesc.Digest.String(), err)
			skippedSignatures = true
			return nil
		}

		sig, err := envelope.Parse(sigDesc.MediaType, sigBlob)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: Skipping signature %s because of error: %v\n", sigDesc.Digest, err)
			skippedSignatures = true
			return nil
		}

		// adding digest to the signature
		sig.Digest = sigManifestDesc.Digest.String()

		// clearing annotations from the SignedArtifact field since they're already
		// displayed as UserDefinedAttributes
		sig.SignedArtifact.Annotations = nil

		output.Signatures = append(output.Signatures, sig)

		return nil
	})
	var errorExceedMaxSignatures cmderr.ErrorExceedMaxSignatures
	if err != nil && !errors.As(err, &errorExceedMaxSignatures) {
		return err
	}

	if err := printOutput(opts.outputFormat, resolvedRef, output); err != nil {
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

func printOutput(outputFormat string, ref string, output inspectOutput) error {
	if outputFormat == cmd.OutputJSON {
		return ioutil.PrintObjectAsJSON(output)
	}

	if len(output.Signatures) == 0 {
		fmt.Printf("%s has no associated signature\n", ref)
		return nil
	}

	fmt.Println("Inspecting all signatures for signed artifact")
	root := tree.New(ref)
	cncfSigNode := root.Add(registry.ArtifactTypeNotation)

	for _, signature := range output.Signatures {
		cncfSigNode.Children = append(cncfSigNode.Children, signature.ToNode(signature.Digest))
	}

	root.Print()
	return nil
}
