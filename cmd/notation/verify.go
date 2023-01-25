package main

import (
	"context"
	"errors"
	"fmt"
	"math"
	"os"
	"reflect"

	"github.com/notaryproject/notation-go"
	notationregistry "github.com/notaryproject/notation-go/registry"
	"github.com/notaryproject/notation-go/verifier"
	"github.com/notaryproject/notation-go/verifier/trustpolicy"
	"github.com/notaryproject/notation/internal/cmd"
	"github.com/notaryproject/notation/internal/ioutil"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"

	"github.com/spf13/cobra"
	"oras.land/oras-go/v2/registry"
)

type verifyOpts struct {
	cmd.LoggingFlagOpts
	SecureFlagOpts
	reference    string
	pluginConfig []string
	userMetadata []string
	outputFormat string
}

type verifyOutput struct {
	Reference    string            `json:"reference"`
	UserMetadata map[string]string `json:"userMetadata"`
	Result       string            `json:"result"`
}

func verifyCommand(opts *verifyOpts) *cobra.Command {
	if opts == nil {
		opts = &verifyOpts{}
	}
	command := &cobra.Command{
		Use:   "verify [reference]",
		Short: "Verify OCI artifacts",
		Long: `Verify OCI artifacts

Prerequisite: added a certificate into trust store and created a trust policy.

Example - Verify a signature on an OCI artifact identified by a digest:
  notation verify <registry>/<repository>@<digest>

Example - Verify a signature on an OCI artifact identified by a tag  (Notation will resolve tag to digest):
  notation verify <registry>/<repository>:<tag>
`,
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return errors.New("missing reference")
			}
			opts.reference = args[0]
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return runVerify(cmd, opts)
		},
	}
	opts.LoggingFlagOpts.ApplyFlags(command.Flags())
	opts.SecureFlagOpts.ApplyFlags(command.Flags())
	command.Flags().StringArrayVarP(&opts.pluginConfig, "plugin-config", "c", nil, "{key}={value} pairs that are passed as it is to a plugin, if the verification is associated with a verification plugin, refer plugin documentation to set appropriate values")
	cmd.SetPflagUserMetadata(command.Flags(), &opts.userMetadata, cmd.PflagUserMetadataVerifyUsage)
	cmd.SetPflagOutput(command.Flags(), &opts.outputFormat, cmd.PflagOutputUsage)
	return command
}

func runVerify(command *cobra.Command, opts *verifyOpts) error {
	// set log level
	ctx := opts.LoggingFlagOpts.SetLoggerLevel(command.Context())

	// initialize
	reference := opts.reference
	sigRepo, err := getSignatureRepository(ctx, &opts.SecureFlagOpts, reference)
	if err != nil {
		return err
	}

	// resolve the given reference and set the digest
	ref, err := resolveReference(command.Context(), &opts.SecureFlagOpts, reference, sigRepo, func(ref registry.Reference, manifestDesc ocispec.Descriptor) {
		if opts.outputFormat == ioutil.OutputPlaintext {
			fmt.Printf("Resolved artifact tag `%s` to digest `%s` before verification.\n", ref.Reference, manifestDesc.Digest.String())
			fmt.Println("Warning: The resolved digest may not point to the same signed artifact, since tags are mutable.")
		}
	})
	if err != nil {
		return err
	}

	// initialize verifier
	verifier, err := verifier.NewFromConfig()
	if err != nil {
		return err
	}

	// set up verification plugin config.
	configs, err := cmd.ParseFlagMap(opts.pluginConfig, cmd.PflagPluginConfig.Name)
	if err != nil {
		return err
	}

	// set up user metadata
	userMetadata, err := cmd.ParseFlagMap(opts.userMetadata, cmd.PflagUserMetadata.Name)
	if err != nil {
		return err
	}

	if opts.outputFormat != ioutil.OutputJson && opts.outputFormat != ioutil.OutputPlaintext {
		return fmt.Errorf("unrecognized output format: %v", opts.outputFormat)
	}

	verifyOpts := notation.RemoteVerifyOptions{
		ArtifactReference: ref.String(),
		PluginConfig:      configs,
		// TODO: need to change MaxSignatureAttempts as a user input flag or
		// a field in config.json
		MaxSignatureAttempts: math.MaxInt64,
		UserMetadata:         userMetadata,
	}

	// core verify process
	_, outcomes, err := notation.Verify(ctx, verifier, sigRepo, verifyOpts)
	// write out on failure
	if err != nil || len(outcomes) == 0 {
		if err != nil {
			var errorVerificationFailed notation.ErrorVerificationFailed
			if !errors.As(err, &errorVerificationFailed) {
				return fmt.Errorf("signature verification failed: %w", err)
			}
		}
		return fmt.Errorf("signature verification failed for all the signatures associated with %s", ref.String())
	}

	// write out on success
	outcome := outcomes[0]
	// print out warning for any failed result with logged verification action
	for _, result := range outcome.VerificationResults {
		if result.Error != nil {
			// at this point, the verification action has to be logged and
			// it's failed
			if opts.outputFormat == ioutil.OutputPlaintext {
				fmt.Printf("Warning: %v was set to %q and failed with error: %v\n", result.Type, result.Action, result.Error)
			}
		}
	}

	return printResult(opts.outputFormat, ref.String(), outcome)
}

func resolveReference(ctx context.Context, opts *SecureFlagOpts, reference string, sigRepo notationregistry.Repository, fn func(registry.Reference, ocispec.Descriptor)) (registry.Reference, error) {
	manifestDesc, ref, err := getManifestDescriptor(ctx, opts, reference, sigRepo)
	if err != nil {
		return registry.Reference{}, err
	}

	// reference is a digest reference
	if err := ref.ValidateReferenceAsDigest(); err == nil {
		return ref, nil
	}

	// reference is a tag reference
	fn(ref, manifestDesc)
	// resolve tag to digest reference
	ref.Reference = manifestDesc.Digest.String()

	return ref, nil
}

func printResult(outputFormat, reference string, outcome *notation.VerificationOutcome) error {

	if reflect.DeepEqual(outcome.VerificationLevel, trustpolicy.LevelSkip) {
		if outputFormat == ioutil.OutputPlaintext {
			fmt.Println("Trust policy is configured to skip signature verification for", reference)
			return nil
		} else {
			output := verifyOutput{Reference: reference, Result: "Skipped", UserMetadata: map[string]string{}}
			return ioutil.PrintObjectAsJson(output)
		}
	}

	// the signature envelope is parsed as part of verification.
	// since user metadata is only printed on successful verification,
	// this error can be ignored
	metadata, _ := outcome.GetUserMetadata()

	if outputFormat == ioutil.OutputJson {
		output := verifyOutput{Reference: reference, Result: "Success", UserMetadata: metadata}
		return ioutil.PrintObjectAsJson(output)
	}

	fmt.Println("Successfully verified signature for", reference)
	fmt.Println("\nThe artifact was signed with the following user metadata.")
	ioutil.PrintMetadataMap(os.Stdout, metadata)
	return nil
}
