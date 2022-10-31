package main

import (
	"errors"
	"os"
	"strings"

	notationregistry "github.com/notaryproject/notation-go/registry"
	"github.com/notaryproject/notation-go/verification"
	"github.com/notaryproject/notation/internal/cmd"
	"github.com/notaryproject/notation/internal/ioutil"

	"github.com/spf13/cobra"
	"oras.land/oras-go/v2/registry"
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
		Long: `Verify signatures associated with the artifact.

Prerequisite: a trusted certificate needs to be generated or added using the command "notation cert". 

notation verify [--plugin-config <key>=<value>...] [--username <username>] [--password <password>] <reference>`,
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
	verifier, err := getVerifier(opts, ref)
	if err != nil {
		return err
	}

	// set up verification plugin config.
	configs, err := cmd.ParseFlagPluginConfig(opts.pluginConfig)
	if err != nil {
		return err
	}

	// core verify process.
	ctx := verification.WithPluginConfig(command.Context(), configs)
	outcomes, err := verifier.Verify(ctx, ref.String())

	// write out.
	return ioutil.PrintVerificationResults(os.Stdout, outcomes, err, ref.Reference)
}

func getVerifier(opts *verifyOpts, ref registry.Reference) (*verification.Verifier, error) {
	authClient, plainHTTP, err := getAuthClient(&opts.SecureFlagOpts, ref)
	if err != nil {
		return nil, err
	}

	repo := notationregistry.NewRepositoryClient(authClient, ref, plainHTTP)

	return verification.NewVerifier(repo)
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
