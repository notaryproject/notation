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
	"github.com/notaryproject/notation/internal/envelope"
	"github.com/notaryproject/notation/internal/ioutil"
	"github.com/notaryproject/notation/internal/osutil"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"

	"github.com/spf13/cobra"
	"oras.land/oras-go/v2/registry"
)

const maxSignatureAttempts = math.MaxInt64

type verifyOpts struct {
	cmd.LoggingFlagOpts
	SecureFlagOpts
	reference          string
	pluginConfig       []string
	userMetadata       []string
	localContent       bool
	trustPolicyScope   string
	localSignaturePath []string
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
			if len(args) > 1 {
				opts.localSignaturePath = append(opts.localSignaturePath, args[1:]...)
			}
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
	command.Flags().BoolVar(&opts.localContent, "local-content", false, "if set, verify local content")
	command.Flags().StringVar(&opts.trustPolicyScope, "scope", "", "trust policy scope for local content verification. If ignored, the global scope is used")
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

	var outcomes []*notation.VerificationOutcome
	var sigRepo notationregistry.Repository
	var artifactPrintout string
	var verifyErr error
	if opts.localContent {
		if osutil.CheckFile(opts.reference) {
			// reference is descriptor.json
			if len(opts.localSignaturePath) == 0 {
				return errors.New("missing signature for descriptor file verification")
			}
			signatures, err := parseSignaturesFromPathArray(opts.localSignaturePath)
			if err != nil {
				return err
			}
			targetDesc, err := getManifestDescriptorFromFile(opts.reference)
			if err != nil {
				return err
			}
			verifyOpts := notation.VerifyOptions{
				TargetAtLocal:    true,
				PluginConfig:     configs,
				UserMetadata:     userMetadata,
				TrustPolicyScope: opts.trustPolicyScope,
			}
			for _, signature := range signatures {
				signatureMediaType, err := envelope.SpeculateSignatureEnvelopeFormat(signature)
				if err != nil {
					return err
				}
				verifyOpts.SignatureMediaType = signatureMediaType
				outcome, sigErr := verifier.Verify(ctx, targetDesc, signature, verifyOpts)
				if sigErr != nil {
					if outcome == nil {
						return fmt.Errorf("signature verification failed: %w", sigErr)
					}
					continue
				}
				outcomes = []*notation.VerificationOutcome{outcome}
				break
			}
			if len(outcomes) == 0 {
				verifyErr = notation.ErrorVerificationFailed{}
			}
			artifactPrintout = opts.reference + "@" + targetDesc.Digest.String()
		} else {
			// reference is oci layout folder
			var layout ociLayout
			layout.path, layout.reference, err = parseOCILayoutReference(opts.reference)
			if err != nil {
				return err
			}
			sigRepo, err = ociLayoutFolderAsRepository(layout.path)
			if err != nil {
				return err
			}
			localVerifyOpts := notation.LocalVerifyOptions{
				LayoutReference:      layout.reference,
				PluginConfig:         configs,
				MaxSignatureAttempts: maxSignatureAttempts,
				UserMetadata:         userMetadata,
				TrustPolicyScope:     opts.trustPolicyScope,
			}
			var targetDesc ocispec.Descriptor
			if len(opts.localSignaturePath) == 0 {
				targetDesc, outcomes, verifyErr = notation.VerifyLocalContent(ctx, verifier, sigRepo, localVerifyOpts)
			} else {
				verifyOpts := notation.VerifyOptions{
					TargetAtLocal:    true,
					PluginConfig:     configs,
					UserMetadata:     userMetadata,
					TrustPolicyScope: opts.trustPolicyScope,
				}
				signatures, err := parseSignaturesFromPathArray(opts.localSignaturePath)
				if err != nil {
					return err
				}
				targetDesc, err = getManifestDescriptorFromOCILayout(ctx, layout.reference, sigRepo)
				if err != nil {
					return err
				}
				for _, signature := range signatures {
					signatureMediaType, err := envelope.SpeculateSignatureEnvelopeFormat(signature)
					if err != nil {
						return err
					}
					verifyOpts.SignatureMediaType = signatureMediaType
					outcome, sigErr := verifier.Verify(ctx, targetDesc, signature, verifyOpts)
					if sigErr != nil {
						if outcome == nil {
							return fmt.Errorf("signature verification failed: %w", sigErr)
						}
						continue
					}
					outcomes = []*notation.VerificationOutcome{outcome}
					break
				}
				if len(outcomes) == 0 {
					verifyErr = notation.ErrorVerificationFailed{}
				}
			}
			artifactPrintout = layout.path + "@" + targetDesc.Digest.String()
		}
	} else {
		reference := opts.reference
		sigRepo, err := getSignatureRepository(ctx, &opts.SecureFlagOpts, reference)
		if err != nil {
			return err
		}
		// resolve the given reference and set the digest
		ref, err := resolveReference(command.Context(), &opts.SecureFlagOpts, reference, sigRepo, func(ref registry.Reference, manifestDesc ocispec.Descriptor) {
			fmt.Fprintf(os.Stderr, "Warning: Always verify the artifact using digest(@sha256:...) rather than a tag(:%s) because resolved digest may not point to the same signed artifact, as tags are mutable.\n", ref.Reference)
		})
		if err != nil {
			return err
		}
		artifactPrintout = ref.String()

		verifyOpts := notation.RemoteVerifyOptions{
			ArtifactReference: ref.String(),
			PluginConfig:      configs,
			// TODO: need to change MaxSignatureAttempts as a user input flag or
			// a field in config.json
			MaxSignatureAttempts: maxSignatureAttempts,
			UserMetadata:         userMetadata,
		}

		// core verify process
		_, outcomes, verifyErr = notation.Verify(ctx, verifier, sigRepo, verifyOpts)
	}

	// write out on failure
	if verifyErr != nil || len(outcomes) == 0 {
		if verifyErr != nil {
			var errorVerificationFailed notation.ErrorVerificationFailed
			if !errors.As(verifyErr, &errorVerificationFailed) {
				return fmt.Errorf("signature verification failed: %w", verifyErr)
			}
		}
		return fmt.Errorf("signature verification failed for all the signatures associated with %s", artifactPrintout)
	}

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
		fmt.Println("Trust policy is configured to skip signature verification for", artifactPrintout)
	} else {
		fmt.Println("Successfully verified signature for", artifactPrintout)
		printMetadataIfPresent(outcome)
	}
	return nil
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

func parseSignaturesFromPathArray(sigPath []string) ([][]byte, error) {
	var signatures [][]byte
	for _, path := range sigPath {
		signature, err := os.ReadFile(path)
		if err != nil {
			return nil, err
		}
		signatures = append(signatures, signature)
	}
	return signatures, nil
}
