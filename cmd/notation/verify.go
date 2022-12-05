package main

import (
	"context"
	"errors"
	"fmt"
	"math"

	"github.com/notaryproject/notation-go"
	notationregistry "github.com/notaryproject/notation-go/registry"
	"github.com/notaryproject/notation-go/verifier"
	"github.com/notaryproject/notation/internal/cmd"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"

	"github.com/spf13/cobra"
	"oras.land/oras-go/v2/registry"
	"oras.land/oras-go/v2/registry/remote"
)

type verifyOpts struct {
	SecureFlagOpts
	reference    string
	pluginConfig []string
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
	opts.ApplyFlags(command.Flags())
	command.Flags().StringArrayVarP(&opts.pluginConfig, "plugin-config", "c", nil, "{key}={value} pairs that are passed as it is to a plugin, if the verification is associated with a verification plugin, refer plugin documentation to set appropriate values")
	return command
}

func runVerify(command *cobra.Command, opts *verifyOpts) error {
	// resolve the given reference and set the digest
	ref, err := resolveReference(command.Context(), &opts.SecureFlagOpts, opts.reference, func(ref registry.Reference, manifestDesc ocispec.Descriptor) {
		fmt.Printf("Resolved artifact tag `%s` to digest `%s` before verification.\n", ref.Reference, manifestDesc.Digest.String())
		fmt.Println("Warning: The resolved digest may not point to the same signed artifact, since tags are mutable.")
	})
	if err != nil {
		return err
	}

	// initialize verifier
	verifier, err := verifier.NewFromConfig()
	if err != nil {
		return err
	}
	authClient, plainHTTP, _ := getAuthClient(&opts.SecureFlagOpts, ref)
	remoteRepo := remote.Repository{
		Client:    authClient,
		Reference: ref,
		PlainHTTP: plainHTTP,
	}
	repo := notationregistry.NewRepository(&remoteRepo)

	// set up verification plugin config
	configs, err := cmd.ParseFlagPluginConfig(opts.pluginConfig)
	if err != nil {
		return err
	}

	verifyOpts := notation.RemoteVerifyOptions{
		ArtifactReference: ref.String(),
		PluginConfig:      configs,
		// TODO: need to change MaxSignatureAttempts as a user input flag or
		// a field in config.json
		MaxSignatureAttempts: math.MaxInt64,
	}

	// core verify process
	_, outcomes, err := notation.Verify(command.Context(), verifier, repo, verifyOpts)

	// write out
	// on failure
	if err != nil || len(outcomes) == 0 {
		return fmt.Errorf("signature verification failed for all the signatures associated with %s", ref.String())
	}

	// on success
	outcome := outcomes[0]
	// print out warning for any failed result with logged verification action
	for _, result := range outcome.VerificationResults {
		if result.Error != nil {
			// at this point, the verification action has to be logged and
			// it's failed
			fmt.Printf("Warning: %v was set to \"logged\" and failed with error: %v\n", result.Type, result.Error)
		}
	}
	fmt.Println("Successfully verified signature for", ref.String())
	return nil
}

func resolveReference(ctx context.Context, opts *SecureFlagOpts, reference string, fn func(registry.Reference, ocispec.Descriptor)) (registry.Reference, error) {
	manifestDesc, ref, err := getManifestDescriptor(ctx, opts, reference)
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
