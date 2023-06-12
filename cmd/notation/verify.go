package main

import (
	"errors"
	"fmt"
	"os"
	"reflect"

	"github.com/notaryproject/notation-go"
	"github.com/notaryproject/notation-go/verifier"
	"github.com/notaryproject/notation-go/verifier/trustpolicy"
	"github.com/notaryproject/notation/cmd/notation/internal/experimental"
	"github.com/notaryproject/notation/internal/cmd"
	"github.com/notaryproject/notation/internal/ioutil"
	"github.com/spf13/cobra"
)

type verifyOpts struct {
	cmd.LoggingFlagOpts
	SecureFlagOpts
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
Example - [Experimental] Verify an OCI artifact using the Referrers API, if not supported (returns 404), fallback to the Referrers tag schema
  notation verify --allow-referrers-api <registry>/<repository>@<digest>

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
				return errors.New("missing reference")
			}
			opts.reference = args[0]
			return nil
		},
		PreRunE: func(cmd *cobra.Command, args []string) error {
			if opts.ociLayout {
				opts.inputType = inputTypeOCILayout
			}
			return experimental.CheckFlagsAndWarn(cmd, "allow-referrers-api", "oci-layout", "scope")
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			if opts.maxSignatureAttempts <= 0 {
				return fmt.Errorf("max-signatures value %d must be a positive number", opts.maxSignatureAttempts)
			}
			return runVerify(cmd, opts)
		},
	}
	opts.LoggingFlagOpts.ApplyFlags(command.Flags())
	opts.SecureFlagOpts.ApplyFlags(command.Flags())
	command.Flags().StringArrayVar(&opts.pluginConfig, "plugin-config", nil, "{key}={value} pairs that are passed as it is to a plugin, if the verification is associated with a verification plugin, refer plugin documentation to set appropriate values")
	cmd.SetPflagUserMetadata(command.Flags(), &opts.userMetadata, cmd.PflagUserMetadataVerifyUsage)
	command.Flags().IntVar(&opts.maxSignatureAttempts, "max-signatures", 100, "maximum number of signatures to evaluate or examine")
	cmd.SetPflagReferrersAPI(command.Flags(), &opts.allowReferrersAPI, fmt.Sprintf(cmd.PflagReferrersUsageFormat, "verify"))
	command.Flags().BoolVar(&opts.ociLayout, "oci-layout", false, "[Experimental] verify the artifact stored as OCI image layout")
	command.Flags().StringVar(&opts.trustPolicyScope, "scope", "", "[Experimental] set trust policy scope for artifact verification, required and can only be used when flag \"--oci-layout\" is set")
	command.MarkFlagsRequiredTogether("oci-layout", "scope")
	experimental.HideFlags(command, experimentalExamples, []string{"allow-referrers-api", "oci-layout", "scope"})
	return command
}

func runVerify(command *cobra.Command, opts *verifyOpts) error {
	// set log level
	ctx := opts.LoggingFlagOpts.SetLoggerLevel(command.Context())

	// initialize
	sigVerifier, err := verifier.NewFromConfig()
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

	// core verify process
	reference := opts.reference
	sigRepo, err := getRepository(ctx, opts.inputType, reference, &opts.SecureFlagOpts, opts.allowReferrersAPI)
	if err != nil {
		return err
	}
	// resolve the given reference and set the digest
	_, resolvedRef, err := resolveReferenceWithWarning(ctx, opts.inputType, reference, sigRepo, "verify")
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
	err = checkVerificationFailure(outcomes, resolvedRef, err)
	if err != nil {
		return err
	}
	reportVerificationSuccess(outcomes, resolvedRef)
	return nil
}

func checkVerificationFailure(outcomes []*notation.VerificationOutcome, printOut string, err error) error {
	// write out on failure
	if err != nil || len(outcomes) == 0 {
		if err != nil {
			var errorVerificationFailed notation.ErrorVerificationFailed
			if !errors.As(err, &errorVerificationFailed) {
				return fmt.Errorf("signature verification failed: %w", err)
			}
		}
		return fmt.Errorf("signature verification failed for all the signatures associated with %s", printOut)
	}
	return nil
}

func reportVerificationSuccess(outcomes []*notation.VerificationOutcome, printout string) {
	// write out on success
	outcome := outcomes[0]
	// print out warning for any failed result with logged verification action
	for _, result := range outcome.VerificationResults {
		if result.Error != nil {
			// at this point, the verification action has to be logged and
			// it's failed
			fmt.Fprintf(os.Stderr, "Warning: %v was set to %q and failed with error: %v\n", result.Type, result.Action, result.Error)
		}
	}
	if reflect.DeepEqual(outcome.VerificationLevel, trustpolicy.LevelSkip) {
		fmt.Println("Trust policy is configured to skip signature verification for", printout)
	} else {
		fmt.Println("Successfully verified signature for", printout)
		printMetadataIfPresent(outcome)
	}
}

func printMetadataIfPresent(outcome *notation.VerificationOutcome) {
	// the signature envelope is parsed as part of verification.
	// since user metadata is only printed on successful verification,
	// this error can be ignored
	metadata, _ := outcome.UserMetadata()

	if len(metadata) > 0 {
		fmt.Println("\nThe artifact was signed with the following user metadata.")
		ioutil.PrintMetadataMap(os.Stdout, metadata)
	}
}
