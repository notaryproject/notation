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

	"github.com/notaryproject/notation-go"
	"github.com/notaryproject/notation/cmd/notation/internal/display"
	"github.com/notaryproject/notation/cmd/notation/internal/experimental"
	"github.com/notaryproject/notation/cmd/notation/internal/option"
	"github.com/notaryproject/notation/internal/cmd"
	"github.com/notaryproject/notation/internal/cmd/verifier"
	"github.com/notaryproject/notation/internal/ioutil"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/spf13/cobra"
)

type verifyOpts struct {
	cmd.LoggingFlagOpts
	SecureFlagOpts
	option.Common
	reference            string
	pluginConfig         []string
	userMetadata         []string
	allowReferrersAPI    bool
	ociLayout            bool
	trustPolicyScope     string
	inputType            inputType
	maxSignatureAttempts int
}

func verifyCommand(opts *verifyOpts) *cobra.Command {
	if opts == nil {
		opts = &verifyOpts{
			inputType: inputTypeRegistry, // remote registry by default
		}
	}
	longMessage := `Verify OCI artifacts

Prerequisite: added a certificate into trust store and created a trust policy.

Example - Verify a signature on an OCI artifact identified by a digest:
  notation verify <registry>/<repository>@<digest>

Example - Verify a signature on an OCI artifact identified by a tag  (Notation will resolve tag to digest):
  notation verify <registry>/<repository>:<tag>
`
	experimentalExamples := `
Example - [Experimental] Verify a signature on an OCI artifact referenced in an OCI layout using trust policy statement specified by scope.
  notation verify --oci-layout <registry>/<repository>@<digest> --scope <trust_policy_scope>

Example - [Experimental] Verify a signature on an OCI artifact identified by a tag and referenced in an OCI layout using trust policy statement specified by scope.
  notation verify --oci-layout <registry>/<repository>:<tag> --scope <trust_policy_scope>
`
	command := &cobra.Command{
		Use:   "verify [reference]",
		Short: "Verify OCI artifacts",
		Long:  longMessage,
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return errors.New("missing reference to the artifact: use `notation verify --help` to see what parameters are required")
			}
			opts.reference = args[0]
			return nil
		},
		PreRunE: func(cmd *cobra.Command, args []string) error {
			if opts.ociLayout {
				opts.inputType = inputTypeOCILayout
			}
			opts.Common.Parse(cmd)
			return experimental.CheckFlagsAndWarn(cmd, "allow-referrers-api", "oci-layout", "scope")
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			if opts.maxSignatureAttempts <= 0 {
				return fmt.Errorf("max-signatures value %d must be a positive number", opts.maxSignatureAttempts)
			}
			if cmd.Flags().Changed("allow-referrers-api") {
				fmt.Fprintln(os.Stderr, "Warning: flag '--allow-referrers-api' is deprecated and will be removed in future versions.")
			}
			return runVerify(cmd, opts)
		},
	}
	opts.LoggingFlagOpts.ApplyFlags(command.Flags())
	opts.SecureFlagOpts.ApplyFlags(command.Flags())
	command.Flags().StringArrayVar(&opts.pluginConfig, "plugin-config", nil, "{key}={value} pairs that are passed as it is to a plugin, if the verification is associated with a verification plugin, refer plugin documentation to set appropriate values")
	cmd.SetPflagUserMetadata(command.Flags(), &opts.userMetadata, cmd.PflagUserMetadataVerifyUsage)
	cmd.SetPflagReferrersAPI(command.Flags(), &opts.allowReferrersAPI, fmt.Sprintf(cmd.PflagReferrersUsageFormat, "verify"))
	command.Flags().IntVar(&opts.maxSignatureAttempts, "max-signatures", 100, "maximum number of signatures to evaluate or examine")
	command.Flags().BoolVar(&opts.ociLayout, "oci-layout", false, "[Experimental] verify the artifact stored as OCI image layout")
	command.Flags().StringVar(&opts.trustPolicyScope, "scope", "", "[Experimental] set trust policy scope for artifact verification, required and can only be used when flag \"--oci-layout\" is set")
	command.MarkFlagsRequiredTogether("oci-layout", "scope")
	experimental.HideFlags(command, experimentalExamples, []string{"oci-layout", "scope"})
	return command
}

func runVerify(command *cobra.Command, opts *verifyOpts) error {
	// set log level
	ctx := opts.LoggingFlagOpts.InitializeLogger(command.Context())

	// initialize
	displayHandler := display.NewVerifyHandler(opts.Printer)
	sigVerifier, err := verifier.GetVerifier(ctx)
	if err != nil {
		return err
	}

	// set up verification plugin config
	configs, err := cmd.ParseFlagMap(opts.pluginConfig, cmd.PflagPluginConfig.Name)
	if err != nil {
		return err
	}

	// set up user metadata
	userMetadata, err := cmd.ParseFlagMap(opts.userMetadata, cmd.PflagUserMetadata.Name)
	if err != nil {
		return err
	}

	// core verify process
	reference := opts.reference
	// always use the Referrers API, if not supported, automatically fallback to
	// the referrers tag schema
	sigRepo, err := getRepository(ctx, opts.inputType, reference, &opts.SecureFlagOpts, false)
	if err != nil {
		return err
	}
	_, resolvedRef, err := resolveReference(ctx, opts.inputType, reference, sigRepo, func(ref string, manifestDesc ocispec.Descriptor) {
		displayHandler.OnResolvingTagReference(ref)
	})
	if err != nil {
		return err
	}
	intendedRef := resolveArtifactDigestReference(resolvedRef, opts.trustPolicyScope)
	verifyOpts := notation.VerifyOptions{
		ArtifactReference:    intendedRef,
		PluginConfig:         configs,
		MaxSignatureAttempts: opts.maxSignatureAttempts,
		UserMetadata:         userMetadata,
	}
	_, outcomes, err := notation.Verify(ctx, sigVerifier, sigRepo, verifyOpts)
	err = ioutil.ComposeVerificationFailurePrintout(outcomes, resolvedRef, err)
	if err != nil {
		return err
	}
	displayHandler.OnVerifySucceeded(outcomes, resolvedRef)
	return displayHandler.Render()
}
