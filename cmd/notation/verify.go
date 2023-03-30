package main

import (
	"context"
	"errors"
	"fmt"
	"math"
	"os"
	"reflect"

	"github.com/notaryproject/notation-go"
	"github.com/notaryproject/notation-go/log"
	notationregistry "github.com/notaryproject/notation-go/registry"
	"github.com/notaryproject/notation-go/verifier"
	"github.com/notaryproject/notation-go/verifier/trustpolicy"
	"github.com/notaryproject/notation/internal/cmd"
	"github.com/notaryproject/notation/internal/ioutil"
	"github.com/opencontainers/go-digest"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"

	"github.com/spf13/cobra"
	"oras.land/oras-go/v2/registry"
)

const maxSignatureAttempts = math.MaxInt64

type verifyOpts struct {
	cmd.LoggingFlagOpts
	SecureFlagOpts
	reference        string
	pluginConfig     []string
	userMetadata     []string
	localContent     bool
	trustPolicyScope string
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

Example - Verify a signature on an OCI artifact referenced in an OCI layout using trust policy statement specified by scope.
  notation verify --local-content <registry>/<repository>@<digest> --scope <trust_policy_scope>

Example - Verify a signature on an OCI artifact identified by a tag and referenced in an OCI layout using trust policy statement specified by scope.
  notation verify --local-content <registry>/<repository>:<tag> --scope <trust_policy_scope>
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
	command.Flags().StringArrayVar(&opts.pluginConfig, "plugin-config", nil, "{key}={value} pairs that are passed as it is to a plugin, if the verification is associated with a verification plugin, refer plugin documentation to set appropriate values")
	cmd.SetPflagUserMetadata(command.Flags(), &opts.userMetadata, cmd.PflagUserMetadataVerifyUsage)
	command.Flags().BoolVar(&opts.localContent, "local-content", false, "verify local artifact")
	command.Flags().StringVar(&opts.trustPolicyScope, "scope", "", "trust policy scope for local artifact verification. This flag is required when local-content is set to true")
	command.MarkFlagsRequiredTogether("local-content", "scope")
	return command
}

func runVerify(command *cobra.Command, opts *verifyOpts) error {
	// set log level
	ctx := opts.LoggingFlagOpts.SetLoggerLevel(command.Context())

	// initialize
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

	// core verify process
	if opts.localContent {
		printOut, outcomes, err := verifyLocal(ctx, opts, verifier, configs, userMetadata)
		if err != nil {
			return err
		}
		onSucess(outcomes, printOut)
	} else {
		printOut, outcomes, err := verifyRemote(ctx, opts, verifier, configs, userMetadata)
		if err != nil {
			return err
		}
		onSucess(outcomes, printOut)
	}
	return nil
}

func verifyRemote(ctx context.Context, opts *verifyOpts, verifier notation.Verifier, configs, userMetadata map[string]string) (string, []*notation.VerificationOutcome, error) {
	reference := opts.reference
	sigRepo, err := getRemoteRepository(ctx, &opts.SecureFlagOpts, reference)
	if err != nil {
		return "", nil, err
	}
	// resolve the given reference and set the digest
	ref, err := resolveReference(ctx, &opts.SecureFlagOpts, reference, sigRepo, func(ref registry.Reference, manifestDesc ocispec.Descriptor) {
		fmt.Fprintf(os.Stderr, "Warning: Always verify the artifact using digest(@sha256:...) rather than a tag(:%s) because resolved digest may not point to the same signed artifact, as tags are mutable.\n", ref.Reference)
	})
	if err != nil {
		return "", nil, err
	}
	verifyOpts := notation.VerifyOptions{
		ArtifactReference: ref.String(),
		PluginConfig:      configs,
		// TODO: need to change MaxSignatureAttempts as a user input flag or
		// a field in config.json
		MaxSignatureAttempts: maxSignatureAttempts,
		UserMetadata:         userMetadata,
	}

	// core verify process
	_, outcomes, err := notation.Verify(ctx, verifier, sigRepo, verifyOpts)
	err = checkFailure(outcomes, ref.String(), err)
	if err != nil {
		return "", nil, err
	}
	return ref.String(), outcomes, nil
}

func verifyLocal(ctx context.Context, opts *verifyOpts, verifier notation.Verifier, configs, userMetadata map[string]string) (string, []*notation.VerificationOutcome, error) {
	layoutPath, layoutReference, err := parseOCILayoutReference(opts.reference)
	if err != nil {
		return "", nil, err
	}
	return verifyFromFolder(ctx, opts, verifier, layoutPath, layoutReference, configs, userMetadata)
}

func verifyFromFolder(ctx context.Context, opts *verifyOpts, verifier notation.Verifier, path, reference string, configs, userMetadata map[string]string) (string, []*notation.VerificationOutcome, error) {
	logger := log.GetLogger(ctx)

	sigRepo, err := notationregistry.NewOCIRepository(path, notationregistry.RepositoryOptions{})
	if err != nil {
		return "", nil, err
	}
	targetDesc, err := sigRepo.Resolve(ctx, reference)
	if err != nil {
		return "", nil, fmt.Errorf("failed to resolve OCI layout reference: %s", err)
	}
	logger.Infof("OCI layout reference %s resolved to target manifest descriptor: %+v", reference, targetDesc)
	if digest.Digest(reference).Validate() != nil {
		// layout.reference is a tag
		fmt.Fprintf(os.Stderr, "Warning: Always verify the artifact using digest(@sha256:...) rather than a tag(:%s) because resolved digest may not point to the same signed artifact, as tags are mutable.\n", reference)
	}
	reference = targetDesc.Digest.String()
	printOut := path + "@" + targetDesc.Digest.String()
	verifyOpts := notation.VerifyOptions{
		ArtifactReference:    opts.trustPolicyScope + "@" + reference,
		PluginConfig:         configs,
		MaxSignatureAttempts: maxSignatureAttempts,
		UserMetadata:         userMetadata,
	}

	// core process
	_, outcomes, err := notation.Verify(ctx, verifier, sigRepo, verifyOpts)
	err = checkFailure(outcomes, printOut, err)
	if err != nil {
		return "", nil, err
	}
	return printOut, outcomes, nil
}

func checkFailure(outcomes []*notation.VerificationOutcome, printOut string, err error) error {
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

func onSucess(outcomes []*notation.VerificationOutcome, printout string) {
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
