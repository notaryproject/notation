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
	"context"
	"errors"
	"fmt"

	notationregistry "github.com/notaryproject/notation-go/registry"
	cmderr "github.com/notaryproject/notation/cmd/notation/internal/errors"
	"github.com/notaryproject/notation/cmd/notation/internal/experimental"
	"github.com/notaryproject/notation/internal/cmd"
	"github.com/opencontainers/go-digest"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/spf13/cobra"
)

type listOpts struct {
	cmd.LoggingFlagOpts
	SecureFlagOpts
	reference         string
	allowReferrersAPI bool
	ociLayout         bool
	inputType         inputType
	maxSignatures     int
}

func listCommand(opts *listOpts) *cobra.Command {
	if opts == nil {
		opts = &listOpts{
			inputType: inputTypeRegistry, // remote registry by default
		}
	}
	longMessage := `List all the signatures associated with signed artifact

Example - List signatures of an OCI artifact:
  notation list <registry>/<repository>@<digest>

Example - List signatures of an OCI artifact identified by a tag (Notation will resolve tag to digest)
  notation list <registry>/<repository>:<tag>
`
	experimentalExamples := `
Example - [Experimental] List signatures of an OCI artifact using the Referrers API. If it's not supported (returns 404), fallback to the Referrers tag schema
  notation list --allow-referrers-api <registry>/<repository>@<digest>

Example - [Experimental] List signatures of an OCI artifact referenced in an OCI layout
  notation list --oci-layout "<oci_layout_path>@<digest>"

Example - [Experimental] List signatures of an OCI artifact identified by a tag and referenced in an OCI layout
  notation list --oci-layout "<oci_layout_path>:<tag>"
`
	command := &cobra.Command{
		Use:     "list [flags] <reference>",
		Aliases: []string{"ls"},
		Short:   "List signatures of the signed artifact",
		Long:    longMessage,
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return errors.New("missing reference to the artifact: use `notation list --help` to see what parameters are required")
			}
			opts.reference = args[0]
			return nil
		},
		PreRunE: func(cmd *cobra.Command, args []string) error {
			if opts.ociLayout {
				opts.inputType = inputTypeOCILayout
			}
			return experimental.CheckFlagsAndWarn(cmd, "allow-referrers-api", "oci-layout")
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			if opts.maxSignatures <= 0 {
				return fmt.Errorf("max-signatures value %d must be a positive number", opts.maxSignatures)
			}
			return runList(cmd.Context(), opts)
		},
	}
	opts.LoggingFlagOpts.ApplyFlags(command.Flags())
	opts.SecureFlagOpts.ApplyFlags(command.Flags())
	cmd.SetPflagReferrersAPI(command.Flags(), &opts.allowReferrersAPI, fmt.Sprintf(cmd.PflagReferrersUsageFormat, "list"))
	command.Flags().BoolVar(&opts.ociLayout, "oci-layout", false, "[Experimental] list signatures stored in OCI image layout")
	experimental.HideFlags(command, "", []string{"allow-referrers-api", "oci-layout"})
	command.Flags().IntVar(&opts.maxSignatures, "max-signatures", 100, "maximum number of signatures to evaluate or examine")
	experimental.HideFlags(command, experimentalExamples, []string{"allow-referrers-api", "oci-layout"})
	return command
}

func runList(ctx context.Context, opts *listOpts) error {
	// set log level
	ctx = opts.LoggingFlagOpts.InitializeLogger(ctx)

	// initialize
	reference := opts.reference
	sigRepo, err := getRepository(ctx, opts.inputType, reference, &opts.SecureFlagOpts, opts.allowReferrersAPI)
	if err != nil {
		return err
	}
	targetDesc, resolvedRef, err := resolveReferenceWithWarning(ctx, opts.inputType, reference, sigRepo, "list")
	if err != nil {
		return err
	}
	// print all signature manifest digests
	return printSignatureManifestDigests(ctx, targetDesc, sigRepo, resolvedRef, opts.maxSignatures)
}

// printSignatureManifestDigests returns the signature manifest digests of
// the subject manifest.
func printSignatureManifestDigests(ctx context.Context, targetDesc ocispec.Descriptor, sigRepo notationregistry.Repository, ref string, maxSigs int) error {
	titlePrinted := false
	printTitle := func() {
		if !titlePrinted {
			fmt.Println(ref)
			fmt.Printf("└── %s\n", notationregistry.ArtifactTypeNotation)
			titlePrinted = true
		}
	}

	var prevDigest digest.Digest
	err := listSignatures(ctx, sigRepo, targetDesc, maxSigs, func(sigManifestDesc ocispec.Descriptor) error {
		// print the previous signature digest
		if prevDigest != "" {
			printTitle()
			fmt.Printf("    ├── %s\n", prevDigest)
		}
		prevDigest = sigManifestDesc.Digest
		return nil
	})
	// print the last signature digest
	if prevDigest != "" {
		printTitle()
		fmt.Printf("    └── %s\n", prevDigest)
	}
	if err != nil {
		var errExceedMaxSignatures cmderr.ErrorExceedMaxSignatures
		if !errors.As(err, &errExceedMaxSignatures) {
			return err
		}
		fmt.Println("Warning:", errExceedMaxSignatures)
	}

	if !titlePrinted {
		fmt.Printf("%s has no associated signature\n", ref)
	}
	return nil
}

// listSignatures lists signatures associated with manifestDesc with number of
// signatures limited by maxSig
func listSignatures(ctx context.Context, sigRepo notationregistry.Repository, manifestDesc ocispec.Descriptor, maxSig int, fn func(sigManifest ocispec.Descriptor) error) error {
	numOfSignatureProcessed := 0
	return sigRepo.ListSignatures(ctx, manifestDesc, func(signatureManifests []ocispec.Descriptor) error {
		for _, sigManifestDesc := range signatureManifests {
			if numOfSignatureProcessed >= maxSig {
				return cmderr.ErrorExceedMaxSignatures{MaxSignatures: maxSig}
			}
			numOfSignatureProcessed++
			if err := fn(sigManifestDesc); err != nil {
				return err
			}
		}
		return nil
	})
}
