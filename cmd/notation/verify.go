package main

import (
	"errors"
	"os"
	"strings"

	"github.com/notaryproject/notation-go"
	notationregistry "github.com/notaryproject/notation-go/registry"
	"github.com/notaryproject/notation-go/verifier"
	"github.com/notaryproject/notation/internal/cmd"
	"github.com/notaryproject/notation/internal/ioutil"

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
		Use:   "verify [flags] <reference>",
		Short: "Verify Artifacts",
		Long:  "Verify signatures associated with the artifact.",
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
	// resolve the given reference and set the digest.
	ref, err := resolveReference(command, opts)
	if err != nil {
		return err
	}

	// initialize verifier.
	verifier, _ := verifier.NewFromConfig()
	authClient, plainHTTP, _ := getAuthClient(&opts.SecureFlagOpts, ref)
	remote_repo := remote.Repository{
		Client:    authClient,
		Reference: ref,
		PlainHTTP: plainHTTP,
	}
	repo := notationregistry.NewRepository(&remote_repo)

	// set up verification plugin config.
	configs, err := cmd.ParseFlagPluginConfig(opts.pluginConfig)
	if err != nil {
		return err
	}

	// core verify process.
	verifyOpts := notation.VerifyOptions{
		ArtifactReference:    ref.String(),
		SignatureMediaType:   "application/cose",
		PluginConfig:         configs,
		MaxSignatureAttempts: 50,
	}
	_, outcomes, err := notation.Verify(command.Context(), verifier, repo, verifyOpts)

	// write out.
	return ioutil.PrintVerificationResults(os.Stdout, outcomes, err, ref.Reference)
}

func resolveReference(command *cobra.Command, opts *verifyOpts) (registry.Reference, error) {
	ref, err := registry.ParseReference(opts.reference)
	if err != nil {
		return registry.Reference{}, err
	}

	if isDigestReference(opts.reference) {
		return ref, nil
	}

	// Resolve tag reference to digest reference.
	manifestDesc, err := getManifestDescriptorFromReference(command.Context(), &opts.SecureFlagOpts, opts.reference)
	if err != nil {
		return registry.Reference{}, err
	}

	ref.Reference = manifestDesc.Digest.String()
	return ref, nil
}

func isDigestReference(reference string) bool {
	parts := strings.SplitN(reference, "/", 2)
	if len(parts) == 1 {
		return false
	}

	index := strings.Index(parts[1], "@")
	return index != -1
}
